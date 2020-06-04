package main

import (
	"github.com/tradeline-tech/workflow/examples/secret/tasks"
	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/server"
	"github.com/tradeline-tech/workflow/types"
)

func main() {
	// Sssss it's secret
	log.DiscardLog()

	if err := server.StartUp(
		true,
		"127.0.0.1",
		8999,
		server.SrvTaskRunners{
			/*
			 * Workflow tasks
			 */
			types.RegisterRemoteTask("genGx"),
			tasks.NewGenGy,
			types.RegisterRemoteTask("genGyx"),
		}); err != nil {
		panic("server launching error : " + err.Error())
	}
}
