package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/tradeline-tech/workflow/datastore"
	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

type WorkflowsType map[string]*types.Workflow
type WorkflowsConfigType map[string]types.TaskConfiguration

// SrvTaskRunners is the list of declared constructor tasks
// to run for the workflow tasks mapping one runner to one task
// they run in the sequence listed here.
type SrvTaskRunners []types.TaskRunnerNewFunc

// Server Workflow Loop Step Algorithm Actions
type reqAction int

const (
	// Continue processing. Non-disrupting action effect.
	continueProcessing reqAction = iota
	// Cease processing requests and exit server.
	exitServer
	// Go back to listening for remote messages
	// typically this is desirable after sending
	// tasks for remote execution.
	waitNextRequest

	ConfigServerMessengerKey = "serverMessenger"
)

type WorkflowsServer struct {
	// gRPC state and Api
	server *grpc.Server
	wrpc.UnimplementedTaskCommunicatorServer

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

func recoverFromPanic(errRef *error) {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if ok {
			*errRef = err
		}
	}
}

// NewWorkflowsServer initializes a workflow server for servicing
// a business group of with an instance of a workflow.
// Multiple workflows with the same declared tasks could run concurrently
// by one Workflow server.
func NewWorkflowsServer(server *grpc.Server, soloWorkflow bool, srvTaskRunners SrvTaskRunners) *WorkflowsServer {
	return &WorkflowsServer{
		server:           server,
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
func (srv *WorkflowsServer) RunWorkflow(gRPCConnToRemote wrpc.TaskCommunicator_RunWorkflowServer) (err error) {
	defer recoverFromPanic(&err)

	ctx := gRPCConnToRemote.Context()

	req := NewWorkflowReq(srv, gRPCConnToRemote)

	for {
		remoteMsq, err := gRPCConnToRemote.Recv()

		if stepCheckIOError(ctx, remoteMsq, err) {
			break
		}

		if nextAction := req.process(ctx, remoteMsq); nextAction == waitNextRequest {
			continue
		} else if nextAction == exitServer {
			break
		}
	}
	req.workflow.Display()

	srv.soloModeTerminate()

	return nil
}

// soloModeTerminate blocks new connections and messages and wait for client
// connections to release before timing out and stopping forcefully.
func (srv *WorkflowsServer) soloModeTerminate() {
	if srv.soloWorkflowMode {
		log.Println("Exiting because in solo workflow mode...")

		stopChan := make(chan struct{})
		go func() {
			srv.server.GracefulStop()
			close(stopChan)
		}()

		timer := time.NewTimer(types.ConnectionTimeoutSec * time.Second)
		select {
		case <-timer.C:
			srv.server.Stop()
		case <-stopChan:
			timer.Stop()
		}
	}
}

// 1. Check whether the protocol errs
func stepCheckIOError(ctx context.Context, remoteMsg *wrpc.RemoteMsg, err error) (exit bool) {
	if err == io.EOF {
		log.Println("remote severed the connection to the server, EOF")

		return true
	}

	if err != nil {
		log.Println(err, ", server message reception error. Is the remote client still up?")

		return true
	}

	if errClient := ctx.Err(); errClient != nil {
		log.Println("server reception error; Is the remote client still up?")

		return true
	}

	if remoteMsg == nil {
		log.Println(err, ", "+
			"received nil remote message. Remote likely closed the connection.")

		return true
	}

	return false
}

func (srv *WorkflowsServer) getMemCachedWorkflow(workflowNameKey string) (*types.Workflow, types.TaskConfiguration) {
	if workflowNameKey == "" {
		return nil, nil
	}

	srv.mu.Lock()
	// attempt to get cached workflow
	retWorkflow := srv.workflows[workflowNameKey]
	retWorkflowConfig := srv.workflowsConfig[workflowNameKey]
	srv.mu.Unlock()

	return retWorkflow, retWorkflowConfig
}

func (srv *WorkflowsServer) setMemCachedWorkflow(
	workflowNameKey string, workflow *types.Workflow, cfg types.TaskConfiguration) {
	if workflowNameKey == "" {
		return
	}

	srv.mu.Lock()
	srv.workflows[workflowNameKey] = workflow
	srv.workflowsConfig[workflowNameKey] = cfg
	srv.mu.Unlock()
}

func readWorkflowFromDB(ctx context.Context, workflowNameKey string) (*types.Workflow, error) {
	readDocument, err := datastore.LoadByName(ctx, workflowNameKey)
	if err != nil {
		return nil, err
	}

	// No workflow found or db connection not used
	if readDocument == nil {
		return nil, nil
	}

	// type assert that the read document is a tasks type collection
	foundWorkflow, okWorkflow := readDocument.(*types.Workflow)
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

// Startup is Cli helper to start a gRPC workflow server without secure transport.
func StartUp(soloWorkflow bool, srvAddress string, srvPort int, serverTaskRunners SrvTaskRunners) error {
	srvAddressPort := fmt.Sprintf("%s:%d", srvAddress, srvPort)
	tcpListener, err := net.Listen("tcp", srvAddressPort)

	if err != nil {
		return fmt.Errorf("failed to listen at tcp %s", srvAddressPort)
	}

	log.Println("listening at ", srvAddressPort)
	gRpcServer := grpc.NewServer()

	workflowServer := NewWorkflowsServer(gRpcServer, soloWorkflow, serverTaskRunners)

	wrpc.RegisterTaskCommunicatorServer(gRpcServer, workflowServer)
	log.Println("Workflow server starting...")

	if err = gRpcServer.Serve(tcpListener); err != nil {
		log.Println(err, "failed to start Workflow server")
	}

	return err
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
		log.Println("Ctrl+C pressed in Terminal")
		GracefulShutdown(srv)
		os.Exit(0)
	}()
}
