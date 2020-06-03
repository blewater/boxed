package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// GenGx is a remotely executed task asking user to pick p, randomly picking g
// and calculating g^x. This is the genesis task and runs remotely
// interactively in stdin/out.
type GenGx struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGx returns a new task that bootstraps a new workflow with a unique name
func NewGenGx(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &GenGx{
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
func (task *GenGx) Do() error {
	var (
		pChoice byte = 101
		err     error
	)
	rand.Seed(time.Now().UnixNano())
	for {
		fmt.Printf("Choose  a Or b Or c:\n\t\t\ta. 101\n\t\t\tb. 113\n\t\t\tc. 199\n")
		if _, err = fmt.Scanf("%c", &pChoice); err == nil && (pChoice == 'a' || pChoice == 'b' || pChoice == 'c') {
			break
		}
	}
	var p int64
	switch pChoice {
	case 'a':
		p = 101
	case 'b':
		p = 113
	case 'c':
		p = 199
	}
	task.Config.Add(secret.P, p)
	roots := getRoots(p)

	// Select a random generator
	g := roots[rand.Int63n(int64(len(roots)))]
	task.Config.Add(secret.G, g)

	x := rand.Int63n(p/2) + (p / 2)
	task.Config.Add(secret.X, x)
	gX := secret.GetModOfPow(g, x, p)

	return remote.SendDataToServer(
		secret.WorkflowNameKey,
		[]string{
			secret.P,
			strconv.FormatInt(p, 10),
			secret.G,
			strconv.FormatInt(g, 10),
			secret.GtoX,
			strconv.FormatInt(gX, 10),
		},
		task.Config)
}

// Validate if task completed
func (task *GenGx) Validate() error {
	_, ok := task.Config.Get(secret.X)
	if !ok {
		return secret.GetValueNotFoundErrFunc(secret.X)
	}

	return nil
}

// Rollback if task failed
func (task *GenGx) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *GenGx) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *GenGx) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *GenGx) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
}

// getRoots returns all the primitive roots modulo nPrime so that g is
// a primitive root modulo n if for every integer a co-prime to n, there is an
// integer k such that gᵏ ≡ a. There is not a known algorithm to find
// them and this is an iterative seek method to find the multiplicative order
// of a modulo n. This is equal to phi(n) (for primes is nPrime) then it is a
// primitive root. g is also called a generator for field Zn (n is prime) such
// that g^1...k enumerates all known values of Zn.
func getRoots(nPrime int64) []int64 {
	var (
		r     int64
		roots []int64
		a     int64 = 1
	)

	for r = 2; r < nPrime; r++ {
		kIdx := secret.GetModOfPow(r, a, nPrime)
		for kIdx > 1 {
			a++
			kIdx = (kIdx * r) % nPrime
		}

		if a == (nPrime - 1) {
			roots = append(roots, r)
		}

		a = 1
	}

	return roots
}
