package tasks

import (
	"github.com/tradeline-tech/workflow"
	"github.com/tradeline-tech/workflow/cfg"
	"github.com/tradeline-tech/workflow/grpc"
)

const (
	ConfigWorkflowName = "WorkflowName"
)

// getWorkflowName is a remotely executed task
// bootstrapping a workflow with a name
type getWorkflowName struct {
	Config *cfg.TasksBootstrapConfiguration
	Task   *workflow.TaskType
}

// NewGetWorkflowName returns a new task that bootstraps a new workflow with a unique name
func NewGetWorkflowName(config *cfg.TasksBootstrapConfiguration) workflow.TaskRunner {
	taskRunner := &getWorkflowName{
		Config: config,
		Task: &workflow.TaskType{
			Name:     workflow.GetTaskName(),
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

// GetProp of this task
func (t *getWorkflowName) GetProp() *workflow.TaskType {
	return t.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (t *getWorkflowName) PostRemoteTasksCompletion(msg *grpc.RemoteMsg) {
	t.Config.Add(ConfigWorkflowName, msg.Datum)
}
