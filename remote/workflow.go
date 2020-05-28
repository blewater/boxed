package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"

	"github.com/tradeline-tech/workflow/grpc"
)

var (
	cfg config.TaskConfiguration
)

// ProcessGRPCMessages Enters the remote messaging processing loop
// nolint:funlen
func ProcessGRPCMessages(ctx context.Context,
	remoteTaskRunners types.RemoteTaskRunnersByKey,
	workflowNameKey string,
	gRpcRemote grpc.TaskCommunicatorClient) error {

	defer recoverFromPanic()

	stream, err := gRpcRemote.RunWorkflow(ctx)

	// Server would detect this from its end in the form of an I/O error
	defer endRemoteToSrvConnection(ctx, stream)

	if err != nil {
		// TODO .Error(ctx, err, "upon opening a streaming connection to the server")

		return err
	}

	if err = grpc.SendWorkflowNameKeyToSrv(ctx, stream, workflowNameKey); err != nil {
		return err
	}

	for {
		// 1. Get server msg
		serverMsg, errIn := stream.Recv()
		if errIn == io.EOF {
			break
		}

		// 2. Exit if any server error
		if err = handleMsgErr(ctx, errIn, serverMsg); err != nil {
			return err
		}

		// 3. Print any sent server output
		if serverMsg.TaskOutput != "" {
			fmt.Println(serverMsg.TaskOutput)
		}

		// 4. Are we done?
		if serverMsg.TaskInProgress == grpc.ServerWorkflowCompletionText {
			fmt.Println("Received workflow completion message. Exiting...")
			break
		}

		// 7. Run any Cli-side tasks
		remoteTasks := serverMsg.RemoteTasks
		err = runRemoteWorkflow(ctx, remoteTaskRunners, workflowNameKey, cfg, remoteTasks, stream)
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

func endRemoteToSrvConnection(ctx context.Context, stream grpc.TaskCommunicator_RunWorkflowClient) {
	err := stream.CloseSend()
	if err != nil {
		fmt.Println(err, "failed to close remote connection to the server")
	}
}

func handleMsgErr(ctx context.Context, err error, serverMsg *grpc.ServerMsg) error {
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
// Asserts that each received task name is located in the declared tasks folder.
// Otherwise, it exits prematurely with error out and notifies the server.
func runRemoteWorkflow(ctx context.Context,
	remoteTaskRunnersByKey types.RemoteTaskRunnersByKey,
	workflowNameKey string,
	remoteConfig config.TaskConfiguration,
	remoteTaskNames []string,
	stream grpc.TaskCommunicator_RunWorkflowClient) error {
	if len(remoteTaskNames) > 0 {

		taskRunners, err := getTaskRunnersToRun(ctx, remoteTaskRunnersByKey, workflowNameKey, remoteTaskNames, stream)
		if err != nil {
			return err
		}

		if len(taskRunners) > 0 {
			if remoteConfig == nil {
				errNilConfig := errors.New("remote configuration is nil. No Server configuration was sent")
				// TODO .Error(ctx, errNilConfig)

				return errNilConfig
			}

			remoteTasks, err := types.NewWorkflow(remoteConfig, workflowNameKey, taskRunners)
			if err != nil {
				// TODO log error
				return err
			}
			if err := runReceivedTasks(ctx, workflowNameKey, remoteTasks, stream); err != nil {
				// TODO .Error(ctx, err, "while running Cli tasks")

				return err
			}
		}
	}

	return nil
}

// getTaskRunnersToRun maps the gRPC received task names to run and returns them.
// upon a task mismatch, it notifies the server and errs.
func getTaskRunnersToRun(ctx context.Context, remoteTaskRunnersMap types.RemoteTaskRunnersByKey,
	workflowNameKey string, remoteTaskNames []string,
	gRpcConn grpc.TaskCommunicator_RunWorkflowClient) (types.TaskRunners, error) {
	taskRunners := make(types.TaskRunners, 0, len(remoteTaskNames))

	for idx, remoteTaskName := range remoteTaskNames {
		taskRunner, ok := remoteTaskRunnersMap[remoteTaskName]
		if !ok {
			if err := chkTaskExists(ctx, workflowNameKey, idx, remoteTaskRunnersMap, remoteTaskName, gRpcConn); err != nil {
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
func chkTaskExists(ctx context.Context,
	workflowNameKey string,
	taskIdx int,
	remoteTaskRunners types.RemoteTaskRunnersByKey,
	remoteTaskName string, stream grpc.TaskCommunicator_RunWorkflowClient) error {
	errMsg := fmt.Sprintf(
		"received from server an unknown remote task name: %s. Cli task names: %v", remoteTaskName,
		reflect.ValueOf(remoteTaskRunners).MapKeys())

	gRpcRemoteMsgTasks := []*grpc.RemoteMsg_Tasks{
		{
			TaskName:  remoteTaskName,
			ErrorMsg:  errMsg,
			Completed: false,
		},
	}

	errUnmatchedTask := errors.New(errMsg)
	errWrapped := grpc.SendTasksErrorMsgToServer(ctx,
		stream,
		workflowNameKey,
		taskIdx,
		remoteTaskName,
		gRpcRemoteMsgTasks,
		errUnmatchedTask)

	if errWrapped != nil {
		return errWrapped
	}

	return errUnmatchedTask
}

func runReceivedTasks(
	ctx context.Context,
	workflowNameKey string,
	workflow *types.Tasks,
	stream grpc.TaskCommunicator_RunWorkflowClient) error {
	remoteTasks := make([]*grpc.RemoteMsg_Tasks, 0, len(workflow.TaskRunners))

	for idx, taskRunner := range workflow.TaskRunners {
		msg := fmt.Sprintf("Initiating task %s...", taskRunner.GetTask().Name)

		fmt.Println(msg)

		if err := grpc.SendMsgToServer(ctx, stream, workflowNameKey, msg); err != nil {
			// TODO .Error(ctx, err, "failed to send task status message")
		}

		if err := types.ValidDo(taskRunner); err != nil {
			remoteTasks = append(remoteTasks, &grpc.RemoteMsg_Tasks{
				TaskName:  workflow.Tasks[idx].Name,
				ErrorMsg:  err.Error(),
				Completed: false,
			})
			errWrapped := grpc.SendTasksErrorMsgToServer(ctx, stream, workflowNameKey, idx, workflow.Tasks[idx].Name, remoteTasks, err)
			if errWrapped != nil {
				return errWrapped
			}

			return err
		}

		remoteTasks = append(remoteTasks, &grpc.RemoteMsg_Tasks{
			TaskName:  taskRunner.GetTask().Name,
			Completed: true,
		})
	}

	if err := grpc.SendTaskCompletionToServer(ctx, stream, workflowNameKey, remoteTasks); err != nil {
		return err
	}

	return nil
}
