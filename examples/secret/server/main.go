package main

import (
	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/server"
	"github.com/tradeline-tech/workflow/server/tasks"
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
			tasks.RemoteTask("genGx"),
			NewGenGy,
			tasks.RemoteTask("genGyx"),
		}); err != nil {
		panic("server launching error : " + err.Error())
	}
}
