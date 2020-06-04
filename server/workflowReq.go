package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

type WorkflowServerReq struct {
	workflowNameKey string
	workflow        *types.Tasks
	WorkflowServer  *WorkflowsServer
	cfg             types.TaskConfiguration
	messenger       types.MsgToRemote
}

func NewWorkflowReq(srv *WorkflowsServer, gRPCConnToRemote wrpc.TaskCommunicator_RunWorkflowServer) *WorkflowServerReq {
	return &WorkflowServerReq{
		WorkflowServer: srv,
		cfg:            config.NewTasksBoostrapConfig(),
		messenger:      wrpc.NewSrvMessenger(gRPCConnToRemote),
	}
}

func (req *WorkflowServerReq) process(
	ctx context.Context, messenger wrpc.TaskCommunicator_RunWorkflowServer, msg *wrpc.RemoteMsg) reqAction {
	var err error

	loopAction := req.stepProcessReceivedMessages(ctx, messenger, msg)
	if loopAction == exitServer || loopAction == waitNextRequest {
		return loopAction
	}

	if err = req.stepRunServerSideTasks(ctx, msg); err != nil {
		return exitServer
	}

	if loopAction := req.stepSendRemoteTasks(); loopAction == exitServer {
		return exitServer
	} else if loopAction == waitNextRequest {
		return waitNextRequest
	}

	return req.stepWorkflowCompleted(ctx)
}

// initWorkflow initializes a pre-existing workflow looking it up by its name
// key within the cached workflow copies then from the database and the two
// previous attempts are fruitless, initialize a new workflow.
func (req *WorkflowServerReq) initWorkflow(
	ctx context.Context, srv *WorkflowsServer) (*types.Tasks, types.TaskConfiguration) {
	var err error
	// 1st Attempt to find workflow
	foundWorkflow, foundCfg := srv.getMemCachedWorkflow(req.workflowNameKey)
	if foundWorkflow != nil {
		return foundWorkflow, foundCfg
	}

	// 2nd Attempt to load workflow from the database
	foundWorkflow, err = readWorkflowFromDB(ctx, req.workflowNameKey)
	if err != nil {
		log.Println("error : ", err, "while initializing workflow from the database")
		return nil, nil
	}

	if foundWorkflow != nil {
		log.Println("DB load error :", err)

		// check if there is a mismatch between declared and DB task runners (func objects)
		memoryDbCheckError := foundWorkflow.InitTasksMemState(req.cfg, srv.TaskRunners)
		if memoryDbCheckError != nil {
			// workflows don't match:
			// log it and then create a new one in the DB as we assume
			// that the correct workflow is the one declared in code.
			log.Println(
				"mongo database workflow mismatch, creating new workflow from taskRunners definitions",
				memoryDbCheckError)
		}

		// workflows db, declared task runners func match
		if memoryDbCheckError == nil {
			srv.setMemCachedWorkflow(req.workflowNameKey, foundWorkflow, req.cfg)
		}
	}

	// Initializing a workflow without any prior task execution
	foundWorkflow, err = types.NewWorkflow(req.cfg, req.messenger, req.workflowNameKey, srv.TaskRunners)
	if err != nil {
		log.Println(err)
	}

	// Cache the created workflow
	srv.setMemCachedWorkflow(req.workflowNameKey, foundWorkflow, req.cfg)

	return foundWorkflow, req.cfg
}

func (req *WorkflowServerReq) setWorkflowByName(ctx context.Context, workflowNameKey string) reqAction {
	// Unknown workflow request if workflowNameKey is empty
	if workflowNameKey == "" {
		return exitServer
	}

	req.workflowNameKey = workflowNameKey
	req.workflow, req.cfg = req.initWorkflow(ctx, req.WorkflowServer)

	req.cfg.Add(ConfigServerMessengerKey, req.messenger)

	return continueProcessing
}

// stepProcessReceivedMessages performs step 2: gets or creates a workflow for
// the requested key. Logs any remote task progress. Save any received information.
func (req *WorkflowServerReq) stepProcessReceivedMessages(
	ctx context.Context, messenger wrpc.TaskCommunicator_RunWorkflowServer, remoteMsg *wrpc.RemoteMsg) reqAction {
	req.setWorkflowByName(ctx, remoteMsg.WorkflowNameKey)

	if nextAction := req.setWorkflowByName(ctx, remoteMsg.WorkflowNameKey); nextAction != continueProcessing {
		return nextAction
	}

	if nextAction := req.logRemoteTaskProgress(remoteMsg); nextAction != continueProcessing {
		return nextAction
	}

	gotRemoteData := req.saveRemoteData(remoteMsg)

	tasksCompleted, err := req.saveRemoteTaskCompletion(ctx, remoteMsg)
	if err != nil {
		return exitServer
	}

	if gotRemoteData && !tasksCompleted {
		return waitNextRequest
	}

	// Process completion or process next tasks
	return continueProcessing
}

// saveRemoteData saves remote data to the tasks configuration for server
// processing.
func (req *WorkflowServerReq) saveRemoteData(remoteMsg *wrpc.RemoteMsg) bool {
	return req.saveRemoteDatumToCfg(remoteMsg) ||
		req.saveRemoteDataToCfg(remoteMsg)
}

