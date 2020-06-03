package remote

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
	"github.com/tradeline-tech/workflow/types"

	"github.com/tradeline-tech/workflow/wrpc"
)

const ConfigRemoteMessengerKey = "remoteMessenger"

type Remote struct {
	ctx                  context.Context
	gRPCRemote           wrpc.TaskCommunicatorClient
	messenger            *wrpc.RemoteMessenger
	cfg                  types.TaskConfiguration
	workflowNameKeyValue string
	remoteTasksMap       RemoteTaskRunnersByKey
}

type RemoteTaskRunnersByKey map[string]types.TaskRunnerNewFunc

func New(
	ctx context.Context,
	gRPCRemote wrpc.TaskCommunicatorClient,
	cfg types.TaskConfiguration,
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
	var messenger types.MsgToSrv

	defer recoverFromPanic()

	gRPCRemoteConnToSrv, err := r.gRPCRemote.RunWorkflow(r.ctx)
	messenger = r.getRemoteMessenger(gRPCRemoteConnToSrv)

	// Server would detect this at its end as an I/O error
	defer endRemoteToSrvConnection(gRPCRemoteConnToSrv)

	if err != nil {
		// TODO .Error(ctx, err, "upon opening a streaming connection to the server")

		return err
	}

	if err = messenger.SendWorkflowNameKeyToSrv(r.workflowNameKeyValue); err != nil {
		return err
	}

	for {
		// 1. Get server msg
		serverMsg, errIn := gRPCRemoteConnToSrv.Recv()
		if errIn == io.EOF {
			break
		}

		// 2. Exit if any server error
		if err = handleSrvMsgErr(serverMsg, errIn); err != nil {
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
		err = r.runRemoteWorkflow(messenger, remoteTasks)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Remote) getRemoteMessenger(gRPCRemoteConnToSrv wrpc.TaskCommunicator_RunWorkflowClient) types.MsgToSrv {
	messenger := wrpc.NewRemoteMessenger(gRPCRemoteConnToSrv)
	r.cfg.Add(ConfigRemoteMessengerKey, messenger)
	return messenger
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		// TODO logger.Debug(context.Background(), "recover:", r)
	}
}

func endRemoteToSrvConnection(stream wrpc.TaskCommunicator_RunWorkflowClient) {
	err := stream.CloseSend()
	if err != nil {
		fmt.Println(err, "failed to close remote connection to the server")
	}
}

func handleSrvMsgErr(serverMsg *wrpc.ServerMsg, err error) error {
	if err != nil {
		// TODO logger.Error(ctx, err, " Failed to receive server messages: %v", err)

		return err
	}

	if serverMsg == nil {
		nilMsgError := errors.New("received nil server message")
		log.Println("error : ", nilMsgError, ", received nil server message")

		return nilMsgError
	}

	serverErrMsg := serverMsg.ErrorMsg
	if serverErrMsg != "" {
		serverErr := errors.New(serverErrMsg)
		log.Println("error : ", serverErr, ", received server error")

		return serverErr
	}

	return nil
}

// Runs received tasks from the server gRPC connection.
// Asserts that each received task name is located in the declared tasks folder
// otherwise it exits prematurely with error out and notifies the server.
func (r *Remote) runRemoteWorkflow(messenger types.MsgToSrv, remoteTaskNames []string) error {
	if len(remoteTaskNames) > 0 {

		taskRunners, err := r.getTaskRunnersToRun(messenger, remoteTaskNames)
		if err != nil {
			return err
		}

		if len(taskRunners) > 0 {
			tasksToRun, err := types.NewWorkflow(r.cfg, nil, r.workflowNameKeyValue, taskRunners)
			if err != nil {
				// TODO log error
				return err
			}
			if err := r.runReceivedTasks(messenger, tasksToRun); err != nil {
				// TODO .Error(ctx, err, "while running Remote tasks")

				return err
			}
		}
	}

	return nil
}

// getTaskRunnersToRun maps the gRPC received task names to run and returns
// them upon a task mismatch, it notifies the server and errs.
func (r *Remote) getTaskRunnersToRun(messenger types.MsgToSrv, remoteTaskNames []string) (types.TaskRunners, error) {
	taskRunners := make(types.TaskRunners, 0, len(remoteTaskNames))

	for idx, remoteTaskName := range remoteTaskNames {
		taskRunner, ok := r.remoteTasksMap[remoteTaskName]
		if !ok {
			if err := r.chkTaskExists(messenger, idx, remoteTaskName); err != nil {
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
func (r *Remote) chkTaskExists(messenger types.MsgToSrv, taskIdx int, remoteTaskName string) error {
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
	errWrapped := messenger.SendTasksErrorMsgToServer(
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
func (r *Remote) runReceivedTasks(messenger types.MsgToSrv, remoteTasks *types.Tasks) error {
	remoteMsgTasks := make([]*wrpc.RemoteMsg_Tasks, 0, len(remoteTasks.TaskRunners))

	for idx, taskRunner := range remoteTasks.TaskRunners {
		msg := fmt.Sprintf("Initiating task %s...", taskRunner.GetTask().Name)

		fmt.Println(msg)

		if err := messenger.SendTaskStatusToServer(r.workflowNameKeyValue, msg); err != nil {
			log.Println("error : ", err, "failed to send task status message")
		}

		if err := types.ValidDo(taskRunner); err != nil {
			remoteMsgTasks = append(remoteMsgTasks, &wrpc.RemoteMsg_Tasks{
				TaskName:  remoteTasks.Tasks[idx].Name,
				ErrorMsg:  err.Error(),
				Completed: false,
			})
			errWrapped := messenger.SendTasksErrorMsgToServer(
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

	if err := messenger.SendTaskCompletionToServer(r.workflowNameKeyValue, remoteMsgTasks); err != nil {
		return err
	}

	return nil
}

// SendDatumToRemote send a string data element to the Server.
func SendDatumToServer(workflowNameKey string, datum string, config types.TaskConfiguration) error {

	remoteMessengerVal, found := config.Get(ConfigRemoteMessengerKey)
	if !found {
		return errors.New("messenger not found in configuration")
	}

	messenger, typeOk := remoteMessengerVal.(types.MsgToSrv)
	if !typeOk {
		return errors.New("invalid messenger")
	}

	return messenger.SendDatumToServer(workflowNameKey, datum)
}

// SendDataToRemote send a string slice of data elements to the Server.
func SendDataToServer(workflowNameKey string, data []string, config types.TaskConfiguration) error {

	remoteMessengerVal, found := config.Get(ConfigRemoteMessengerKey)
	if !found {
		return errors.New("messenger not found in configuration")
	}

	messenger, typeOk := remoteMessengerVal.(types.MsgToSrv)
	if !typeOk {
		return errors.New("invalid messenger")
	}

	return messenger.SendDataToServer(workflowNameKey, data)
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
	port int, remoteTaskRunners types.TaskRunners) error {
	ctx := context.Background()

	client, err := connectToServerWithoutTLS(ctx, serverAddress, port)
	if err != nil {
		return err
	}

	cfg := config.NewTasksBoostrapConfig()
	cfg.Add(types.WorkflowKey, "dh-secret")

	remote := New(
		ctx,
		client,
		cfg,
		"dh-secret",
		copyTaskRunnersToMap(remoteTaskRunners))

	return remote.ProcessGRPCMessages()
}

func copyTaskRunnersToMap(runners types.TaskRunners) RemoteTaskRunnersByKey {
	runnersMap := make(RemoteTaskRunnersByKey)
	for _, runner := range runners {
		runnerType := strings.ToLower(strings.Split(fmt.Sprintf("%T", runner(nil)), ".")[1])
		runnersMap[runnerType] = runner
	}

	return runnersMap
}
