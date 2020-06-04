package config

type TasksBootstrapConfig struct {
	kv map[string]interface{}
}

func NewTasksBootstrapConfig() *TasksBootstrapConfig {
	return &TasksBootstrapConfig{
		kv: make(map[string]interface{}),
	}
}

func (cfg *TasksBootstrapConfig) Add(key string, value interface{}) {
	cfg.kv[key] = value
}

func (cfg *TasksBootstrapConfig) Get(key string) (interface{}, bool) {
	val, found := cfg.kv[key]

	return val, found
}
