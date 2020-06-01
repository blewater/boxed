package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"google.golang.org/grpc"

	"github.com/tradeline-tech/workflow/pkg/config"
	. "github.com/tradeline-tech/workflow/types"

	"github.com/tradeline-tech/workflow/wrpc"
)

type Remote struct {
	ctx                  context.Context
	gRPCRemote           wrpc.TaskCommunicatorClient
	cfg                  config.TaskConfiguration
	workflowNameKeyValue string
	remoteTasksMap       RemoteTaskRunnersByKey
}

type RemoteTaskRunnersByKey map[string]TaskRunnerNewFunc

func New(
	ctx context.Context,
	gRPCRemote wrpc.TaskCommunicatorClient,
	cfg config.TaskConfiguration,
	userWorkflowNameKeyValue string,
	declaredRemoteTasksMap RemoteTaskRunnersByKey) *Remote {

	return &Remote{
		ctx:                  ctx,
		gRPCRemote:           gRPCRemote,
		cfg:                  cfg,
		workflowNameKeyValue: userWorkflowNameKeyValue,
		remoteTasksMap:       declaredRemoteTasksMap,
	}
}

// ProcessGRPCMessages Enters the remote messaging processing loop
func (r *Remote) ProcessGRPCMessages() error {

	defer recoverFromPanic()

	gRPCRemoteConnToSrv, err := r.gRPCRemote.RunWorkflow(r.ctx)

	// Server would detect this at its end as an I/O error
	defer endRemoteToSrvConnection(r.ctx, gRPCRemoteConnToSrv)

	if err != nil {
		// TODO .Error(ctx, err, "upon opening a streaming connection to the server")

		return err
	}

	if err = wrpc.SendWorkflowNameKeyToSrv(gRPCRemoteConnToSrv, r.workflowNameKeyValue); err != nil {
		return err
	}

	for {
		// 1. Get server msg
		serverMsg, errIn := gRPCRemoteConnToSrv.Recv()
		if errIn == io.EOF {
			break
		}

		// 2. Exit if any server error
		if err = handleSrvMsgErr(r.ctx, serverMsg, errIn); err != nil {
			return err
		}

		// 3. Print any sent server output
		if serverMsg.TaskOutput != "" {
			fmt.Println(serverMsg.TaskOutput)
		}

		// 4. Are we done?
		if serverMsg.TaskInProgress == wrpc.ServerWorkflowCompletionText {
			fmt.Println("Received workflow completion message. Exiting...")
			break
		}

		// 5. Run any Remote-side tasks
		remoteTasks := serverMsg.RemoteTasks
		err = r.runRemoteWorkflow(gRPCRemoteConnToSrv, remoteTasks)
		if err != nil {
			return err
		}
	}

	return nil
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		// TODO logger.Debug(context.Background(), "recover:", r)
	}
}

func endRemoteToSrvConnection(ctx context.Context, stream wrpc.TaskCommunicator_RunWorkflowClient) {
	err := stream.CloseSend()
	if err != nil {
		fmt.Println(err, "failed to close remote connection to the server")
	}
}

func handleSrvMsgErr(ctx context.Context, serverMsg *wrpc.ServerMsg, err error) error {
	if err != nil {
		// TODO logger.Error(ctx, err, " Failed to receive server messages: %v", err)

		return err
	}

	if serverMsg == nil {
		nilMsgError := errors.New("received nil server message")
		// TODO logger.Error(ctx, nilMsgError, "exiting")

		return nilMsgError
	}

	serverErrMsg := serverMsg.ErrorMsg
	if serverErrMsg != "" {
		serverErr := errors.New(serverErrMsg)
		// TODO logger.Error(ctx, serverErr, "received server error")

		return serverErr
	}

	return nil
}

// Runs received tasks from the server gRPC connection.
// Asserts that each received task name is located in the declared tasks folder
// otherwise it exits prematurely with error out and notifies the server.
func (r *Remote) runRemoteWorkflow(gRPCConn wrpc.TaskCommunicator_RunWorkflowClient, remoteTaskNames []string) error {
	if len(remoteTaskNames) > 0 {

		taskRunners, err := r.getTaskRunnersToRun(gRPCConn, remoteTaskNames)
		if err != nil {
			return err
		}

		if len(taskRunners) > 0 {
			tasksToRun, err := NewWorkflow(r.cfg, r.workflowNameKeyValue, taskRunners)
			if err != nil {
				// TODO log error
				return err
			}
			if err := r.runReceivedTasks(gRPCConn, tasksToRun); err != nil {
				// TODO .Error(ctx, err, "while running Remote tasks")

				return err
			}
		}
	}

	return nil
}

