package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/tradeline-tech/argo/pkg/logger"

	"github.com/tradeline-tech/workflow/cfg"
	"github.com/tradeline-tech/workflow/common"
	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/grpc"
	"github.com/tradeline-tech/workflow/server/tasks"
)

type WorkflowsType map[string]*common.Tasks

// srvTaskRunners is the list of declared constructor tasks
// to run for the workflow tasks mapping one runner to one task
// they run in the sequence listed here
var srvTaskRunners = []common.TaskRunnerNewFunc{
	tasks.NewGetWorkflowName,
}

// Messaging loop actions
const (
	noAction    = 0
	breakAction = 1
	contAction  = 2
)

type WorkflowsServer struct {
	grpc.UnimplementedTaskCommunicatorServer
	gRPCServer grpc.TaskCommunicator_RunWorkflowServer

	// Serialize access to concurrent workflows config access
	mu sync.Mutex

	// Cached org workflows, config
	workflows   WorkflowsType
	TaskRunners []common.TaskRunnerNewFunc
}

var (
	config cfg.TaskConfiguration
)

func recoverFromPanic() {
	if r := recover(); r != nil {
		logger.Debug(context.Background(), "recover:", r)
	}
}

// NewWorkflowsServer initializes a workflow server for servicing
// a business group of with an instance of a workflow.
// Multiple workflows with the same declared tasks could run concurrently
// by one Workflow server.
func NewWorkflowsServer() *WorkflowsServer {
	return &WorkflowsServer{
		workflows: make(WorkflowsType),
		// Link the declared task runners above with the workflow server
		TaskRunners: srvTaskRunners,
	}
}

func (srv *WorkflowsServer) GetgRPCServer() grpc.TaskCommunicator_RunWorkflowServer {
	return srv.gRPCServer
}

// RunTasks opens a bidirectional stream to the client
// init fabric database config
// runs workflow tasks for the org whose token the Cli sends in the gRPC stream
func (srv *WorkflowsServer) RunTasks(stream grpc.TaskCommunicator_RunWorkflowServer) error {
	var workflowReq = &common.Tasks{}

	defer recoverFromPanic()

	ctx := stream.Context()

	defer safeSaveWorkflow(workflowReq)

	// Could re-enter when execution flow resumes from remote tasks
	for {
		clientMsg, err := stream.Recv()

		if stepCheckIOError(ctx, clientMsg, err) {
			break
		}

		if loopAction := stepProcessReceivedMessages(ctx, clientMsg, workflowReq); loopAction == breakAction {
			break
		} else if loopAction == contAction {
			continue
		}

		if stepWorkflowCompletedRemotely(ctx, clientMsg, workflowReq, stream) {
			break
		}

		if err = stepRunServerSideTasks(ctx, workflowReq, stream, clientMsg); err != nil {
			break
		}

		if loopAction := stepSendRemoteTasks(ctx, workflowReq, stream); loopAction == breakAction {
			break
		} else if loopAction == contAction {
			continue
		}

		stepWorkflowCompleted(ctx, workflowReq, stream)
	}

	return nil
}

// 1. Check whether the protocol errs
func stepCheckIOError(ctx context.Context, remoteMsg *grpc.RemoteMsg, err error) (exit bool) {
	if errClient := ctx.Err(); errClient != nil {

		return true
	}

	if err == io.EOF {
		logger.Info(ctx, "server EOF:", err)

		return true
	}

	if err != nil {
		logger.Error(ctx, err, "Message streaming error")

		return true
	}

	if remoteMsg == nil {
		logger.Info(ctx, "Received nil Client Message: Cli likely closed the connection. Closing the server connection...")

		return true
	}

	return false
}

// 2. Is a remote connection working on tasks?
func stepProcessReceivedMessages(ctx context.Context, remoteMsg *grpc.RemoteMsg,
	workflow *common.Tasks) (nextAction int) {
	if remoteMsg.TaskInProgress != "" {
		if remoteMsg.ErrorMsg != "" {
			logger.Error(ctx, errors.New(remoteMsg.ErrorMsg),
				"Cli error while processing task:", remoteMsg.TaskInProgress)

			fmt.Printf("Cli task %s erred: %s\n", remoteMsg.TaskInProgress, remoteMsg.ErrorMsg)

			if err := workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
				logger.Error(ctx, err, "while processing Cli messages")

				return breakAction
			}

			defer safeSaveWorkflow(workflow)

			return breakAction
		}

		cliTaskMsg := fmt.Sprintf("Cli tasks feedback: %s\n", remoteMsg.TaskInProgress)
		fmt.Printf("%s\n", cliTaskMsg)
		logger.Info(ctx, remoteMsg)

		return contAction
	}

	return noAction
}

// 3. Have the completed remote tasks concluded the workflow?
func stepWorkflowCompletedRemotely(
	ctx context.Context,
	remoteMsg *grpc.RemoteMsg,
	workflow *common.Tasks,
	stream grpc.TaskCommunicator_RunWorkflowServer) (exit bool) {
	if !remoteMsg.TasksCompleted {
		return false
	}

	defer safeSaveWorkflow(workflow)

	if err := workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
		logger.Error(ctx, err, "server workflow state is invalid")

		return true
	}

	if workflow.SetWorkflowCompletedChecked(ctx) {
		if err := common.SignalSrvWorkflowCompletion(stream, workflow.GetLen()); err != nil {
			logger.Error(ctx, err, "completion messaging error")
		}

		return true
	}

	return false
}

