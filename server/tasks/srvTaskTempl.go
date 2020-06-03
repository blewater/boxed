package tasks

import (
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// remoteGenericTaskForSrv is a placeholder of remotely executed tasks for
// registering custom remote task names within the server workflow with a name.
type remoteGenericTaskForSrv struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGetWorkflowName returns a new task runner func to register a remote
// workflow task for the server workflow
func NewRemoteGenericTaskForSrv(remoteTaskName string) types.TaskRunnerNewFunc {

	return func(config types.TaskConfiguration) types.TaskRunner {
		taskRunner := &remoteGenericTaskForSrv{
			Task: &types.TaskType{
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
func (t *remoteGenericTaskForSrv) GetTask() *types.TaskType { return t.Task }

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (t *remoteGenericTaskForSrv) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {}
