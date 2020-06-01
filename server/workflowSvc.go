package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"

	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

type WorkflowsType map[string]*types.Tasks

// SrvTaskRunners is the list of declared constructor tasks
// to run for the workflow tasks mapping one runner to one task
// they run in the sequence listed here.
type SrvTaskRunners []types.TaskRunnerNewFunc

// Server Workflow Loop Step Algorithm Actions
const (
	// Continue processing. Non-disrupting action effect.
	continueProcessing = 0
	// Stop processing this request.
	haltRequestAction = 1
	// Go back to listening for remote messages
	// typically this is desirable after sending
	// tasks for remote execution.
	skipLoopAction = 2
)

type WorkflowsServer struct {
	wrpc.UnimplementedTaskCommunicatorServer
	gRPCServer wrpc.TaskCommunicator_RunWorkflowServer

	srvTaskRunners SrvTaskRunners

	// Serialize access to concurrent workflows config access
	mu sync.Mutex

	// Cached org workflows, config
	workflows   WorkflowsType
	TaskRunners []types.TaskRunnerNewFunc
}

type WorkflowServerReq struct {
	WorkflowNameKey string
	Workflow        *types.Tasks
	WorkflowServer  *WorkflowsServer
}

var (
	cfg config.TaskConfiguration
)

func recoverFromPanic() {
	if r := recover(); r != nil {
		log.Println("recover: ", r)
	}
}

// NewWorkflowsServer initializes a workflow server for servicing
// a business group of with an instance of a workflow.
// Multiple workflows with the same declared tasks could run concurrently
// by one Workflow server.
func NewWorkflowsServer(srvTaskRunners SrvTaskRunners) *WorkflowsServer {
	return &WorkflowsServer{
		srvTaskRunners: srvTaskRunners,
		workflows:      make(WorkflowsType),
		// Link the declared task runners above with the workflow server
		TaskRunners: srvTaskRunners,
	}
}

func (srv *WorkflowsServer) GetgRPCServer() wrpc.TaskCommunicator_RunWorkflowServer {
	return srv.gRPCServer
}

// RunTasks opens a bidirectional stream to a remote client and
// runs tasks with workflow name-key received from the remote gRPC client.
func (srv *WorkflowsServer) RunWorkflow(stream wrpc.TaskCommunicator_RunWorkflowServer) error {
	var wReq = WorkflowServerReq{WorkflowServer: srv}
	defer recoverFromPanic()

	ctx := stream.Context()

	defer safeSaveWorkflow(wReq.Workflow)

	// Could re-enter when execution flow resumes from remote tasks
	for {
		clientMsg, err := stream.Recv()

		if stepCheckIOError(ctx, clientMsg, err) {
			break
		}

		if loopAction := wReq.stepProcessReceivedMessages(ctx, clientMsg); loopAction == haltRequestAction {
			break
		} else if loopAction == skipLoopAction {
			continue
		}

		if stepWorkflowCompletedRemotely(ctx, clientMsg, wReq.Workflow, stream) {
			break
		}

		if err = stepRunServerSideTasks(ctx, wReq.Workflow, stream, clientMsg); err != nil {
			break
		}

		if loopAction := stepSendRemoteTasks(wReq.Workflow, stream); loopAction == haltRequestAction {
			break
		} else if loopAction == skipLoopAction {
			continue
		}

		stepWorkflowCompleted(ctx, wReq.Workflow, stream)
	}

	return nil
}

// 1. Check whether the protocol errs
func stepCheckIOError(ctx context.Context, remoteMsg *wrpc.RemoteMsg, err error) (exit bool) {
	if err == io.EOF {
		log.Println("remote severed the connection to the server, EOF")

		return true
	}

	if err != nil {
		log.Println("error : ", err, ", server message reception error. Is the remote client still up?")

		return true
	}

	if err != nil {
		log.Println("error : ", err, ", server receiving messaging error")

		return true
	}

	if remoteMsg == nil {
		log.Println("error : ", err, ", "+
			"received nil remote message. Remote likely closed the connection.")

		return true
	}

	return false
}

// stepProcessReceivedMessages performs step 2: gets or creates a workflow for
// the requested key. Logs any remote task progress.
func (r *WorkflowServerReq) stepProcessReceivedMessages(
	ctx context.Context, remoteMsg *wrpc.RemoteMsg) (nextAction int) {
	// Cannot continue processing this request if workflowNameKey is empty
	if remoteMsg.WorkflowNameKey == "" {
		return haltRequestAction
	}

	r.WorkflowNameKey = remoteMsg.WorkflowNameKey
	r.Workflow = r.WorkflowServer.initWorkflow(ctx, r.WorkflowNameKey)

	if remoteMsg.TaskInProgress != "" {
		nextAction := printRemoteTaskError(remoteMsg, r.Workflow)
		if nextAction != continueProcessing {
			return nextAction
		}

		remoteTaskMsg := fmt.Sprintf("Remote tasks feedback: %s\n", remoteMsg.TaskInProgress)
		log.Printf("%s\n", remoteTaskMsg)

		return skipLoopAction
	}

	return continueProcessing
}

