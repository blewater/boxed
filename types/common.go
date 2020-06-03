// Package common declares shared types for server and remotes.
package types

import (
	"github.com/tradeline-tech/workflow/wrpc"
)

const WorkflowKey = "workflowKey"

type MsgToRemote interface {
	SendServerTaskProgressToRemote(taskName, msgText string, errIn error) error
	SendRemoteTasksToRun(taskNames []string) error
	SendErrMsgToRemote(textErrorForRemote string, preExistingError error) error
	SignalSrvWorkflowCompletion(tasksLength int) error
	SendDatumToRemote(datum string) error
	SendDataToRemote(data []string) error
}

type TaskConfiguration interface {
	Add(key string, value interface{})
	Get(key string) (interface{}, bool)
}

type MsgToSrv interface {
	SendWorkflowNameKeyToSrv(workflowNameKey string) error
	SendTasksErrorMsgToServer(
		workflowNameKey string,
		taskIndex int,
		currentTaskName string,
		remoteTasks []*wrpc.RemoteMsg_Tasks,
		errTaskExec error) error
	SendTaskStatusToServer(workflowNameKey, taskStatusMsg string) error
	SendDatumToServer(workflowNameKey, datum string) error
	SendDataToServer(workflowNameKey string, data []string) error
	SendTaskCompletionToServer(workflowNameKey string, tasks []*wrpc.RemoteMsg_Tasks) error
}
