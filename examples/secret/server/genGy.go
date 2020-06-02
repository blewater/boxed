package main

import (
	"fmt"
	"math/rand"

	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

const (
	G    = "g"
	P    = "p"
	GtoY = "gy"
)

var (
	getValueNotFoundErrFunc = func(v string) error { return fmt.Errorf("%s not found within configuration", v) }
)

// GenGy is a remotely executed task
// bootstrapping a workflow with a name
type GenGy struct {
	Config config.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGy returns a new task that bootstraps a new workflow with a unique name
func NewGenGy(config config.TaskConfiguration) types.TaskRunner {
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

	if err := task.Validate(); err != nil {
		return err
	}
	pi, ok := task.Config.Get(P)
	if !ok {
		return getValueNotFoundErrFunc(P)
	}
	p := pi.(int64)
	fmt.Println(p)

	gi, ok := task.Config.Get(G)
	if !ok {
		return getValueNotFoundErrFunc(G)
	}
	g := gi.(int64)
	fmt.Println(g)

	task.Config.Add(GtoY, getModOfPow(g, rand.Int63n(p), p))

	return nil
}

// Validate if task completed
func (task *GenGy) Validate() error {
	_, ok := task.Config.Get(G)
	if !ok {
		return getValueNotFoundErrFunc(G)
	}
	_, ok = task.Config.Get(P)
	if !ok {
		return getValueNotFoundErrFunc(P)
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

// getModOfPow encrypts or decrypts a value for the RSA algorithm using a simple
// linear complexity Oexp modular exponentiation algorithm such that
// i^exp mod n => ((r1=1^i mod n) * (r2=r1^i mod n)* ... * (rexp^i mod n)) mod n
// if integer is a message m and pow is e then it returns c mod n the encrypted message per RSA
// if integer is an encrypted message c and pow is d then it returns m mod n the decrypted message per RSA
func getModOfPow(integer, exponent, n int64) int64 {
	var res int64 = 1
	for i := res; i <= exponent; i++ {
		res = (res * integer) % n
	}
	return res
}
