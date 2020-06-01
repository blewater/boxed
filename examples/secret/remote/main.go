package main

import (
	"fmt"
	"os"

	"github.com/tradeline-tech/workflow/client"
	"github.com/tradeline-tech/workflow/client/tasks"
	"github.com/tradeline-tech/workflow/types"
)

func main() {
	serverAddress := "127.0.0.1"
	port := 8999
	err := client.StartWorkflow(serverAddress, port, types.TaskRunners{
		/*
		 * Set Remote Tasks here
		 */
		tasks.NewLaunchWorkflow,
	})
	if err != nil {
		fmt.Printf("server @ %s:%d with error : %s\n", serverAddress, port, err)
		os.Exit(1)
	}
}
