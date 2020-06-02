package main

import (
	"fmt"
	"os"

	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
)

const WorkflowNameKey = "secret"

func main() {
	serverAddress := "127.0.0.1"
	port := 8999
	err := remote.StartWorkflow(serverAddress, port, types.TaskRunners{
		/*
		 * Set Remote Tasks here
		 */
		NewGenGx,
	})
	if err != nil {
		fmt.Printf("server @ %s:%d with error : %s\n", serverAddress, port, err)
		os.Exit(1)
	}
}
