package config

type TasksBootstrapConfig struct {
	kv map[string]interface{}
}

func NewTasksBoostrapConfig() *TasksBootstrapConfig {
	return &TasksBootstrapConfig{
		kv: make(map[string]interface{}),
	}
}

// func (cfg *tasksBootstrapConfiguration) GetRPC() grpc.TaskCommunicator_RunWorkflowClient {
// 	return cfg.gRPC
// }
//
// func (cfg *tasksBootstrapConfiguration) SetRPC(
// 	gRPC grpc.TaskCommunicator_RunWorkflowClient) {
// 	cfg.gRPC = gRPC
// 	cfg.Add(GRPCKey, gRPC)
// }
//
func (cfg *TasksBootstrapConfig) Add(key string, value interface{}) {
	cfg.kv[key] = value
}

func (cfg *TasksBootstrapConfig) Get(key string) (interface{}, bool) {
	val, found := cfg.kv[key]

	return val, found
}
