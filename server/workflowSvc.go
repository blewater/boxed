package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"

	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

type WorkflowsType map[string]*types.Tasks
type WorkflowsConfigType map[string]types.TaskConfiguration

// SrvTaskRunners is the list of declared constructor tasks
// to run for the workflow tasks mapping one runner to one task
// they run in the sequence listed here.
type SrvTaskRunners []types.TaskRunnerNewFunc

// Server Workflow Loop Step Algorithm Actions
type reqActionType int

const (
	// Continue processing. Non-disrupting action effect.
	continueProcessing reqActionType = iota
	// Stop processing this request.
	haltRequestAction
	// Go back to listening for remote messages
	// typically this is desirable after sending
	// tasks for remote execution.
	pauseServerProcessing
	ConfigServerMessengerKey = "serverMessenger"
)

type WorkflowsServer struct {
	// gRPC
	wrpc.UnimplementedTaskCommunicatorServer
	gRPCServer wrpc.TaskCommunicator_RunWorkflowServer

	// Whether to run a single workflow and exit upon error or completion
	soloWorkflowMode bool
	srvTaskRunners   SrvTaskRunners

	// Serialize access to concurrent workflows config access
	mu sync.Mutex

	// Cached org workflows, config
	workflows       WorkflowsType
	workflowsConfig WorkflowsConfigType
	TaskRunners     []types.TaskRunnerNewFunc
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		log.Println("recover: ", r)
	}
}

// NewWorkflowsServer initializes a workflow server for servicing
// a business group of with an instance of a workflow.
// Multiple workflows with the same declared tasks could run concurrently
// by one Workflow server.
func NewWorkflowsServer(soloWorkflow bool, srvTaskRunners SrvTaskRunners) *WorkflowsServer {
	return &WorkflowsServer{
		soloWorkflowMode: soloWorkflow,
		srvTaskRunners:   srvTaskRunners,
		workflows:        make(WorkflowsType),
		// Link the declared task runners above with the workflow server
		TaskRunners:     srvTaskRunners,
		workflowsConfig: make(WorkflowsConfigType),
	}
}

// RunTasks opens a bidirectional stream to a remote client and
// runs tasks with workflow name-key received from the remote gRPC client.
func (srv *WorkflowsServer) RunWorkflow(gRPCConnToRemote wrpc.TaskCommunicator_RunWorkflowServer) error {
	defer recoverFromPanic()
	var req = WorkflowServerReq{WorkflowServer: srv}
	req.setServerMessenger(gRPCConnToRemote)

	ctx := gRPCConnToRemote.Context()

	defer req.safeSaveWorkflow()

	for {
		clientMsg, err := gRPCConnToRemote.Recv()

		if stepCheckIOError(ctx, clientMsg, err) {
			break
		}

		if loopAction := req.stepProcessReceivedMessages(ctx, clientMsg); loopAction == haltRequestAction {
			break
		} else if loopAction == pauseServerProcessing {
			continue
		}

		if req.stepIsWorkflowCompletedRemotely(ctx, clientMsg) {
			break
		}

		if err = req.stepRunServerSideTasks(ctx, clientMsg); err != nil {
			break
		}

		if loopAction := req.stepSendRemoteTasks(); loopAction == haltRequestAction {
			break
		} else if loopAction == pauseServerProcessing {
			continue
		}

		req.stepWorkflowCompleted(ctx)
	}

	return nil
}

// 1. Check whether the protocol errs
func stepCheckIOError(ctx context.Context, remoteMsg *wrpc.RemoteMsg, err error) (exit bool) {
	if err == io.EOF {
		log.Println("remote severed the connection to the server, EOF")

		return true
	}

	if err != nil {
		log.Println("error : ", err, ", server message reception error. Is the remote client still up?")

		return true
	}

	if errClient := ctx.Err(); errClient != nil {
		log.Println("server reception error; Is the remote client still up?")

		return true
	}

	if remoteMsg == nil {
		log.Println("error : ", err, ", "+
			"received nil remote message. Remote likely closed the connection.")

		return true
	}

	return false
}

