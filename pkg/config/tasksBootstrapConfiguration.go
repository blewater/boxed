package config

type tasksBootstrapConfiguration struct {
	kv map[string]interface{}
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
