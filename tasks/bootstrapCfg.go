package tasks

import (
	"github.com/tradeline-tech/workflow/msg"
)

// TasksBootstrapConfiguration is a map collection of configuration string key
// generic (interface{}) value pairs
type TasksBootstrapConfiguration struct {
	kv map[string]interface{}
	stream msg.TaskCommunicator_RunTasksServer
}

var (
	config = &TasksBootstrapConfiguration{
		kv: make(map[string]interface{}),
	}
)

func GetConfig() *TasksBootstrapConfiguration{
	return config
}

func (cfg *TasksBootstrapConfiguration)NewTasksStream(
	stream msg.TaskCommunicator_RunTasksServer) *TasksBootstrapConfiguration{
	cfg.stream = stream

	return cfg
}

func (cfg *TasksBootstrapConfiguration)Add(key string, value interface{}) *TasksBootstrapConfiguration{
	cfg.kv[key] = value

	return cfg
}

func (cfg *TasksBootstrapConfiguration)Get(key string) interface{}{
	return cfg.kv[key]
}
