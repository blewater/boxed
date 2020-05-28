package grpc

import (
	"errors"
	"fmt"

	"github.com/tradeline-tech/workflow/pkg/config"
)

const (
	gRPCSrvKey                   = "grpcsrv"
	ServerWorkflowCompletionText = "workflow_completed"
)

type RPCSrvTaskConfiguration interface {
	GetRPCSrv() TaskCommunicator_RunWorkflowServer
	SetRPCSrv(TaskCommunicator_RunWorkflowServer)
}

func GetRPCSrv(cfg config.TaskConfiguration) TaskCommunicator_RunWorkflowServer {
	srv, ok := cfg.Get(gRPCSrvKey)
	if !ok {
		return nil
	}
	gRpcSrv := srv.(TaskCommunicator_RunWorkflowServer)
	return gRpcSrv
}

func SetRPCSrv(cfg config.TaskConfiguration, gRPCSrv TaskCommunicator_RunWorkflowServer) {
	cfg.Add(gRPCSrvKey, gRPCSrv)
}

//
// Collection of all server 2 remote gRPC message functions
//

// SendServerTaskProgressToRemote a server task text update on progress or error to the remote client
func SendServerTaskProgressToRemote(gRPCSrv TaskCommunicator_RunWorkflowServer,
	taskName, msgText string, errIn error) error {
	msg := &ServerMsg{}

	if errIn != nil {
		msg.ErrorMsg = errIn.Error()
		return gRPCSrv.Send(msg)
	}

	msg.TaskInProgress = taskName
	msg.TaskOutput = msgText

	return gRPCSrv.Send(msg)
}

// SendRemoteTasksToRun streams a message to the remote streaming connection
// of tasks to execute
func SendRemoteTasksToRun(gRPCSrv TaskCommunicator_RunWorkflowServer,
	taskNames []string) error {
	if len(taskNames) == 0 {
		return errors.New("no remote tasks to send")
	}

	errSend := gRPCSrv.Send(&ServerMsg{
		RemoteTasks: taskNames,
	})

	return fmt.Errorf("sending tasks to remote failed: %v", errSend)
}

// SendErrMsgToRemote streams error message to the remote streaming connection
// textErrorForRemote: is the text message we will send to the remote client
// preExistingError to include in the response
func SendErrMsgToRemote(gRpcSrv TaskCommunicator_RunWorkflowServer,
	textErrorForRemote string, preExistingError error) error {
	errSend := gRpcSrv.Send(&ServerMsg{
		ErrorMsg: textErrorForRemote,
	})

	if errSend != nil {
		if preExistingError == nil {
			return errSend
		}

		return fmt.Errorf("workflow: %v, errSend:%v", preExistingError, errSend)
	}

	if preExistingError != nil {
		return preExistingError
	}

	return nil
}

// SignalSrvWorkflowCompletion streams to cli that all work is done
func SignalSrvWorkflowCompletion(gRPCSrv TaskCommunicator_RunWorkflowServer, tasksLength int) error {
	fmt.Println("The workflow completed sending completion message to Cli...")

	errSend := gRPCSrv.Send(&ServerMsg{
		TaskInProgress: ServerWorkflowCompletionText,
		TaskOutput: fmt.Sprintf("The workflow completed execution for %d tasks.\n",
			tasksLength),
	})

	if errSend != nil {

		return errSend
	}

	return nil
}
