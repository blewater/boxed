package wrpc

import (
	"errors"
	"fmt"
)

//
// Collection of remote 2 Server gRPC message functions
//

const (
	errTextWorkflowKeyMissing     = "remote workflow name key required for messaging server apps"
	remoteTextWorkflowNameMissing = "missing remote workflow name key value to identify the workflow at server."
)

type RemoteMessenger struct {
	ConnToSrv TaskCommunicator_RunWorkflowClient
}

func NewRemoteMessenger(connToSrv TaskCommunicator_RunWorkflowClient) *RemoteMessenger {
	return &RemoteMessenger{ConnToSrv: connToSrv}
}

// SendWorkflowNameKeyToSrv sends the org identifier to the server
func (msg *RemoteMessenger) SendWorkflowNameKeyToSrv(workflowNameKey string) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	if err := msg.ConnToSrv.Send(&RemoteMsg{WorkflowNameKey: workflowNameKey}); err != nil {
		fmt.Println("error : ", err, ", failed to stream a remote gRPC message to server")

		return err
	}

	return nil
}

// SendTasksErrorMsgToServer sends error message to the Server for current task execution error
func (msg *RemoteMessenger) SendTasksErrorMsgToServer(
	workflowNameKey string,
	taskIndex int,
	currentTaskName string,
	remoteTasks []*RemoteMsg_Tasks,
	errTaskExec error) error {

	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	errSend := msg.ConnToSrv.Send(&RemoteMsg{
		ErrorMsg: fmt.Sprintf("error %v, remote tasks index: %d, task name: %s\n",
			errTaskExec, taskIndex, currentTaskName),
		WorkflowNameKey: workflowNameKey,
		Tasks:           remoteTasks,
		TaskInProgress:  currentTaskName,
		TasksCompleted:  false,
	})

	if errSend != nil {
		return fmt.Errorf("error err:%v, errSEnd:%v", errTaskExec, errSend)
	}

	return errTaskExec
}

// SendTaskStatusToServer is the means for a remote to send task progress to
// the Server.
func (msg *RemoteMessenger) SendTaskStatusToServer(workflowNameKey, taskStatusMsg string) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	errSend := msg.ConnToSrv.Send(&RemoteMsg{
		WorkflowNameKey: workflowNameKey,
		TaskInProgress:  taskStatusMsg,
		TasksCompleted:  false,
	})

	if errSend != nil {
		return errSend
	}

	return nil
}

// SendDatumToRemote is the means for a remote to send a single string data
// element to the Server.
func (msg *RemoteMessenger) SendDatumToServer(workflowNameKey, datum string) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	errSend := msg.ConnToSrv.Send(&RemoteMsg{
		WorkflowNameKey: workflowNameKey,
		Datum:           datum,
	})

	if errSend != nil {
		return errSend
	}

	return nil
}

// SendDataToRemote is the means for a remote to send a slick of string data
// elements to the Server.
func (msg *RemoteMessenger) SendDataToServer(workflowNameKey string, data []string) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	errSend := msg.ConnToSrv.Send(&RemoteMsg{
		WorkflowNameKey: workflowNameKey,
		Data:            data,
	})

	if errSend != nil {
		return errSend
	}

	return nil
}

// SendTaskCompletionToServer sends tasks completion
func (msg *RemoteMessenger) SendTaskCompletionToServer(workflowNameKey string, tasks []*RemoteMsg_Tasks) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	outMsg := fmt.Sprintf("signaling remoted completion for tasks: %v\n", tasks)
	fmt.Println(outMsg)

	errSend := msg.ConnToSrv.Send(&RemoteMsg{
		WorkflowNameKey: workflowNameKey,
		Tasks:           tasks,
		TasksCompleted:  true,
	})

	if errSend != nil {
		// TODO .Error(ctx, errSend, "failed to signal completion of the remote tasks to the server")

		return errSend
	}

	return nil
}
