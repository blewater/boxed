package main

import (
	"fmt"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// GenGyx is a remotely executed task
// bootstrapping a workflow with a name
type GenGyx struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGxy returns a new task that calculates G^(y*x)
func NewGenGyx(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &GenGyx{
		Config: config,
		Task: &types.TaskType{
			Name:     types.GetTaskName(),
			IsServer: false,
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

	gYX := secret.GetModOfPow(gY, x, p)
	var (
		guess, hits int64
	)
	i := 0
	for ; i < 3; i++ {
		fmt.Printf("Can you guess the exchanged secret? (trial %d)\n", i+1)
		if _, err = fmt.Scanln(&guess); err != nil {
			fmt.Println("error : ", err, ", input expected to be a positive integer")
		}
		if gYX == guess {
			hits++
		}
	}
	if hits == 0 {
		fmt.Println("Sorry better luck next time :)")
		return nil
	}
	switch hits {
	case 1:
		fmt.Println("Good job! Beginner's luck? Have another go for a chance to beat this.")
	case 2, 3:
		fmt.Println("Excellent job! Please share your coding solution to feature it on this site :)")
	}

	return nil
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
