package types

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tradeline-tech/workflow/ascii"
	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/wrpc"
)

type TaskRunner interface {
	Do() error
	Validate() error
	Rollback() error
	// GetProp returns a task config property
	GetProp(key string) (interface{}, bool)
	// GetTask returns this runner's task
	// TaskType contains the task data we save in mongo i.e. Name
	GetTask() *TaskType
	// PostRemoteTasksCompletion performs any server workflow task work upon
	// completing the remote task work e.g., saving remote task configuration
	// to workflow's state
	PostRemoteTasksCompletion(msg *wrpc.RemoteMsg)
}

// TaskRunnerNewFunc is a workflow task runner constructor
type TaskRunnerNewFunc = func(cfg TaskConfiguration) TaskRunner
type TaskRunners = []TaskRunnerNewFunc

// Workflow is the model for the network workflow collection of tasks.
type Workflow struct {

	// Persisted unique managed automatically by mongo driver
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// Unique user provided workflow Id, and employed as key within server workflows map.
	Name                   string            `bson:"name" json:"name"`
	Tasks                  []*TaskType       `bson:"tasks" json:"tasks"`
	TasksConfig            TaskConfiguration `bson:"tasksConfig" json:"tasksConfig"`
	LastTaskIndexCompleted int               `bson:"lastTaskIndexCompleted" json:"lastTaskIndexCompleted"`
	LastTaskNameCompleted  string            `bson:"lastTaskNameCompleted" json:"lastTaskNameCompleted"`
	Completed              bool              `bson:"completed" json:"completed"`
	CompletedAt            time.Time         `bson:"completedAt" json:"completedAt"`
	CreatedAt              time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time         `bson:"updatedAt" json:"updatedAt"`

	// Memory transient interfaces
	TaskRunners  []TaskRunner `bson:"-"`
	srvMessenger MsgToRemote  `bson:"-"`
}

// NewWorkflow gets a new initialized workflow type instance. It is a collection of tasks.
func NewWorkflow(
	cfg TaskConfiguration,
	srvMessenger MsgToRemote,
	workflowName string,
	tasksRunners TaskRunners) (*Workflow, error) {
	if workflowName == "" {
		return nil, errors.New("empty workflow workflowName is not allowed")
	}
	workflowTasks := make([]*TaskType, 0, 10)

	workflowTaskRunners := make([]TaskRunner, 0, 10)

	for _, taskRunner := range tasksRunners {
		newTaskRunner := taskRunner(cfg)
		newTask := newTaskRunner.GetTask()

		workflowTaskRunners = append(workflowTaskRunners, newTaskRunner)
		workflowTasks = append(workflowTasks, newTask)
	}

	return &Workflow{
		ID:                     primitive.NilObjectID,
		Name:                   workflowName,
		Tasks:                  workflowTasks,
		TaskRunners:            workflowTaskRunners,
		LastTaskIndexCompleted: -1,
		TasksConfig:            cfg,
		srvMessenger:           srvMessenger,
	}, nil
}

// DBDocument interface methods
func (workflow *Workflow) GetID() interface{} {
	return workflow.ID
}

func (workflow *Workflow) SetID(id string) error {

	mongoID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		workflow.ID = mongoID
	}

	return err
}

func (workflow *Workflow) SetCreatedAt(createdAt time.Time) {
	workflow.CreatedAt = createdAt
}

func (workflow *Workflow) SetUpdatedAt(updatedAt time.Time) {
	workflow.UpdatedAt = updatedAt
}

// GetLen returns the workflow length of contained tasks.
func (workflow *Workflow) GetLen() int {
	return len(workflow.Tasks)
}

// GetPendingRemoteTaskNames returns the task names due for execution.
func (workflow *Workflow) GetPendingRemoteTaskNames() []string {
	remoteTaskNames := make([]string, 0, 2)

	if workflow == nil || len(workflow.Tasks) == 0 {
		return remoteTaskNames
	}

	for i := workflow.LastTaskIndexCompleted + 1; i < len(workflow.Tasks); i++ {
		task := workflow.Tasks[i]

		if task.IsServer {
			break
		}

		remoteTaskNames = append(remoteTaskNames, strings.ToLower(task.Name))
	}

	return remoteTaskNames
}

