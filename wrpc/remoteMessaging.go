package wrpc

import (
	"context"
	"errors"
	"fmt"
	"log"
)

const (
	gRPCKey                       = "grpc"
	errTextWorkflowKeyMissing     = "remote workflow name key required for messaging server apps"
	remoteTextWorkflowNameMissing = "missing remote workflow name key value to identify the workflow at server."
)

//
// Collection of remote 2 Server gRPC message functions
//

// SendWorkflowNameKeyToSrv sends the org identifier to the server
func SendWorkflowNameKeyToSrv(stream TaskCommunicator_RunWorkflowClient, workflowNameKey string) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	if err := stream.Send(&RemoteMsg{WorkflowNameKey: workflowNameKey}); err != nil {
		log.Println("error : ", err, ", failed to stream a remote gRPC message to server")

		return err
	}

	return nil
}

// SendTasksErrorMsgToServer sends error message to the Server for current task execution error
func SendTasksErrorMsgToServer(ctx context.Context,
	gRPCConn TaskCommunicator_RunWorkflowClient,
	workflowNameKey string,
	taskIndex int,
	currentTaskName string,
	remoteTasks []*RemoteMsg_Tasks,
	errTaskExec error) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	errSend := gRPCConn.Send(&RemoteMsg{
		ErrorMsg: fmt.Sprintf("error %v, remote tasks index: %d, task name: %s\n",
			errTaskExec, taskIndex, currentTaskName),
		WorkflowNameKey: workflowNameKey,
		Tasks:           remoteTasks,
		TaskInProgress:  currentTaskName,
		TasksCompleted:  false,
	})

	if errSend != nil {
		// TODO .Error(ctx, errSend, "failed to send error to server", err)

		return fmt.Errorf("error err:%v, errSEnd:%v", errTaskExec, errSend)
	}

	return errTaskExec
}

// SendMsgToServer sends error message to the Server for current task execution error
func SendMsgToServer(ctx context.Context,
	stream TaskCommunicator_RunWorkflowClient,
	workflowNameKey, taskStatusMsg string) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	errSend := stream.Send(&RemoteMsg{
		WorkflowNameKey: workflowNameKey,
		TaskInProgress:  taskStatusMsg,
		TasksCompleted:  false,
	})

	if errSend != nil {
		// TODO .Error(ctx, errSend, "failed to send task status message")

		return errSend
	}

	return nil
}

// SendTaskCompletionToServer sends tasks completion
func SendTaskCompletionToServer(ctx context.Context,
	stream TaskCommunicator_RunWorkflowClient,
	workflowNameKey string, tasks []*RemoteMsg_Tasks) error {
	if workflowNameKey == "" {
		fmt.Println(remoteTextWorkflowNameMissing)

		return errors.New(errTextWorkflowKeyMissing)
	}

	outMsg := fmt.Sprintf("signaling remoted completion for tasks: %v\n", tasks)
	// TODO .Debug(ctx, outMsg)
	fmt.Println(outMsg)

	errSend := stream.Send(&RemoteMsg{
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