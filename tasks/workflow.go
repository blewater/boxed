package tasks

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/tradeline-tech/protos/namespace"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/msg"

)

// WorkflowType is the model for the network workflow collection of tasks.
type WorkflowType struct {
	ID                     primitive.ObjectID          `bson:"_id" json:"_id"`
	// task names persisted in mongo
	Tasks []*TaskType `bson:"tasks" json:"tasks"`
	TasksConfig            *TasksBootstrapConfiguration `bson:"tasksConfig" json:"tasksConfig"`
	// task runtime biz logic implementers: Do, Validate, Rollback
	LastTaskIndexCompleted int                         `bson:"lastTaskIndexCompleted" json:"lastTaskIndexCompleted"`
	LastTaskNameCompleted  string                      `bson:"lastTaskNameCompleted" json:"lastTaskNameCompleted"`
	Completed              bool                        `bson:"completed" json:"completed"`
	CompletedAt            time.Time                   `bson:"completedAt" json:"completedAt"`
	CreatedAt              time.Time                   `bson:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time                   `bson:"updatedAt" json:"updatedAt"`
	// Task memory resident initialized interfaces
	TaskRunners            []TaskRunner                `bson:"-" json:"-"`
}

// NewWorkflow initializes the workflow struct
func NewWorkflow(config *TasksBootstrapConfiguration, token string, tasksRunnerNewFunc []TaskRunnerNewFunc) *WorkflowType {
	workflowTasks := make([]*TaskType, 0, 10)

	workflowTaskRunners := make([]TaskRunner, 0, 10)

	for _, taskRunner := range tasksRunnerNewFunc {
		newTaskRunner := taskRunner(config)
		newTask := newTaskRunner.GetProp()

		workflowTaskRunners = append(workflowTaskRunners, newTaskRunner)
		workflowTasks = append(workflowTasks, newTask)
	}

	return &WorkflowType{
		ID:                     primitive.NilObjectID,
		Tasks:                  workflowTasks,
		TaskRunners:            workflowTaskRunners,
		LastTaskIndexCompleted: -1,
		TasksConfig:            config,
	}
}

// ReSetMemoryState updates an existing workflow with the task runners that can be created only in memory
func (workflow *WorkflowType) ReSetMemoryState(config *TasksBootstrapConfiguration,
	tasksRunnerNewFunc []TaskRunnerNewFunc) error {
	if len(tasksRunnerNewFunc) != len(workflow.Tasks) {
		return fmt.Errorf(`error workflow tasks length: %d != task runners length: %d\n
			Adjust your constructor task runners to match the database workflow tasks with ID: %v`,
			len(workflow.Tasks), len(tasksRunnerNewFunc), workflow.ID)
	}

	workflow.TasksConfig = config
	workflowTaskRunners := make([]TaskRunner, 0, 10)

	for idx, taskRunner := range tasksRunnerNewFunc {
		newTaskRunner := taskRunner(workflow.TasksConfig)

		dbTask := workflow.Tasks[idx]
		newMemoryTask := newTaskRunner.GetProp()
		// Copy values not reference
		*newMemoryTask = *dbTask

		workflowTaskRunners = append(workflowTaskRunners, newTaskRunner)
	}

	workflow.TaskRunners = workflowTaskRunners

	workflow.LastTaskIndexCompleted = -1
	workflow.Completed = false

	return nil
}

// Run is the workflow runner for server tasks;
// interrupts execution for Cli tasks
func (workflow *WorkflowType) Run(ctx context.Context) (execErr error) {
	if workflow.SetWorkflowCompletedChecked(ctx) {
		return nil
	}

	// Continuing with remaining workflow work
	return workflow.doRemainingTasks(ctx)
}

// SendTaskUpdateToRemote streams a server task text update on progress or error to the remote client
func (workflow *WorkflowType) SendTaskUpdateToRemote(taskIndex int, msgText string, errIn error) error {
	if taskIndex < 0 || taskIndex >= len(workflow.Tasks) {
		return nil
	}

	taskName := workflow.Tasks[taskIndex].Name

	return msg.TaskUpdateToRemote(GetConfig().stream, taskName, msgText, errIn)
}

func (workflow *WorkflowType) doRemainingTasks(ctx context.Context) error {
	var messagingErr error
	i := workflow.LastTaskIndexCompleted + 1

	for ; i < len(workflow.TaskRunners); i++ {
		nextTaskRunner := workflow.TaskRunners[i]

		if !nextTaskRunner.GetProp().IsServer {
			return nil
		}

		messagingErr = workflow.SendTaskUpdateToRemote(i,
			fmt.Sprintf("\nResuming server task %s....\n", workflow.Tasks[i].Name),
			nil)

		err := ValidDo(nextTaskRunner)
		workflow.Tasks[i] = nextTaskRunner.GetProp()

		if err != nil {
			return err
		}

		messagingErr = workflow.SendTaskUpdateToRemote(i,
			fmt.Sprintf("\nTask %s completed.\n", nextTaskRunner.GetProp().Name),
			nil)

		workflow.LastTaskIndexCompleted = i
		workflow.LastTaskNameCompleted = nextTaskRunner.GetProp().Name

		if err = datastore.Save(ctx, workflow); err != nil {
			return err
		}
	}

	return messagingErr
}

// CopyCliTasksProgress saves Cli tasks feedback to server workflow
// it loops through Cli tasks and copies progress to server workflow till end of Cli tasks length
// or first unsuccessful cli task
func (workflow *WorkflowType) CopyCliTasksProgress(clientMsg *namespace.ClientMsg) error {
	if clientMsg.Tasks == nil || len(clientMsg.Tasks) == 0 {
		return errors.New("error nil cli tasks")
	}

	for cliTaskIndex := 0; workflow.LastTaskIndexCompleted+1 < len(workflow.Tasks) &&
		cliTaskIndex < len(clientMsg.Tasks); cliTaskIndex++ {
		nextWorkflowTask := workflow.Tasks[workflow.LastTaskIndexCompleted+1]
		nextCliTask := clientMsg.Tasks[cliTaskIndex]

		if nextWorkflowTask.IsServer || !strings.EqualFold(nextWorkflowTask.Name, nextCliTask.TaskName) {
			return fmt.Errorf("error mismatch between next Cli task %s, Workflow task %s",
				nextCliTask.TaskName, nextWorkflowTask.Name)
		}

		nextWorkflowTask.Log = clientMsg.Tasks[cliTaskIndex].ErrorMsg
		nextWorkflowTask.Completed = clientMsg.Tasks[cliTaskIndex].Completed

		if nextWorkflowTask.Log != "" || !nextWorkflowTask.Completed {
			return errors.New(nextWorkflowTask.Log)
		}

		nextWorkflowTask.CompletedAt = time.Now()
		workflow.LastTaskNameCompleted = nextCliTask.TaskName
		workflow.LastTaskIndexCompleted++

		workflow.saveCliConfigResults(clientMsg)
	}

	return nil
}

// Save Cli Tasks TasksConfig Results Here
// Cli Task config answers -> workflow changes of the current tasks[LastIndexCompleted]
// Checks if it's a Cli task and completed
func (workflow *WorkflowType) saveCliConfigResults(clientMsg *namespace.ClientMsg) {
	cliTask := workflow.Tasks[workflow.LastTaskIndexCompleted]
	if cliTask.Completed && !cliTask.IsServer {
		cliTaskRunner := workflow.TaskRunners[workflow.LastTaskIndexCompleted]
		cliTaskRunner.PostCli(clientMsg)
	}
}

// SetWorkflowCompletedChecked checks if the lastTaskIndexCompleted has progressed past
// all the completed tasks and returns T/F whether the workflow has completed
func (workflow *WorkflowType) SetWorkflowCompletedChecked(ctx context.Context) bool {
	if workflow.Completed {
		return true
	}

	if workflow.LastTaskIndexCompleted >= len(workflow.Tasks)-1 {
		workflow.CompletedAt = time.Now()
		workflow.Completed = true
		workflow.LastTaskIndexCompleted = len(workflow.Tasks) - 1

		return true
	}

	return false
}

// GetTaskName gets the string function name for the current goroutine
// executing func in stack frame (-1)
// trimming path/package characters up to function name accounting that for first 3 characters
// to discard "New" so xxx/workflow/NewInitCli becomes InitCli
// Re: https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
func GetTaskName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	parts := strings.Split(fn.Name(), ".")

	// remove new from name
	if len(parts) > 0 {
		return parts[len(parts)-1][3:]
	}

	return ""
}

func (workflow *WorkflowType) PrintTasks() {
	for _, task := range workflow.Tasks {
		task.Print()
	}
}
