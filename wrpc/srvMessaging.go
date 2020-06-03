package wrpc

import (
	"errors"
	"fmt"
	"log"
)

const (
	ServerWorkflowCompletionText = "workflow_completed"
)

type SrvMessenger struct {
	ConnToRemote TaskCommunicator_RunWorkflowServer
}

type MsgToRemote interface {
	SendServerTaskProgressToRemote(taskName, msgText string, errIn error) error
	SendRemoteTasksToRun(taskNames []string) error
	SendErrMsgToRemote(textErrorForRemote string, preExistingError error) error
	SignalSrvWorkflowCompletion(tasksLength int) error
	SendDatumToRemote(datum string) error
	SendDataToRemote(data []string) error
}

func NewSrvMessenger(connToSrv TaskCommunicator_RunWorkflowServer) MsgToRemote {
	return &SrvMessenger{ConnToRemote: connToSrv}
}

//
// Collection of all server 2 remote gRPC message functions
//

// SendServerTaskProgressToRemote a server task text update on progress or error to the remote client
func (msg *SrvMessenger) SendServerTaskProgressToRemote(taskName, msgText string, errIn error) error {
	serverMessage := &ServerMsg{}

	if errIn != nil {
		serverMessage.ErrorMsg = errIn.Error()
		return msg.ConnToRemote.Send(serverMessage)
	}

	serverMessage.TaskInProgress = taskName
	serverMessage.TaskOutput = msgText

	return msg.ConnToRemote.Send(serverMessage)
}

// SendRemoteTasksToRun streams a message to the remote streaming connection
// of tasks to execute
func (msg *SrvMessenger) SendRemoteTasksToRun(taskNames []string) error {
	if len(taskNames) == 0 {
		return errors.New("no remote tasks to send")
	}

	if errSend := msg.ConnToRemote.Send(&ServerMsg{RemoteTasks: taskNames}); errSend != nil {

		return fmt.Errorf("sending tasks to remote failed: %v", errSend)
	}

	return nil
}

// SendErrMsgToRemote streams error message to the remote streaming connection
// textErrorForRemote: is the text message we will send to the remote client
// preExistingError to include in the response
func (msg *SrvMessenger) SendErrMsgToRemote(textErrorForRemote string, preExistingError error) error {
	errSend := msg.ConnToRemote.Send(&ServerMsg{
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
func (msg *SrvMessenger) SignalSrvWorkflowCompletion(tasksLength int) error {
	log.Println("The workflow completed sending completion message to remote client...")

	errSend := msg.ConnToRemote.Send(&ServerMsg{
		TaskInProgress: ServerWorkflowCompletionText,
		TaskOutput: fmt.Sprintf("The workflow completed execution for %d tasks.\n",
			tasksLength),
	})

	if errSend != nil {

		return errSend
	}

	return nil
}

// SendDatumToRemote is the means for the server to send a single string data
// element to a remote.
func (msg *SrvMessenger) SendDatumToRemote(datum string) error {
	errSend := msg.ConnToRemote.Send(&ServerMsg{
		Datum: datum,
	})

	if errSend != nil {
		return errSend
	}

	return nil
}

// SendDataToRemote is the means for the server to send a slice of string data
// to a remote.
func (msg *SrvMessenger) SendDataToRemote(data []string) error {
	errSend := msg.ConnToRemote.Send(&ServerMsg{
		Data: data,
	})

	if errSend != nil {
		return errSend
	}

	return nil
}