// getTaskRunnersToRun maps the gRPC received task names to run and returns
// them upon a task mismatch, it notifies the server and errs.
func (r *Remote) getTaskRunnersToRun(
	gRPCConn wrpc.TaskCommunicator_RunWorkflowClient, remoteTaskNames []string) (TaskRunners, error) {
	taskRunners := make(TaskRunners, 0, len(remoteTaskNames))

	for idx, remoteTaskName := range remoteTaskNames {
		taskRunner, ok := r.remoteTasksMap[remoteTaskName]
		if !ok {
			if err := r.chkTaskExists(gRPCConn, idx, remoteTaskName); err != nil {
				return nil, err
			}
		}

		if ok {
			taskRunners = append(taskRunners, taskRunner)
		}
	}

	return taskRunners, nil
}

// Checks whether it receives an unknown task name from the server to run.
// Notifies server if sent task is not found.
func (r *Remote) chkTaskExists(
	gRPCConn wrpc.TaskCommunicator_RunWorkflowClient, taskIdx int, remoteTaskName string) error {
	errMsg := fmt.Sprintf(
		"received from server an unknown remote task name: %s. Remote task names: %v", remoteTaskName,
		reflect.ValueOf(r.remoteTasksMap).MapKeys())

	gRpcRemoteMsgTasks := []*wrpc.RemoteMsg_Tasks{
		{
			TaskName:  remoteTaskName,
			ErrorMsg:  errMsg,
			Completed: false,
		},
	}

	errUnmatchedTask := errors.New(errMsg)
	errWrapped := wrpc.SendTasksErrorMsgToServer(
		r.ctx,
		gRPCConn,
		r.workflowNameKeyValue,
		taskIdx,
		remoteTaskName,
		gRpcRemoteMsgTasks,
		errUnmatchedTask)

	if errWrapped != nil {
		return errWrapped
	}

	return errUnmatchedTask
}

// runReceivedTasks runs the requested remoteTasks and sends the combined
// success status or number of successful tasks + last failed task with error.
func (r *Remote) runReceivedTasks(gRPCConn wrpc.TaskCommunicator_RunWorkflowClient, remoteTasks *Tasks) error {
	remoteMsgTasks := make([]*wrpc.RemoteMsg_Tasks, 0, len(remoteTasks.TaskRunners))

	for idx, taskRunner := range remoteTasks.TaskRunners {
		msg := fmt.Sprintf("Initiating task %s...", taskRunner.GetTask().Name)

		fmt.Println(msg)

		if err := wrpc.SendMsgToServer(r.ctx, gRPCConn, r.workflowNameKeyValue, msg); err != nil {
			// TODO .Error(ctx, err, "failed to send task status message")
		}

		if err := ValidDo(taskRunner); err != nil {
			remoteMsgTasks = append(remoteMsgTasks, &wrpc.RemoteMsg_Tasks{
				TaskName:  remoteTasks.Tasks[idx].Name,
				ErrorMsg:  err.Error(),
				Completed: false,
			})
			errWrapped := wrpc.SendTasksErrorMsgToServer(
				r.ctx,
				gRPCConn,
				r.workflowNameKeyValue,
				idx, remoteTasks.Tasks[idx].Name,
				remoteMsgTasks,
				err)
			if errWrapped != nil {
				return errWrapped
			}

			return err
		}

		remoteMsgTasks = append(remoteMsgTasks, &wrpc.RemoteMsg_Tasks{
			TaskName:  taskRunner.GetTask().Name,
			Completed: true,
		})
	}

	if err := wrpc.SendTaskCompletionToServer(
		r.ctx, gRPCConn, r.workflowNameKeyValue, remoteMsgTasks); err != nil {
		return err
	}

	return nil
}

// connectToServerWithoutTLS connects to the workflow server without public
// asymmetric encryption (TLS) to the provided address and port.
func connectToServerWithoutTLS(
	ctx context.Context, serverAddress string, port int) (wrpc.TaskCommunicatorClient, error) {
	clientConn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d",
		serverAddress, port),
		grpc.WithInsecure(),
		grpc.WithBlock())

	if err != nil {
		log.Println(err)
	}

	gRPCRemote := wrpc.NewTaskCommunicatorClient(clientConn)

	return gRPCRemote, err
}

// StartWorkflow connects to the server and runs the declared workflow.
func StartWorkflow(
	serverAddress string,
	port int, remoteTaskRunners TaskRunners) error {
	ctx := context.Background()

	client, err := connectToServerWithoutTLS(ctx, serverAddress, port)
	if err != nil {
		return err
	}

	cfg := config.NewTasksBoostrapConfig()
	cfg.Add(WorkflowKey, "dh-secret")

	remote := New(
		ctx,
		client,
		cfg,
		"dh-secret",
		copyTaskRunnersToMap(remoteTaskRunners))

	return remote.ProcessGRPCMessages()
}

func copyTaskRunnersToMap(runners TaskRunners) RemoteTaskRunnersByKey {
	runnersMap := make(RemoteTaskRunnersByKey)
	for _, runner := range runners {
		runnerType := strings.ToLower(strings.Split(fmt.Sprintf("%T", runner(nil)), ".")[1])
		runnersMap[runnerType] = runner
	}

	return runnersMap
}
