package tasks

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// validate is a server executed task and calculates g^y.
type validate struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewValidate returns a server task that calculates g^y on the server.
func NewValidate(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &validate{
		Config: config,
		Task: &types.TaskType{
			Name:       types.GetTaskName(),
			IsServer:   false,
			RunDoFirst: true,
		},
	}

	return taskRunner
}

// Do the task
func (task *validate) Do() error {
	rand.Seed(time.Now().UnixNano())

	if err := task.Validate(); err != nil {
		return err
	}

	gxy, err := secret.GetValue(task.Config, secret.GXtoY)
	if err != nil {
		return err
	}

	gyx, err := secret.GetValue(task.Config, secret.GYtoX)
	if err != nil {
		return err
	}

	if res := secret.SecretIsSame(gyx, gxy); !res {
		return errors.New("validation failed")
	}

	fmt.Println("Successful calculation.")

	return nil
}

// Validate if task completed
func (task *validate) Validate() error {
	_, ok := task.Config.Get(secret.GXtoY)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.GtoX)
	}

	_, ok = task.Config.Get(secret.GYtoX)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.GYtoX)
	}

	return nil
}

// Rollback if task failed
func (task *validate) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *validate) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *validate) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *validate) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}