func (srv *WorkflowsServer) getMemCachedWorkflow(workflowNameKey string) (*types.Tasks, types.TaskConfiguration) {
	if workflowNameKey == "" {
		return nil, nil
	}

	srv.mu.Lock()
	// attempt to get cached workflow
	retWorkflow, _ := srv.workflows[workflowNameKey]
	retWorkflowConfig, _ := srv.workflowsConfig[workflowNameKey]
	srv.mu.Unlock()

	return retWorkflow, retWorkflowConfig
}

func (srv *WorkflowsServer) setMemCachedWorkflow(
	workflowNameKey string, workflow *types.Tasks, cfg types.TaskConfiguration) {
	if workflowNameKey == "" {
		return
	}

	srv.mu.Lock()
	srv.workflows[workflowNameKey] = workflow
	srv.workflowsConfig[workflowNameKey] = cfg
	srv.mu.Unlock()
}

func readWorkflowFromDB(ctx context.Context, workflowNameKey string) (*types.Tasks, error) {
	readDocument, err := datastore.LoadByName(ctx, workflowNameKey)
	if err != nil {
		return nil, err
	}

	// No workflow found or db connection not used
	if readDocument == nil {
		return nil, nil
	}

	// type assert that the read document is a tasks type collection
	foundWorkflow, okWorkflow := readDocument.(*types.Tasks)
	if !okWorkflow {
		return nil, errors.New("failed to convert read document to a workflow")
	}
	return foundWorkflow, nil
}

// SendDatumToRemote send a string data element to the Server.
func SendDatumToRemote(workflowNameKey string, datum string, config types.TaskConfiguration) error {

	remoteMessengerVal, found := config.Get(ConfigServerMessengerKey)
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
func SendDataToRemote(data []string, config types.TaskConfiguration) error {

	remoteMessengerVal, found := config.Get(ConfigServerMessengerKey)
	if !found {
		return errors.New("messenger not found in configuration")
	}

	messenger, typeOk := remoteMessengerVal.(types.MsgToRemote)
	if !typeOk {
		return errors.New("invalid messenger")
	}

	return messenger.SendDataToRemote(data)
}

func StartUp(
	soloWorkflow bool,
	srvAddress string,
	srvPort int,
	serverTaskRunners SrvTaskRunners) (gRpcServer *grpc.Server, port int, err error) {
	srvAddressPort := fmt.Sprintf("%s:%d", srvAddress, srvPort)
	tcpListener, err := net.Listen("tcp", srvAddressPort)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen at tcp %s", srvAddressPort)
	}
	log.Println("listening at ", srvAddressPort)
	gRpcServer = grpc.NewServer()

	workflowServer := NewWorkflowsServer(soloWorkflow, serverTaskRunners)

	wrpc.RegisterTaskCommunicatorServer(gRpcServer, workflowServer)
	log.Println("Workflow server starting...")

	if err = gRpcServer.Serve(tcpListener); err != nil {
		log.Println(err, "failed to start Workflow server")
	}

	return gRpcServer, port, nil
}

// GracefulShutdown handles the server shutdown process after receiving a signal
func GracefulShutdown(srv *grpc.Server) {
	log.Println("Termination signal received. Shutting down everything gracefully.")

	srv.GracefulStop()

	if srv != nil {
		srv.GracefulStop()
		log.Println("server stopped")
	}

	log.Println("Shutdown process completed.")
}

// SetupSigTermCloseHandler creates a 'listener' on a new goroutine which will
// notify the program if it receives an interrupt from the OS. We then handle
// this by calling our clean up procedure and exiting the program.
func SetupSigTermCloseHandler(srv *grpc.Server) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		GracefulShutdown(srv)
		os.Exit(0)
	}()
}
