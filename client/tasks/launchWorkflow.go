package tasks

import (
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

const (
	LaunchWorkflowKey = "launchworkflow"
)

// LaunchWorkflow is a remotely executed task
// bootstrapping a workflow with a name
type LaunchWorkflow struct {
	Config config.TaskConfiguration
	Task   *types.TaskType
}

// NewLaunchWorkflow returns a new task that bootstraps a new workflow with a unique name
func NewLaunchWorkflow(config config.TaskConfiguration) types.TaskRunner {
	taskRunner := &LaunchWorkflow{
		Config: config,
		Task: &types.TaskType{
			Name:     types.GetTaskName(),
			IsServer: false,
		},
	}

	return taskRunner
}

// Do the task
func (t *LaunchWorkflow) Do() error {
	return nil
}

// Validate if task completed
func (t *LaunchWorkflow) Validate() error {
	return nil
}

// Rollback if task failed
func (t *LaunchWorkflow) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (t *LaunchWorkflow) GetProp(key string) (interface{}, bool) {
	return t.Config.Get(key)
}

// GetTask returns this runner's task
func (t *LaunchWorkflow) GetTask() *types.TaskType {
	return t.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (t *LaunchWorkflow) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}