// InitTasksMemState updates an existing workflow with the task runners that
// can be created only in memory.
func (workflow *Workflow) InitTasksMemState(config TaskConfiguration,
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
		newMemoryTask := newTaskRunner.GetTask()
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
// interrupts execution for Remote tasks
func (workflow *Workflow) Run(ctx context.Context) (execErr error) {
	if workflow.SetWorkflowCompletedChecked(ctx) {
		return nil
	}

	// Continuing with remaining workflow work
	return workflow.doRemainingTasks(ctx)
}

// SendTaskUpdateToRemote streams a server task text update on progress or
// error to the remote client.
func (workflow *Workflow) SendTaskUpdateToRemote(taskIndex int, msgText string, errIn error) error {
	if taskIndex < 0 || taskIndex >= len(workflow.Tasks) {
		return nil
	}

	runner := workflow.TaskRunners[taskIndex]
	taskName := runner.GetTask().Name

	return workflow.srvMessenger.SendServerTaskProgressToRemote(taskName, msgText, errIn)
}

func (workflow *Workflow) doRemainingTasks(ctx context.Context) error {
	var messagingErr error
	i := workflow.LastTaskIndexCompleted + 1

	for ; i < len(workflow.TaskRunners); i++ {
		nextTaskRunner := workflow.TaskRunners[i]

		if !nextTaskRunner.GetTask().IsServer {
			return nil
		}

		messagingErr = workflow.SendTaskUpdateToRemote(i,
			fmt.Sprintf("Running server task %s....", workflow.Tasks[i].Name),
			nil)
		log.ErrorLog(messagingErr)

		err := ValidDo(nextTaskRunner)
		workflow.Tasks[i] = nextTaskRunner.GetTask()

		workflow.Display()

		if err != nil {
			return err
		}

		messagingErr = workflow.SendTaskUpdateToRemote(i,
			fmt.Sprintf("\nTask %s completed.\n", nextTaskRunner.GetTask().Name),
			nil)

		workflow.LastTaskIndexCompleted = i
		workflow.LastTaskNameCompleted = nextTaskRunner.GetTask().Name

		if err = datastore.Upsert(ctx, workflow); err != nil {
			return err
		}
	}

	return messagingErr
}

// SendRemoteTasks sends tasks that need to be executed remotely.
func (workflow *Workflow) SendRemoteTasksToRun() (sentRemoteTasks bool, err error) {
	if workflow.LastTaskIndexCompleted+1 >= workflow.GetLen() ||
		workflow.Tasks[workflow.LastTaskIndexCompleted+1].IsServer {
		return false, nil
	}

	remoteTaskNames := make([]string, 0, 2)

	for i := workflow.LastTaskIndexCompleted + 1; i < workflow.GetLen(); i++ {
		task := workflow.Tasks[i]

		if task.IsServer {
			break
		}

		remoteTaskNames = append(remoteTaskNames, task.Name)
	}

	if len(remoteTaskNames) == 0 {
		return false, nil
	}

	errSend := workflow.srvMessenger.SendRemoteTasksToRun(remoteTaskNames)

	if errSend != nil {
		return false, errSend
	}

	return true, nil
}

// CopyRemoteTasksProgress saves remote tasks feedback in the server workflow
// and eventually to the data store. Tt loops through received remote tasks and
// copies progress to server workflow till end of remote tasks length -- or
// first unsuccessful remote task.
func (workflow *Workflow) CopyRemoteTasksProgress(remoteMsg *wrpc.RemoteMsg) error {
	if remoteMsg.Tasks == nil || len(remoteMsg.Tasks) == 0 {
		return errors.New("error nil cli tasks")
	}

	for remoteTaskIdx := 0; workflow.LastTaskIndexCompleted+1 < len(workflow.Tasks) &&
		remoteTaskIdx < len(remoteMsg.Tasks); remoteTaskIdx++ {
		nextWorkflowTask := workflow.Tasks[workflow.LastTaskIndexCompleted+1]
		nextRemoteTask := remoteMsg.Tasks[remoteTaskIdx]

		if nextWorkflowTask.IsServer || !strings.EqualFold(nextWorkflowTask.Name, nextRemoteTask.TaskName) {
			return fmt.Errorf("error mismatch between next Remote task %s, Workflow task %s",
				nextRemoteTask.TaskName, nextWorkflowTask.Name)
		}

		nextWorkflowTask.Log = remoteMsg.Tasks[remoteTaskIdx].ErrorMsg
		nextWorkflowTask.Completed = remoteMsg.Tasks[remoteTaskIdx].Completed

		if nextWorkflowTask.Log != "" || !nextWorkflowTask.Completed {
			log.Println("Remote task:", nextWorkflowTask.Name, "failed :", nextWorkflowTask.Log)
			return errors.New(nextWorkflowTask.Log)
		}

		nextWorkflowTask.CompletedAt = time.Now()
		workflow.LastTaskNameCompleted = nextRemoteTask.TaskName
		workflow.LastTaskIndexCompleted++

		if nextRemoteTask.Completed {
			log.Println("Remote task:", strings.TrimSpace(nextWorkflowTask.Name), "completed.")
		}
	}

	return nil
}

// SetWorkflowCompletedChecked checks if the lastTaskIndexCompleted has progressed past
// all the completed tasks and returns T/F whether the workflow has completed.
func (workflow *Workflow) SetWorkflowCompletedChecked(ctx context.Context) bool {
	if workflow.Completed {
		return true
	}

	// Zero tasks to do? Sounds like a unit test :)
	if workflow.GetLen() == 0 {
		workflow.CompletedAt = time.Now()
		workflow.Completed = true

		return true
	}

	if workflow.LastTaskIndexCompleted+1 == workflow.GetLen() {
		workflow.CompletedAt = time.Now()
		workflow.Completed = true

		return true
	}

	return false
}

// GetTaskName gets the string function name for the current goroutine
// executing func in stack frame (-1) trimming path/package characters up to
// function name accounting that for the first 3 characters to discard "New"
// i.e., NewCalcX becomes CalcX. Ref:
// stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
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

func (workflow *Workflow) PrintTasks() {
	for _, task := range workflow.Tasks {
		task.Print()
	}
}

// Display the workflow state in color ascii.
func (workflow *Workflow) Display() {
	var style = simpletable.StyleRounded

	tree := simpletable.New()
	tree.SetStyle(style)

	fmt.Println()
	for idx, task := range workflow.Tasks {

		tree.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Text: "Task " + strconv.Itoa(idx+1)},
				{Align: simpletable.AlignCenter, Text: "Workflow " + ascii.Blue(workflow.Name)},
				{Align: simpletable.AlignCenter, Text: "Task Type"},
			},
		}

		tree.Body = &simpletable.Body{
			Cells: [][]*simpletable.Cell{
				{
					&simpletable.Cell{Text: "Name: " + ascii.Blue(task.Name)},
					&simpletable.Cell{Text: task.DisplayStatus()},
					&simpletable.Cell{Text: task.DisplayType()},
				},
			},
		}

		tree.Footer = &simpletable.Footer{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignCenter, Span: 3, Text: ascii.Red(task.Log)},
			},
		}

		tree.Println()
	}
}