// 4. Resume server workflow tasks
func stepRunServerSideTasks(ctx context.Context,
	serverOrgWorkflow *common.Tasks,
	stream grpc.TaskCommunicator_RunWorkflowServer, remoteMsg *grpc.RemoteMsg) error {
	errIn := serverOrgWorkflow.Run(ctx)

	if err := handleMsgErr(ctx, errIn, stream, remoteMsg); err != nil {
		fmt.Println(err, "server workflow task running error")

		return err
	}

	return nil
}

// 5. Are pending remote tasks for execution?
func stepSendRemoteTasks(ctx context.Context,
	serverWorkflow *common.Tasks,
	gRPCSrv grpc.TaskCommunicator_RunWorkflowServer) (nextAction int) {
	err := common.SendRemoteTasksToDo(gRPCSrv, serverWorkflow.GetPendingRemoteTaskNames())
	if err != nil {
		fmt.Print(err, "sending remote task failed")

		return breakAction
	}

	return contAction
}

// 6. Has the workflow completed?
func stepWorkflowCompleted(ctx context.Context,
	serverWorkflow *common.Tasks, gRPCSrv grpc.TaskCommunicator_RunWorkflowServer) (exit bool) {
	if serverWorkflow.SetWorkflowCompletedChecked(ctx) {
		if err := common.SignalSrvWorkflowCompletion(gRPCSrv, serverWorkflow.GetLen()); err != nil {
			fmt.Println(err, "sending workflow completion messaging error")

			return true
		}
	}

	return false
}

func safeSaveWorkflow(srvWorkflow *common.Tasks) {
	ctx := context.Background()

	if srvWorkflow != nil {
		// in the context of defer we absorb errors
		_ = datastore.Upsert(ctx, srvWorkflow)
	}
}

func (srv *WorkflowsServer) getMemCachedWorkflow(workflowNameKey string) *common.Tasks {
	if workflowNameKey == "" {
		return nil
	}

	srv.mu.Lock()
	// attempt to get cached workflow
	retWorkflow, _ := srv.workflows[workflowNameKey]
	srv.mu.Unlock()

	return retWorkflow
}

func (srv *WorkflowsServer) setMemCachedWorkflow(workflowNameKey string, workflow *common.Tasks) {
	if workflowNameKey == "" {
		return
	}

	srv.mu.Lock()
	srv.workflows[workflowNameKey] = workflow
	srv.mu.Unlock()
}

// initWorkflow initializes a pre-existing workflow looking it up by its name key
// within the cached workflow copies
// -- then
// from the database
// and the two previous attempts are fruitless, initialize a new workflow
func (srv *WorkflowsServer) initWorkflow(ctx context.Context,
	workflowNameKey string) *common.Tasks {
	var err error
	// 1st Attempt to find workflow
	foundWorkflow := srv.getMemCachedWorkflow(workflowNameKey)
	if foundWorkflow != nil {
		return foundWorkflow
	}

	// 2nd Attempt to load workflow from the database
	foundWorkflow, err = readWorkflowFromDB(ctx, workflowNameKey)
	if err != nil {
		// TODO log err
		return nil
	}

	if foundWorkflow != nil {
		// TODO log err
		fmt.Println("DB load error :", err)

		// check if there is a mismatch between declared and DB task runners (func objects)
		memoryDbCheckError := foundWorkflow.InitTasksMemState(config, srv.TaskRunners)
		if memoryDbCheckError != nil {
			// workflows don't match:
			// log it and then create a new one in the DB as we assume that the correct workflow is the one
			// declared in code
			fmt.Println("mongo database workflow mismatch, creating new workflow from taskRunners definitions",
				memoryDbCheckError)
		}

		// workflows db, declared task runners func match
		if memoryDbCheckError == nil {
			srv.setMemCachedWorkflow(workflowNameKey, foundWorkflow)
		}
	}

	// Initializing a workflow without any prior task execution
	foundWorkflow, err = common.NewWorkflow(config, workflowNameKey, srv.TaskRunners)
	if err != nil {
		fmt.Println(err)
	}

	// Cache the created workflow
	srv.setMemCachedWorkflow(workflowNameKey, foundWorkflow)

	return foundWorkflow
}

func readWorkflowFromDB(ctx context.Context, workflowNameKey string) (*common.Tasks, error) {
	readDocument, err := datastore.LoadByName(ctx, workflowNameKey)
	if err != nil {
		// TODO log error
		return nil, err
	}

	// type assert that the read document is a tasks type collection
	foundWorkflow, okWorkflow := readDocument.(*common.Tasks)
	if !okWorkflow {
		return nil, errors.New("failed to convert read document to a workflow")
	}
	return foundWorkflow, nil
}

// handle multiple error cases:
// server workflow error
// combine with send error if any
// received client workflow error
func handleMsgErr(ctx context.Context, err error, gRPCSrv grpc.TaskCommunicator_RunWorkflowServer,
	remoteMsg *grpc.RemoteMsg) error {
	if err != nil {
		fmt.Println(err, "Server workflow error: ", err)

		errCombined := common.SendErrMsgToRemote(gRPCSrv, "Server workflow task error", err)
		if errCombined != nil {
			return errCombined
		}

		return err
	}

	clientErrMsg := remoteMsg.ErrorMsg
	if clientErrMsg != "" {
		clientErr := errors.New(clientErrMsg)
		fmt.Println(clientErr, "received client error")

		return clientErr
	}

	return nil
}
