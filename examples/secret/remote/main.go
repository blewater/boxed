package main

import (
	"github.com/tradeline-tech/workflow/examples/secret/tasks"
	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
)

const workflowName = "dh-secret"

func main() {
	// Sssss it's secret
	log.DiscardLog()

	if err := remote.StartWorkflow(
		workflowName,
		"127.0.0.1",
		8999,
		types.TaskRunners{
			/*
			 * Set Remote Tasks here
			 */
			tasks.NewGenGx,
			tasks.NewGenGyx,
			tasks.NewValidate,
		}); err != nil {
		panic(err)
	}
}
