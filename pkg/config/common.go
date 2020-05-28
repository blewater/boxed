package config

type TaskConfiguration interface {
	Add(key string, value interface{})
	Get(key string) (interface{}, bool)
	// GetRPC() grpc.TaskCommunicator_RunWorkflowClient
	// SetRPC(grpc.TaskCommunicator_RunWorkflowClient)
}
