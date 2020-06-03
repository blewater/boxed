package types

import (
	"fmt"
	"time"
)

// TaskType represents a unit of work that needs to happen on the server or client side
type TaskType struct {
	CompletedAt time.Time `bson:"completedAt" json:"completedAt"`
	// index for internal tracking
	Name     string `bson:"name" json:"name"`
	Log      string `bson:"log" json:"log"`
	IsServer bool   `bson:"isServer" json:"isServer"`
	// False: run Validate() first and only upon error, run Do(), Validate()
	// True: run Do() first then Validate().
	RunDoFirst bool `bson:"alwaysRun" json:"alwaysRun"`
	Completed  bool `bson:"completed" json:"completed"`
}

type HandlerFuncType func() error

func (t *TaskType) Print() {
	fmt.Printf(
		"Task name: %s, isServer: #{task.IsServer},  description #{task.Description}", t.Name)

	if t.Completed {
		fmt.Printf(", not completed yet.\n")

		return
	}

	fmt.Println(", is done!")
}

// doWithRollback executes a task's Do() with Rollback upon failure and returns combined errors
func doAndRollback(do, rollback HandlerFuncType) error {
	if doErr := do(); doErr != nil {
		if rollbackErr := rollback(); rollbackErr != nil {
			return fmt.Errorf("failed with %s and on rollback: %s", doErr, rollbackErr)
		}

		return doErr
	}

	return nil
}

// ValidDo calls these task runners functions
// 1.
// Validate() to check if the task is needed to run,
// 2.
// Upon error -> Do()
// 3.
// upon error -> Rollback(),
// upon success Validate()
func ValidDo(runner TaskRunner) error {
	task := runner.GetTask()

	// If is a mandatory task do not validate in the beginning
	if !task.RunDoFirst {
		if errValidation := runner.Validate(); errValidation == nil {
			task.setCompleted()
			fmt.Println("Validate() succeeded exiting...")

			return nil
		}
	}

	if err := doAndRollback(runner.Do, runner.Rollback); err != nil {
		task.Log = err.Error()
		return err
	}

	if err := runner.Validate(); err != nil {
		task.Log = err.Error()
		return err
	}

	// Set task complete if validation completed
	task.setCompleted()

	return nil
}

func (t *TaskType) GetTaskTypeDesc() string {
	if t.IsServer {
		return "server"
	}

	return "Remote"
}

func (t *TaskType) setCompleted() {
	if !t.Completed {
		t.CompletedAt = time.Now()
		t.Completed = true
	}
}
