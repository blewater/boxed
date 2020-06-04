package tasks

import (
	"fmt"
	"strconv"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// GenGyx is a remotely executed task and calculates g^yx the shared secret.
type GenGyx struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGxy returns a new task that calculates g^(y*x)
func NewGenGyx(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &GenGyx{
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
func (task *GenGyx) Do() error {
	if err := task.Validate(); err != nil {
		return err
	}

	p, err := secret.GetValue(task.Config, secret.P)
	if err != nil {
		return err
	}

	gY, err := secret.GetValue(task.Config, secret.GtoY)
	if err != nil {
		return err
	}

	x, err := secret.GetValue(task.Config, secret.X)
	if err != nil {
		return err
	}

	gyx := secret.GetModOfPow(gY, x, p)
	var (
		guess, hits int64
	)
	for i := 0; i < 3; i++ {
		fmt.Printf("3 tries to guess the exchanged secret? (trial %d) : ", i+1)
		if i == 0 {
			_, err = fmt.Scanf("\n%d\n", &guess)
		} else {
			_, err = fmt.Scanf("%d\n", &guess)

		}
		if err != nil {
			fmt.Println(err, ", input expected to be a positive integer")
		}
		if gyx == guess {
			hits++
		}
	}
	fmt.Printf("The secret is --> %d\n", gyx)
	switch hits {
	case 0:
		fmt.Println("Sorry better luck next time :)")
	case 1:
		fmt.Println("Good job! Beginner's luck? Have another go for a chance to beat this :)")
	case 2, 3:
		fmt.Println("Excellent job! Please share your coding solution to feature it on this site :)")
	}

	return remote.SendDataToServer(
		secret.WorkflowNameKey,
		[]string{
			secret.GYtoX,
			strconv.FormatInt(gyx, 10),
		},
		task.Config)
}

// Validate if task completed
func (task *GenGyx) Validate() error {
	_, ok := task.Config.Get(secret.G)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.G)
	}

	_, ok = task.Config.Get(secret.X)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.X)
	}

	_, ok = task.Config.Get(secret.GtoY)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.GtoY)
	}

	return nil
}

// Rollback if task failed
func (task *GenGyx) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *GenGyx) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *GenGyx) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *GenGyx) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}
