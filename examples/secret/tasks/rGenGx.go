package tasks

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/tradeline-tech/workflow/examples/secret"
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

// genGx is a remotely executed task asking user to pick p, randomly picking g
// and calculating g^x. This is the genesis task and runs remotely
// interactively in stdin/out.
type genGx struct {
	Config types.TaskConfiguration
	Task   *types.TaskType
}

// NewGenGx returns a new task that bootstraps a new workflow with a unique name
func NewGenGx(config types.TaskConfiguration) types.TaskRunner {
	taskRunner := &genGx{
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
func (task *genGx) Do() error {
	rand.Seed(time.Now().UnixNano())

	printIntro()

	p := getPrimeFromUser()

	task.Config.Add(secret.P, p)

	roots := getRoots(p)

	// Select a random generator g from the
	// primitive roots of P.
	g := roots[rand.Int63n(int64(len(roots)))]
	task.Config.Add(secret.G, g)

	// Choose random x and save for future use
	// when we receive g^y
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

func getPrimeFromUser() int64 {
	var (
		pChoice byte = 101
		err     error
	)

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
	return p
}

func printIntro() {
	fmt.Printf("\n%s\n%s\n%s\n%s\n\n",
		"This workflow employs a secure method to share a common secret between two parties.",
		"Despite the usage of randomizer for seeding, the final result is a shared secret",
		"calculated differently but producing the same result in the server and remote party.",
		"Your choice below determines the upper limit of the final answer: 1...(n-1).")
}

// Validate if task completed
func (task *genGx) Validate() error {
	_, ok := task.Config.Get(secret.X)
	if !ok {
		return config.GetValueNotFoundErrFunc(secret.X)
	}

	return nil
}

// Rollback if task failed
func (task *genGx) Rollback() error {
	return nil
}

// GetProp returns a task config property
func (task *genGx) GetProp(key string) (interface{}, bool) {
	return task.Config.Get(key)
}

// GetTask returns this runner's task
func (task *genGx) GetTask() *types.TaskType {
	return task.Task
}

// PostRemoteTasksCompletion performs any server workflow task work upon
// completing the remote task work e.g., saving remote task configuration
// to workflow's state
func (task *genGx) PostRemoteTasksCompletion(msg *wrpc.RemoteMsg) {
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
