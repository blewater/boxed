package tasks

import (
	"github.com/tradeline-tech/workflow/grpc"
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"
)

const (
	ConfigWorkflowName = "WorkflowName"
)

// getWorkflowName is a remotely executed task
// bootstrapping a workflow with a name
type getWorkflowName struct {
	Config config.TaskConfiguration
	Task   *types.TaskType
}

// NewGetWorkflowName returns a new task that bootstraps a new workflow with a unique name
func NewGetWorkflowName(config config.TaskConfiguration) types.TaskRunner {
	taskRunner := &getWorkflowName{
		Config: config,
		Task: &types.TaskType{
			Name:     types.GetTaskName(),
			IsServer: false,
		},
	}

	return taskRunner
}

// Do the task
func (t *getWorkflowName) Do() error {
	return nil
}

// Validate if task completed
func (t *getWorkflowName) Validate() error {
	return nil
}

// Rollback if task failed
func (t *getWorkflowName) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (t *getWorkflowName) GetProp(key string) (interface{}, bool) {
	return t.Config.Get(key)
}

// GetTask returns the task of this runner
func (t *getWorkflowName) GetTask() *types.TaskType {
	return t.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (t *getWorkflowName) PostRemoteTasksCompletion(msg *grpc.RemoteMsg) {
	t.Config.Add(ConfigWorkflowName, msg.Datum)
}
