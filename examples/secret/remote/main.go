package main

import (
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
)

const workflowName = "dh-secret"

func main() {
	if err := remote.StartWorkflow(
		workflowName,
		"127.0.0.1",
		8999,
		types.TaskRunners{
			/*
			 * Set Remote Tasks here
			 */
			NewGenGx,
			NewGenGyx,
		}); err != nil {
		panic(err)
	}
}
