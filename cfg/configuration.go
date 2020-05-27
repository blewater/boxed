package cfg

import (
	"github.com/tradeline-tech/workflow/grpc"
)

const (
	GRPCSrvKey = "grpcsrv"
	GRPCKey    = "grpc"
)

type tasksBootstrapConfiguration struct {
	kv      map[string]interface{}
	gRPCSrv grpc.TaskCommunicator_RunWorkflowServer
	gRPC    grpc.TaskCommunicator_RunWorkflowClient
}

type TaskConfiguration interface {
	Add(key string, value interface{})
	Get(key string) (interface{}, bool)
	GetRPCSrv() grpc.TaskCommunicator_RunWorkflowServer
	SetRPCSrv(grpc.TaskCommunicator_RunWorkflowServer)
	GetRPC() grpc.TaskCommunicator_RunWorkflowClient
	SetRPC(grpc.TaskCommunicator_RunWorkflowClient)
}

func NewTaskConfiguration() TaskConfiguration {
	return &tasksBootstrapConfiguration{
		kv: make(map[string]interface{}),
	}
}

func (cfg *tasksBootstrapConfiguration) GetRPCSrv() grpc.TaskCommunicator_RunWorkflowServer {
	return cfg.gRPCSrv
}

func (cfg *tasksBootstrapConfiguration) SetRPCSrv(
	gRPCSrv grpc.TaskCommunicator_RunWorkflowServer) {
	cfg.gRPCSrv = gRPCSrv
	cfg.Add(GRPCSrvKey, gRPCSrv)
}

func (cfg *tasksBootstrapConfiguration) GetRPC() grpc.TaskCommunicator_RunWorkflowClient {
	return cfg.gRPC
}

func (cfg *tasksBootstrapConfiguration) SetRPC(
	gRPC grpc.TaskCommunicator_RunWorkflowClient) {
	cfg.gRPC = gRPC
	cfg.Add(GRPCKey, gRPC)
}

func (cfg *tasksBootstrapConfiguration) Add(key string, value interface{}) {
	cfg.kv[key] = value
}

func (cfg *tasksBootstrapConfiguration) Get(key string) (interface{}, bool) {
	val, found := cfg.kv[key]

	return val, found
}
