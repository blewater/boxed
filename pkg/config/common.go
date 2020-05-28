package config

type tasksBootstrapConfiguration struct {
	kv map[string]interface{}
}

type TaskConfiguration interface {
	Add(key string, value interface{})
	Get(key string) (interface{}, bool)
	// GetRPC() grpc.TaskCommunicator_RunWorkflowClient
	// SetRPC(grpc.TaskCommunicator_RunWorkflowClient)
}

func NewTaskConfiguration() TaskConfiguration {
	return &tasksBootstrapConfiguration{
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
func (cfg *tasksBootstrapConfiguration) Add(key string, value interface{}) {
	cfg.kv[key] = value
}

func (cfg *tasksBootstrapConfiguration) Get(key string) (interface{}, bool) {
	val, found := cfg.kv[key]

	return val, found
}
