package types

import (
	"github.com/tradeline-tech/workflow/wrpc"
)

// remoteGenericTaskForSrv is a placeholder of remotely executed tasks for
// registering custom remote task names within the server workflow with a name.
type remoteGenericTaskForSrv struct {
	Config TaskConfiguration
	Task   *TaskType
}

// RegisterRemoteTask returns a new task runner func to register a remote
// workflow task for the workflow
func RegisterRemoteTask(remoteTaskName string) TaskRunnerNewFunc {

	return func(config TaskConfiguration) TaskRunner {
		taskRunner := &remoteGenericTaskForSrv{
			Task: &TaskType{
				Name:     remoteTaskName,
				IsServer: false,
			},
		}

		return taskRunner
	}
}

// Run the task
func (t *remoteGenericTaskForSrv) Do() error { return nil }

// Validate if task completed
func (t *remoteGenericTaskForSrv) Validate() error { return nil }

// Rollback if task failed
func (t *remoteGenericTaskForSrv) Rollback() error { return nil }

// GetProp returns a task config property
func (t *remoteGenericTaskForSrv) GetProp(key string) (interface{}, bool) { return t.Config.Get(key) }

// GetTask returns the task of this runner
func (t *remoteGenericTaskForSrv) GetTask() *TaskType { return t.Task }

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (t *remoteGenericTaskForSrv) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {}
