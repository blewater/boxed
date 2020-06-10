package server

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// genGy is a server executed task and calculates g^y.
type genGy struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGy returns a server task that calculates g^y on the server.
func NewGenGy(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &genGy{
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
func (task *genGy) Do() error {
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

	return SendDataToRemote([]string{
		secret.GtoY,
		strconv.FormatInt(gY, 10),
	}, task.Config)
}

// Validate if task completed
func (task *genGy) Validate() error {
	_, ok := task.Config.Get(secret.G)
	if !ok {
		return config.GetValueNotFoundErrFunc(secret.G)
	}
	_, ok = task.Config.Get(secret.P)
	if !ok {
		return config.GetValueNotFoundErrFunc(secret.P)
	}
	_, ok = task.Config.Get(secret.GtoX)
	if !ok {
		return config.GetValueNotFoundErrFunc(secret.GtoX)
	}

	return nil
}

// Rollback if task failed
func (task *genGy) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *genGy) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *genGy) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *genGy) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}
