package tasks

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/server"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// GenGxy is a server executed task and calculates g^y.
type GenGxy struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGxy returns a server task that calculates g^y on the server.
func NewGenGxy(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &GenGxy{
		Config: config,
		Task: &types.TaskType{
			Name:       types.GetTaskName(),
			IsServer:   true,
			RunDoFirst: true,
		},
	}

	return taskRunner
}

// Do the task
func (task *GenGxy) Do() error {
	rand.Seed(time.Now().UnixNano())

	if err := task.Validate(); err != nil {
		return err
	}
	p, err := secret.GetValue(task.Config, secret.P)
	if err != nil {
		return err
	}

	y, err := secret.GetValue(task.Config, secret.Y)
	if err != nil {
		return err
	}

	gx, err := secret.GetValue(task.Config, secret.GtoX)
	if err != nil {
		return err
	}

	gyx, err := secret.GetValue(task.Config, secret.GYtoX)
	if err != nil {
		return err
	}

	gxy := secret.GetModOfPow(gx, y, p)

	return server.SendDataToRemote([]string{
		secret.IsSecretEq,
		strconv.FormatBool(secret.SecretIsSame(gyx, gxy)),
		secret.GYtoX,
		strconv.FormatInt(gyx, 10),
		secret.GXtoY,
		strconv.FormatInt(gxy, 10),
	}, task.Config)
}

// Validate if task completed
func (task *GenGxy) Validate() error {
	_, ok := task.Config.Get(secret.P)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.P)
	}

	_, ok = task.Config.Get(secret.Y)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.Y)
	}

	_, ok = task.Config.Get(secret.GYtoX)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.GYtoX)
	}

	_, ok = task.Config.Get(secret.GtoX)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.GtoX)
	}

	return nil
}

// Rollback if task failed
func (task *GenGxy) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *GenGxy) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *GenGxy) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *GenGxy) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}
