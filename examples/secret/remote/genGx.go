package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

const (
	GtoX = "gx"
)

// GenGy is a remotely executed task
// bootstrapping a workflow with a name
type GenGy struct {
	Config config.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGx returns a new task that bootstraps a new workflow with a unique name
func NewGenGx(config config.TaskConfiguration) types.TaskRunner {
	taskRunner := &GenGy{
		Config: config,
		Task: &types.TaskType{
			Name:     types.GetTaskName(),
			IsServer: false,
		},
	}

	return taskRunner
}

// Do the task
func (task *GenGy) Do() error {
	var (
		pChoice byte
		err     error
	)
	rand.Seed(time.Now().UnixNano())
	for {
		fmt.Printf(`Choose one prime: a Or b Or c\n
										a. 101\n
										b. 113\n
										c. 199\n`)
		if _, err = fmt.Scanln(pChoice); err == nil && (pChoice == 'a' || pChoice == 'b' || pChoice == 'c') {
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
	roots := getRoots(p)

	// Select a random generator
	g := roots[rand.Int63n(int64(len(roots)))]

	x := rand.Int63n(p/2) + (p / 2)
	gX := getModOfPow(g, x, p)

	return remote.SendDatumToServer(
		WorkflowNameKey,
		strconv.FormatInt(gX, 10),
		task.Config)
}

// Validate if task completed
func (task *GenGy) Validate() error {
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

// getModOfPow returns the modulo of raising an integer ^ exponent
// in a simple and moderately optimized manner using the property
// i^e%n => ((r1=1^i mod n) * (r2=r1^i mod n)* ... * (rexp^i mod n)) mod n
func getModOfPow(integer, exponent, n int64) int64 {
	var res int64 = 1
	for i := res; i <= exponent; i++ {
		res = (res * integer) % n
	}

	return res
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
		kIdx := getModOfPow(r, a, nPrime)
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