// printRemoteTaskError logs remote tasks errors which result to halting any
// further execution for this request.
func printRemoteTaskError(remoteMsg *wrpc.RemoteMsg, workflow *types.Tasks) int {
	if remoteMsg.ErrorMsg != "" {
		log.Println(errors.New(remoteMsg.ErrorMsg),
			"logger error while processing task:", remoteMsg.TaskInProgress)

		log.Printf("Remote task %s erred: %s\n", remoteMsg.TaskInProgress, remoteMsg.ErrorMsg)

		if err := workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
			log.Println("error : ", err, "while processing Remote messages")

			return haltRequestAction
		}

		defer safeSaveWorkflow(workflow)

		return haltRequestAction
	}

	return continueProcessing
}

// stepWorkflowCompletedRemotely performs step 3: Have the completed remote
// tasks concluded the workflow?
func stepWorkflowCompletedRemotely(
	ctx context.Context,
	remoteMsg *wrpc.RemoteMsg,
	workflow *types.Tasks,
	stream wrpc.TaskCommunicator_RunWorkflowServer) (exit bool) {
	if !remoteMsg.TasksCompleted {
		return false
	}

	defer safeSaveWorkflow(workflow)

	if err := workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
		// TODO .Error(ctx, err, "server workflow state is invalid")

		return true
	}

	if workflow.SetWorkflowCompletedChecked(ctx) {
		if err := wrpc.SignalSrvWorkflowCompletion(stream, workflow.GetLen()); err != nil {
			log.Println(err, ", completion messaging error")
		}

		return true
	}

	return false
}

// stepRunServerSideTasks performs step 4: Resume server workflow tasks.
func stepRunServerSideTasks(ctx context.Context,
	serverWorkflow *types.Tasks,
	stream wrpc.TaskCommunicator_RunWorkflowServer, remoteMsg *wrpc.RemoteMsg) error {
	errIn := serverWorkflow.Run(ctx)

	if err := handleAnyServerOrRemoteErr(errIn, stream, remoteMsg); err != nil {
		fmt.Println(err, ", workflow task running error")

		return err
	}

	return nil
}

// stepSendRemoteTasks performs step 5: Are pending remote tasks for execution?
func stepSendRemoteTasks(serverWorkflow *types.Tasks,
	gRPCSrv wrpc.TaskCommunicator_RunWorkflowServer) (nextAction int) {
	err := wrpc.SendRemoteTasksToRun(gRPCSrv, serverWorkflow.GetPendingRemoteTaskNames())
	if err != nil {
		fmt.Print(err, "sending remote task failed")

		return haltRequestAction
	}

	return skipLoopAction
}

// stepWorkflowCompleted performs step 6: Has the workflow completed?
func stepWorkflowCompleted(ctx context.Context,
	serverWorkflow *types.Tasks, gRPCSrv wrpc.TaskCommunicator_RunWorkflowServer) (exit bool) {
	if serverWorkflow.SetWorkflowCompletedChecked(ctx) {
		if err := wrpc.SignalSrvWorkflowCompletion(gRPCSrv, serverWorkflow.GetLen()); err != nil {
			fmt.Println(err, "sending workflow completion messaging error")

			return true
		}
	}

	return false
}

func safeSaveWorkflow(srvWorkflow *types.Tasks) {
	ctx := context.Background()

	if srvWorkflow != nil {
		// in the context of defer we absorb errors
		_ = datastore.Upsert(ctx, srvWorkflow)
	}
}

func (srv *WorkflowsServer) getMemCachedWorkflow(workflowNameKey string) *types.Tasks {
	if workflowNameKey == "" {
		return nil
	}

	srv.mu.Lock()
	// attempt to get cached workflow
	retWorkflow, _ := srv.workflows[workflowNameKey]
	srv.mu.Unlock()

	return retWorkflow
}

func (srv *WorkflowsServer) setMemCachedWorkflow(workflowNameKey string, workflow *types.Tasks) {
	if workflowNameKey == "" {
		return
	}

	srv.mu.Lock()
	srv.workflows[workflowNameKey] = workflow
	srv.mu.Unlock()
}

