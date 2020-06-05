package config

import (
	"fmt"
	"strconv"
	"strings"
)

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

func GetValueNotFoundErrFunc(v string) error {
	return fmt.Errorf("%s not found within configuration", strings.TrimSpace(v))
}

func (cfg *TasksBootstrapConfig) GetInt64(key string) (int64, error) {
	interfaceVal, ok := cfg.Get(key)
	if !ok {
		return 0, GetValueNotFoundErrFunc(key)
	}

	var int64Val int64
	switch v := interfaceVal.(type) {
	case int64:
		int64Val = v
	case int:
		int64Val = int64(v)
	case string:
		intVal, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		int64Val = int64(intVal)
	default:
		return 0, fmt.Errorf("%v of type %T", interfaceVal, interfaceVal)
	}

	return int64Val, nil
}
