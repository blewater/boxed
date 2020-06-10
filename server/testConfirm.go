package server

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// confirm is a server executed task and calculates g^y.
type confirm struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewConfirm returns a server task that calculates g^y on the server.
func NewConfirm(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &confirm{
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
func (task *confirm) Do() error {
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

	if res := secret.IsSame(gyx, gxy); !res {
		panic("confirmation of final result failed, " +
			strconv.FormatInt(gyx, 10) + ", " + strconv.FormatInt(gxy, 10))
	}

	fmt.Println("Successful calculation.")

	return nil
}

// Validate if task completed
func (task *confirm) Validate() error {
	_, ok := task.Config.Get(secret.GXtoY)
	if !ok {
		return config.GetValueNotFoundErrFunc(secret.GtoX)
	}

	_, ok = task.Config.Get(secret.GYtoX)
	if !ok {
		return config.GetValueNotFoundErrFunc(secret.GYtoX)
	}

	return nil
}

// Rollback if task failed
func (task *confirm) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *confirm) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *confirm) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *confirm) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}
