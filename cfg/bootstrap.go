package cfg

import (
	"github.com/tradeline-tech/workflow/grpc"
)

// While this configuration could be bootstrapped with the configuration of the gRPC Server
// it could be the business case that its content is known upon the 1st gRPC Workflow Server request.
// TasksBootstrapConfiguration is a map collection of configuration pairs of
// string key, generic (interface{}) values
type TasksBootstrapConfiguration struct {
	kv      map[string]interface{}
	gRPCSrv grpc.TaskCommunicator_RunWorkflowServer
}

var (
	config = &TasksBootstrapConfiguration{
		kv: make(map[string]interface{}),
	}
)

func Get() *TasksBootstrapConfiguration {
	return config
}

func (cfg *TasksBootstrapConfiguration) GetgRPCSrv() grpc.TaskCommunicator_RunWorkflowServer {
	return config.gRPCSrv
}

func (cfg *TasksBootstrapConfiguration) SetgRPCSrv(
	gRPCSrv grpc.TaskCommunicator_RunWorkflowServer) *TasksBootstrapConfiguration {
	cfg.gRPCSrv = gRPCSrv

	return cfg
}

func (cfg *TasksBootstrapConfiguration) NewTasksStream(
	stream grpc.TaskCommunicator_RunWorkflowServer) *TasksBootstrapConfiguration {
	cfg.gRPCSrv = stream

	return cfg
}

func (cfg *TasksBootstrapConfiguration) Add(key string, value interface{}) *TasksBootstrapConfiguration {
	cfg.kv[key] = value

	return cfg
}

func (cfg *TasksBootstrapConfiguration) Get(key string) interface{} {
	return cfg.kv[key]
}
