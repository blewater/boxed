package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tradeline-tech/workflow/client"
)

func main() {
	serverAddress := "127.0.0.1"
	port := 8999
	err := client.StartWorkflow(context.Background(), serverAddress, port)
	if err != nil {
		fmt.Printf("server @ %s:%d with error : %s\n", serverAddress, port, err)
		os.Exit(1)
	}
}