func (req *WorkflowServerReq) saveRemoteDatumToCfg(remoteMsg *wrpc.RemoteMsg) bool {
	if remoteMsg.Datum != "" {
		// generic key "datum" for single value communications
		req.cfg.Add("datum", remoteMsg.Datum)

		return true
	}

	return false
}

func (req *WorkflowServerReq) saveRemoteDataToCfg(remoteMsg *wrpc.RemoteMsg) bool {
	if len(remoteMsg.Data) > 0 {
		for i := 0; i < len(remoteMsg.Data); i += 2 {
			key := strings.ToLower(remoteMsg.Data[i])
			value := remoteMsg.Data[i+1]
			req.cfg.Add(key, value)
		}

		return true
	}

	return false
}

// saveRemoteTaskCompletion checks for any remote tasks completion.
func (req *WorkflowServerReq) saveRemoteTaskCompletion(
	ctx context.Context, remoteMsg *wrpc.RemoteMsg) (tasksCompleted bool, err error) {
	tasksCompleted = remoteMsg.TasksCompleted
	if !tasksCompleted {
		return tasksCompleted, err
	}

	defer req.safeSaveWorkflow()

	if err = req.workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
		log.Println("error : ", err, ", server workflow state is invalid")

		return tasksCompleted, err
	}

	return tasksCompleted, err
}

// logRemoteTaskProgress prints any remote tasks execution errors in which case
// it halts execution and if not it logs any tasks progress.
func (req *WorkflowServerReq) logRemoteTaskProgress(remoteMsg *wrpc.RemoteMsg) reqAction {
	if nextAction := req.printRemoteTaskError(remoteMsg); nextAction != continueProcessing {

		return nextAction
	}

	if remoteMsg.TaskInProgress != "" {
		remoteTaskMsg := fmt.Sprintf("Remote tasks feedback: %s\n", remoteMsg.TaskInProgress)
		log.Printf("%s\n", remoteTaskMsg)

		return waitNextRequest
	}

	return continueProcessing
}

// stepRunServerSideTasks checks whether to resume processing server workflow tasks.
func (req *WorkflowServerReq) stepRunServerSideTasks(ctx context.Context, remoteMsg *wrpc.RemoteMsg) error {
	errIn := req.workflow.Run(ctx)

	if err := req.handleAnyServerOrRemoteErr(errIn, remoteMsg); err != nil {
		log.Println(err, ", workflow task running error")

		return err
	}

	return nil
}

// stepSendRemoteTasks checks for pending remote tasks to send for execution.
func (req *WorkflowServerReq) stepSendRemoteTasks() reqAction {
	if req.workflow.GetLen() == 0 {
		return continueProcessing
	}

	remoteTasksToExecute := req.workflow.GetPendingRemoteTaskNames()
	if len(remoteTasksToExecute) == 0 {
		return continueProcessing
	}

	err := req.messenger.SendRemoteTasksToRun(remoteTasksToExecute)
	if err != nil {
		log.Print(err, "sending remote task failed")

		return exitServer
	}

	return waitNextRequest
}

// stepWorkflowCompleted checks whether the workflow has completed.
func (req *WorkflowServerReq) stepWorkflowCompleted(ctx context.Context) reqAction {
	if req.workflow.SetWorkflowCompletedChecked(ctx) {
		if err := req.messenger.SignalSrvWorkflowCompletion(req.workflow.GetLen()); err != nil {
			log.Println(err, "sending workflow completion messaging error")

			return exitServer
		}

		if req.WorkflowServer.soloWorkflowMode {
			return exitServer
		}

		// Likely having same impact as continue processing
		// because it's the last step in the response loop.
		return waitNextRequest
	}

	return continueProcessing
}

// handleAnyServerOrRemoteErr combines handling server and remote errors. The
// last server task execution workflow error and any received client workflow errors.
func (req *WorkflowServerReq) handleAnyServerOrRemoteErr(lastServerTaskError error, remoteMsg *wrpc.RemoteMsg) error {
	if lastServerTaskError != nil {
		log.Println(lastServerTaskError, "Server workflow error: ", lastServerTaskError)

		errCombined := req.messenger.SendErrMsgToRemote(
			"Server workflow task error", lastServerTaskError)
		if errCombined != nil {
			return errCombined
		}

		return lastServerTaskError
	}
	// TODO is this needed?
	clientErrMsg := remoteMsg.ErrorMsg
	if clientErrMsg != "" {
		clientErr := errors.New(clientErrMsg)
		log.Println(clientErr, "received client error")

		return clientErr
	}

	return nil
}

// printRemoteTaskError logs remote tasks errors which result to halting any
// further execution for this request.
func (req *WorkflowServerReq) printRemoteTaskError(remoteMsg *wrpc.RemoteMsg) reqAction {
	if remoteMsg.ErrorMsg != "" {
		log.Println(errors.New(remoteMsg.ErrorMsg),
			"logger error while processing task:", remoteMsg.TaskInProgress)

		log.Printf("Remote task %s erred: %s\n", remoteMsg.TaskInProgress, remoteMsg.ErrorMsg)

		if err := req.workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
			log.Println("error : ", err, "while processing Remote messages")

			return exitServer
		}

		defer req.safeSaveWorkflow()

		return exitServer
	}

	return continueProcessing
}

func (req *WorkflowServerReq) safeSaveWorkflow() {
	ctx := context.Background()

	if req != nil {
		// in the context of defer we absorb errors
		_ = datastore.Upsert(ctx, req.workflow)
	}
}