// initWorkflow initializes a pre-existing workflow looking it up by its name
// key within the cached workflow copies then from the database and the two
// previous attempts are fruitless, initialize a new workflow.
func (srv *WorkflowsServer) initWorkflow(ctx context.Context,
	workflowNameKey string) *types.Tasks {
	var err error
	// 1st Attempt to find workflow
	foundWorkflow := srv.getMemCachedWorkflow(workflowNameKey)
	if foundWorkflow != nil {
		return foundWorkflow
	}

	// 2nd Attempt to load workflow from the database
	foundWorkflow, err = readWorkflowFromDB(ctx, workflowNameKey)
	if err != nil {
		log.Println("error : ", err, "while initializing workflow from the database")
		return nil
	}

	if foundWorkflow != nil {
		log.Println("DB load error :", err)

		// check if there is a mismatch between declared and DB task runners (func objects)
		memoryDbCheckError := foundWorkflow.InitTasksMemState(cfg, srv.TaskRunners)
		if memoryDbCheckError != nil {
			// workflows don't match:
			// log it and then create a new one in the DB as we assume
			// that the correct workflow is the one declared in code.
			fmt.Println(
				"mongo database workflow mismatch, creating new workflow from taskRunners definitions",
				memoryDbCheckError)
		}

		// workflows db, declared task runners func match
		if memoryDbCheckError == nil {
			srv.setMemCachedWorkflow(workflowNameKey, foundWorkflow)
		}
	}

	// Initializing a workflow without any prior task execution
	foundWorkflow, err = types.NewWorkflow(cfg, workflowNameKey, srv.TaskRunners)
	if err != nil {
		fmt.Println(err)
	}

	// Cache the created workflow
	srv.setMemCachedWorkflow(workflowNameKey, foundWorkflow)

	return foundWorkflow
}

func readWorkflowFromDB(ctx context.Context, workflowNameKey string) (*types.Tasks, error) {
	readDocument, err := datastore.LoadByName(ctx, workflowNameKey)
	if err != nil {
		return nil, err
	}

	// No workflow found or db connection not used
	if readDocument == nil {
		return nil, nil
	}

	// type assert that the read document is a tasks type collection
	foundWorkflow, okWorkflow := readDocument.(*types.Tasks)
	if !okWorkflow {
		return nil, errors.New("failed to convert read document to a workflow")
	}
	return foundWorkflow, nil
}

// handleAnyServerOrRemoteErr combines handling server and remote errors. The
// last server task execution workflow error and any received client workflow errors.
func handleAnyServerOrRemoteErr(lastServerTaskError error,
	gRPCSrv wrpc.TaskCommunicator_RunWorkflowServer, remoteMsg *wrpc.RemoteMsg) error {
	if lastServerTaskError != nil {
		log.Println(lastServerTaskError, "Server workflow error: ", lastServerTaskError)

		errCombined := wrpc.SendErrMsgToRemote(gRPCSrv,
			"Server workflow task error", lastServerTaskError)
		if errCombined != nil {
			return errCombined
		}

		return lastServerTaskError
	}

	clientErrMsg := remoteMsg.ErrorMsg
	if clientErrMsg != "" {
		clientErr := errors.New(clientErrMsg)
		fmt.Println(clientErr, "received client error")

		return clientErr
	}

	return nil
}

func StartUp(
	srvAddress string,
	srvPort int, serverTaskRunners SrvTaskRunners) (gRpcServer *grpc.Server, port int, err error) {
	srvAddressPort := fmt.Sprintf("%s:%d", srvAddress, srvPort)
	tcpListener, err := net.Listen("tcp", srvAddressPort)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen at tcp %s", srvAddressPort)
	}
	log.Println("listening at ", srvAddressPort)
	gRpcServer = grpc.NewServer()

	workflowServer := NewWorkflowsServer(serverTaskRunners)

	wrpc.RegisterTaskCommunicatorServer(gRpcServer, workflowServer)
	log.Println("Workflow server starting...")

	if err = gRpcServer.Serve(tcpListener); err != nil {
		log.Println(err, "failed to start Workflow server")
	}

	return gRpcServer, port, nil
}

// GracefulShutdown handles the server shutdown process after receiving a signal
func GracefulShutdown(srv *grpc.Server) {
	log.Println("Termination signal received. Shutting down everything gracefully.")

	srv.GracefulStop()

	if srv != nil {
		srv.GracefulStop()
		log.Println("server stopped")
	}

	log.Println("Shutdown process completed.")
}

// SetupSigTermCloseHandler creates a 'listener' on a new goroutine which will
// notify the program if it receives an interrupt from the OS. We then handle
// this by calling our clean up procedure and exiting the program.
func SetupSigTermCloseHandler(srv *grpc.Server) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		GracefulShutdown(srv)
		os.Exit(0)
	}()
}
