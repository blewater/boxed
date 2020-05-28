package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/tradeline-tech/workflow/common"
	"github.com/tradeline-tech/workflow/pkg/config"

	"github.com/tradeline-tech/argo/pkg/logger"

	"github.com/tradeline-tech/workflow/grpc"
)

var (
	cfg config.TaskConfiguration
)

// ProcessGRPCMessages Enters the remote messaging processing loop
// nolint:funlen
func ProcessGRPCMessages(ctx context.Context,
	remoteTaskRunners common.RemoteTaskRunnersByKey,
	workflowNameKey string,
	gRpcRemote grpc.TaskCommunicatorClient) error {

	defer recoverFromPanic()

	stream, err := gRpcRemote.RunWorkflow(ctx)

	// Server would detect this from its end in the form of an I/O error
	defer endRemoteToSrvConnection(ctx, stream)

	if err != nil {
		logger.Error(ctx, err, "upon opening a streaming connection to the server")

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
		logger.Debug(context.Background(), "recover:", r)
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
		logger.Error(ctx, err, " Failed to receive server messages: %v", err)

		return err
	}

	if serverMsg == nil {
		nilMsgError := errors.New("received nil server message")
		logger.Error(ctx, nilMsgError, "exiting")

		return nilMsgError
	}

	serverErrMsg := serverMsg.ErrorMsg
	if serverErrMsg != "" {
		serverErr := errors.New(serverErrMsg)
		logger.Error(ctx, serverErr, "received server error")

		return serverErr
	}

	return nil
}

func runRemoteWorkflow(ctx context.Context,
	remoteTaskRunners common.RemoteTaskRunnersByKey,
	workflowNameKey string,
	remoteConfig config.TaskConfiguration,
	remoteTaskNames []string,
	stream grpc.TaskCommunicator_RunWorkflowClient) error {
	var remoteentTaskRunners common.TaskRunners

	if len(remoteTaskNames) > 0 {

		remoteentTaskRunners = make(common.TaskRunners, 0, len(remoteTaskNames))

		for idx, remoteTaskName := range remoteTaskNames {
			taskRunner, ok := remoteTaskRunners[remoteTaskName]
			if !ok {
				if err := chkCliTaskExists(ctx, workflowNameKey, idx, remoteTaskRunners, remoteTaskName, stream); err != nil {
					return err
				}
			}

			if ok {
				remoteentTaskRunners = append(remoteentTaskRunners, taskRunner)
			}
		}

		if len(remoteentTaskRunners) > 0 {
			if remoteConfig == nil {
				errNilConfig := errors.New("remote configuration is nil. No Server configuration was sent")
				logger.Error(ctx, errNilConfig)

				return errNilConfig
			}

			remoteentWorkflow, err := common.NewWorkflow(remoteConfig, workflowNameKey, remoteentTaskRunners)
			if err != nil {
				// TODO log error
				return err
			}
			if err := runReceivedTasks(ctx, workflowNameKey, remoteentWorkflow, stream); err != nil {
				logger.Error(ctx, err, "while running Cli tasks")

				return err
			}
		}
	}

	return nil
}

// Checks whether it receives an unknown task name from the server to run here on Cli
// and couldn't find it.
func chkCliTaskExists(ctx context.Context,
	token string,
	idx int,
	remoteTaskRunners common.RemoteTaskRunnersByKey,
	remoteTaskName string, stream grpc.TaskCommunicator_RunWorkflowClient) error {
	errMsg := fmt.Sprintf(
		"received from server an unknown remote task name: %s. Cli task names: %v", remoteTaskName,
		reflect.ValueOf(remoteTaskRunners).MapKeys())
	grpcCliMsgTasks := []*grpc.RemoteMsg_Tasks{
		{
			TaskName:  remoteTaskName,
			ErrorMsg:  errMsg,
			Completed: false,
		},
	}
	errUnmatchedCliTask := errors.New(errMsg)

	errWrapped := grpc.SendTasksErrorMsgToServer(ctx, stream, token, idx, remoteTaskName, grpcCliMsgTasks,
		errUnmatchedCliTask)
	if errWrapped != nil {
		return errWrapped
	}

	return errUnmatchedCliTask
}

func runReceivedTasks(
	ctx context.Context,
	token string,
	workflow *common.Tasks,
	stream grpc.TaskCommunicator_RunWorkflowClient) error {
	remoteTasks := make([]*grpc.RemoteMsg_Tasks, 0, len(workflow.TaskRunners))

	for idx, taskRunner := range workflow.TaskRunners {
		msg := fmt.Sprintf("Initiating task %s...", taskRunner.GetTask().Name)

		fmt.Println(msg)

		if err := grpc.SendMsgToServer(ctx, stream, token, msg); err != nil {
			logger.Error(ctx, err, "failed to send task status message")
		}

		if err := common.ValidDo(taskRunner); err != nil {
			remoteTasks = append(remoteTasks, &grpc.RemoteMsg_Tasks{
				TaskName:  workflow.Tasks[idx].Name,
				ErrorMsg:  err.Error(),
				Completed: false,
			})
			errWrapped := grpc.SendTasksErrorMsgToServer(ctx, stream, token, idx, workflow.Tasks[idx].Name, remoteTasks, err)
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

	if err := grpc.SendTaskCompletionToServer(ctx, stream, token, remoteTasks); err != nil {
		return err
	}

	return nil
}
