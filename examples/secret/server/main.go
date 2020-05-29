package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tradeline-tech/workflow/server"
	"github.com/tradeline-tech/workflow/server/tasks"
)

func main() {
	ctx := context.Background()
	localhost := "127.0.0.1"
	srvPort := 8999
	gRPCServer, port, err := server.StartUp(ctx, localhost, srvPort, server.SrvTaskRunners{
		/*
		 * Server tasks
		 */
		tasks.NewGetWorkflowName,
	})
	if err != nil {
		fmt.Printf("server launching error : %s \n", err)
		os.Exit(1)
	}

	// Select ctrl-C to terminate server
	server.SetupSigTermCloseHandler(gRPCServer)

	fmt.Printf("Workflow server started @ localhost:%d\n", port)
	fmt.Println("Press Ctrl-C to exit server...")
}
