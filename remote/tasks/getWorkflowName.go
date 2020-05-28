package tasks

import (
	"github.com/tradeline-tech/workflow/common"
	"github.com/tradeline-tech/workflow/grpc"
	"github.com/tradeline-tech/workflow/pkg/config"
)

const (
	ConfigWorkflowName = "WorkflowName"
)

// getWorkflowName is a remotely executed task
// bootstrapping a workflow with a name
type getWorkflowName struct {
	Config config.TaskConfiguration
	Task   *common.TaskType
}

// NewGetWorkflowName returns a new task that bootstraps a new workflow with a unique name
func NewGetWorkflowName(config config.TaskConfiguration) common.TaskRunner {
	taskRunner := &getWorkflowName{
		Config: config,
		Task: &common.TaskType{
			Name:     common.GetTaskName(),
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

// GetTask returns this runner's task
func (t *getWorkflowName) GetTask() *common.TaskType {
	return t.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (t *getWorkflowName) PostRemoteTasksCompletion(msg *grpc.RemoteMsg) {
}
