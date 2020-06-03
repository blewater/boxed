package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/pkg/config"
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

func (req *WorkflowServerReq) setServerMessenger(
	gRPCSrvConnToRemote wrpc.TaskCommunicator_RunWorkflowServer) {
	req.messenger = wrpc.NewSrvMessenger(gRPCSrvConnToRemote)
	req.cfg.Add(ConfigServerMessengerKey, req.messenger)
}

// initWorkflow initializes a pre-existing workflow looking it up by its name
// key within the cached workflow copies then from the database and the two
// previous attempts are fruitless, initialize a new workflow.
func (req *WorkflowServerReq) initWorkflow(
	ctx context.Context, srv *WorkflowsServer) (*types.Tasks, types.TaskConfiguration) {
	var err error
	// 1st Attempt to find workflow
	foundWorkflow, cfg := srv.getMemCachedWorkflow(req.workflowNameKey)
	if foundWorkflow != nil {
		return foundWorkflow, cfg
	}

	// Create new config for either db-fetched workflow or new workflow instance.
	cfg = config.NewTasksBoostrapConfig()
	// 2nd Attempt to load workflow from the database
	foundWorkflow, err = readWorkflowFromDB(ctx, req.workflowNameKey)
	if err != nil {
		log.Println("error : ", err, "while initializing workflow from the database")
		return nil, nil
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
			srv.setMemCachedWorkflow(req.workflowNameKey, foundWorkflow, cfg)
		}
	}

	// Initializing a workflow without any prior task execution
	foundWorkflow, err = types.NewWorkflow(cfg, req.messenger, req.workflowNameKey, srv.TaskRunners)
	if err != nil {
		fmt.Println(err)
	}

	// Cache the created workflow
	srv.setMemCachedWorkflow(req.workflowNameKey, foundWorkflow, cfg)

	return foundWorkflow, cfg
}

// stepProcessReceivedMessages performs step 2: gets or creates a workflow for
// the requested key. Logs any remote task progress. Save any received information.
func (req *WorkflowServerReq) stepProcessReceivedMessages(
	ctx context.Context, remoteMsg *wrpc.RemoteMsg) (nextAction int) {
	// Cannot continue processing this request if workflowNameKey is empty
	if remoteMsg.WorkflowNameKey == "" {
		return haltRequestAction
	}

	req.workflowNameKey = remoteMsg.WorkflowNameKey
	req.workflow, req.cfg = req.initWorkflow(ctx, req.WorkflowServer)

	if remoteMsg.TaskInProgress != "" {
		nextAction := req.printRemoteTaskError(remoteMsg)
		if nextAction != continueProcessing {
			return nextAction
		}

		remoteTaskMsg := fmt.Sprintf("Remote tasks feedback: %s\n", remoteMsg.TaskInProgress)
		log.Printf("%s\n", remoteTaskMsg)

		return pauseServerProcessing
	}

	if len(remoteMsg.Data) > 0 {
		for i := 0; i < len(remoteMsg.Data); i += 2 {
			key := strings.ToLower(remoteMsg.Data[i])
			value := remoteMsg.Data[i+1]
			req.cfg.Add(key, value)
		}
	}

	if remoteMsg.Datum != "" {
		// generic key "datum" for single value communications
		req.cfg.Add("datum", remoteMsg.Datum)
	}

	return continueProcessing
}

// stepIsWorkflowCompletedRemotely performs step 3: Have the completed remote
// tasks concluded the workflow?
func (req *WorkflowServerReq) stepIsWorkflowCompletedRemotely(
	ctx context.Context,
	remoteMsg *wrpc.RemoteMsg) (exit bool) {
	if !remoteMsg.TasksCompleted {
		return false
	}

	defer req.safeSaveWorkflow()

	if err := req.workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
		log.Println("error : ", err, ", server workflow state is invalid")

		return true
	}

	if req.workflow.SetWorkflowCompletedChecked(ctx) {
		if err := req.messenger.SignalSrvWorkflowCompletion(req.workflow.GetLen()); err != nil {
			log.Println(err, ", completion messaging error")
		}

		return true
	}

	return false
}

// stepRunServerSideTasks performs step 4: Resume server workflow tasks.
func (req *WorkflowServerReq) stepRunServerSideTasks(ctx context.Context, remoteMsg *wrpc.RemoteMsg) error {
	errIn := req.workflow.Run(ctx)

	if err := req.handleAnyServerOrRemoteErr(errIn, remoteMsg); err != nil {
		fmt.Println(err, ", workflow task running error")

		return err
	}

	return nil
}

// stepSendRemoteTasks performs step 5: Are pending remote tasks for execution?
func (req *WorkflowServerReq) stepSendRemoteTasks() (nextAction int) {
	if req.workflow.GetLen() == 0 {
		return continueProcessing
	}

	remoteTasksToExecute := req.workflow.GetPendingRemoteTaskNames()
	if len(remoteTasksToExecute) == 0 {
		return continueProcessing
	}

	err := req.messenger.SendRemoteTasksToRun(remoteTasksToExecute)
	if err != nil {
		fmt.Print(err, "sending remote task failed")

		return haltRequestAction
	}

	return pauseServerProcessing
}

// stepWorkflowCompleted performs step 6: Has the workflow completed?
func (req *WorkflowServerReq) stepWorkflowCompleted(ctx context.Context) (exit bool) {
	if req.workflow.SetWorkflowCompletedChecked(ctx) {
		if err := req.messenger.SignalSrvWorkflowCompletion(req.workflow.GetLen()); err != nil {
			fmt.Println(err, "sending workflow completion messaging error")

			return true
		}
	}

	return false
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

	clientErrMsg := remoteMsg.ErrorMsg
	if clientErrMsg != "" {
		clientErr := errors.New(clientErrMsg)
		fmt.Println(clientErr, "received client error")

		return clientErr
	}

	return nil
}

// printRemoteTaskError logs remote tasks errors which result to halting any
// further execution for this request.
func (req *WorkflowServerReq) printRemoteTaskError(remoteMsg *wrpc.RemoteMsg) int {
	if remoteMsg.ErrorMsg != "" {
		log.Println(errors.New(remoteMsg.ErrorMsg),
			"logger error while processing task:", remoteMsg.TaskInProgress)

		log.Printf("Remote task %s erred: %s\n", remoteMsg.TaskInProgress, remoteMsg.ErrorMsg)

		if err := req.workflow.CopyRemoteTasksProgress(remoteMsg); err != nil {
			log.Println("error : ", err, "while processing Remote messages")

			return haltRequestAction
		}

		defer req.safeSaveWorkflow()

		return haltRequestAction
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
