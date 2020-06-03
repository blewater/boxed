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
	"github.com/tradeline-tech/workflow/pkg/config"
	"github.com/tradeline-tech/workflow/types"
	"github.com/tradeline-tech/workflow/wrpc"
)

type WorkflowsType map[string]*types.Tasks

// SrvTaskRunners is the list of declared constructor tasks
// to run for the workflow tasks mapping one runner to one task
// they run in the sequence listed here.
type SrvTaskRunners []types.TaskRunnerNewFunc

// Server Workflow Loop Step Algorithm Actions
const (
	// Continue processing. Non-disrupting action effect.
	continueProcessing = 0
	// Stop processing this request.
	haltRequestAction = 1
	// Go back to listening for remote messages
	// typically this is desirable after sending
	// tasks for remote execution.
	skipLoopAction = 2
)

type WorkflowsServer struct {
	wrpc.UnimplementedTaskCommunicatorServer
	gRPCServer wrpc.TaskCommunicator_RunWorkflowServer

	srvTaskRunners SrvTaskRunners

	// Serialize access to concurrent workflows config access
	mu sync.Mutex

	// Cached org workflows, config
	workflows   WorkflowsType
	TaskRunners []types.TaskRunnerNewFunc
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
func NewWorkflowsServer(srvTaskRunners SrvTaskRunners) *WorkflowsServer {
	return &WorkflowsServer{
		srvTaskRunners: srvTaskRunners,
		workflows:      make(WorkflowsType),
		// Link the declared task runners above with the workflow server
		TaskRunners: srvTaskRunners,
	}
}

func (srv *WorkflowsServer) GetgRPCServer() wrpc.TaskCommunicator_RunWorkflowServer {
	return srv.gRPCServer
}

// RunTasks opens a bidirectional stream to a remote client and
// runs tasks with workflow name-key received from the remote gRPC client.
func (srv *WorkflowsServer) RunWorkflow(stream wrpc.TaskCommunicator_RunWorkflowServer) error {
	var wReq = WorkflowServerReq{WorkflowServer: srv}
	defer recoverFromPanic()

	ctx := stream.Context()

	defer req.safeSaveWorkflow()

	// Could re-enter when execution flow resumes from remote tasks
	for {
		clientMsg, err := stream.Recv()

		if stepCheckIOError(ctx, clientMsg, err) {
			break
		}

		if loopAction := wReq.stepProcessReceivedMessages(ctx, clientMsg); loopAction == haltRequestAction {
			break
		} else if loopAction == skipLoopAction {
			continue
		}

		if stepWorkflowCompletedRemotely(ctx, clientMsg, wReq.Workflow, stream) {
			break
		}

		if err = stepRunServerSideTasks(ctx, wReq.Workflow, stream, clientMsg); err != nil {
			break
		}

		if loopAction := stepSendRemoteTasks(wReq.Workflow, stream); loopAction == haltRequestAction {
			break
		} else if loopAction == skipLoopAction {
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

func (srv *WorkflowsServer) getMemCachedWorkflow(workflowNameKey string) *types.Tasks {
	if workflowNameKey == "" {
		return nil
	}

	srv.mu.Lock()
	// attempt to get cached workflow
	retWorkflow, _ := srv.workflows[workflowNameKey]
	srv.mu.Unlock()

	return retWorkflow
}

func (srv *WorkflowsServer) setMemCachedWorkflow(workflowNameKey string, workflow *types.Tasks) {
	if workflowNameKey == "" {
		return
	}

	srv.mu.Lock()
	srv.workflows[workflowNameKey] = workflow
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


		return lastServerTaskError
	}

	clientErrMsg := remoteMsg.ErrorMsg
	if clientErrMsg != "" {
		clientErr := errors.New(clientErrMsg)
		fmt.Println(clientErr, "received client error")

		return clientErr
	}

	return nil
}

func StartUp(
	srvAddress string,
	srvPort int, serverTaskRunners SrvTaskRunners) (gRpcServer *grpc.Server, port int, err error) {
	srvAddressPort := fmt.Sprintf("%s:%d", srvAddress, srvPort)
	tcpListener, err := net.Listen("tcp", srvAddressPort)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen at tcp %s", srvAddressPort)
	}
	log.Println("listening at ", srvAddressPort)
	gRpcServer = grpc.NewServer()

	workflowServer := NewWorkflowsServer(serverTaskRunners)

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
