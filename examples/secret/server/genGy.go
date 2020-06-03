package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/server"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// GenGy is a remotely executed task
// bootstrapping a workflow with a name
type GenGy struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGy returns a new task that bootstraps a new workflow with a unique name
func NewGenGy(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &GenGy{
		Config: config,
		Task: &types.TaskType{
			Name:     types.GetTaskName(),
			IsServer: true,
		},
	}

	return taskRunner
}

// Do the task
func (task *GenGy) Do() error {
	rand.Seed(time.Now().UnixNano())

	if err := task.Validate(); err != nil {
		return err
	}
	p, err := secret.GetValue(task.Config, secret.P)
	if err != nil {
		return err
	}

	g, err := secret.GetValue(task.Config, secret.G)
	if err != nil {
		return err
	}

	y := rand.Int63n(p-1) + 2
	task.Config.Add(secret.Y, y)

	gY := secret.GetModOfPow(g, y, p)

	return server.SendDataToServer(
		secret.WorkflowNameKey,
		[]string{
			secret.GtoY,
			strconv.FormatInt(gY, 10),
		},
		task.Config)
}

// Validate if task completed
func (task *GenGy) Validate() error {
	_, ok := task.Config.Get(secret.G)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.G)
	}
	_, ok = task.Config.Get(secret.P)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.P)
	}
	_, ok = task.Config.Get(secret.GtoX)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.GtoX)
	}

	return nil
}

// Rollback if task failed
func (task *GenGy) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *GenGy) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *GenGy) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *GenGy) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}
