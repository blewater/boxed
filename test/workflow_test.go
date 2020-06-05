package test

import (
	"os"
	"testing"

	"github.com/tradeline-tech/workflow/examples/secret/tasks"
	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/server"
	"github.com/tradeline-tech/workflow/test/ttasks"
	"github.com/tradeline-tech/workflow/types"
)

func TestPerformWorkflow(t *testing.T) {
	type args struct {
		soloWorkflow      bool
		srvAddress        string
		srvPort           int
		serverTaskRunners server.SrvTaskRunners
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{
		name: "soloMutualAgreeingCalculation",
		args: args{
			soloWorkflow:      true,
			srvAddress:        "127.0.0.1",
			srvPort:           7492,
			serverTaskRunners: nil,
		},
		wantErr: false,
	},
	}
	log.InitLog(os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			// release main go routine, run Server in the background
			go func() {
				if err = server.StartUp(
					true,
					tt.args.srvAddress,
					tt.args.srvPort,
					server.SrvTaskRunners{
						/*
						 * Workflow tasks
						 */
						types.RegisterRemoteTask("genGx"),
						tasks.NewGenGy,
						types.RegisterRemoteTask("genGyx"),
						tasks.NewGenGxy,
						types.RegisterRemoteTask("confirm"),
					}); err != nil {
					t.Errorf("error server failed to launch : %s", err.Error())
				}
			}()

			// Remote
			if err = remote.StartWorkflow(
				tt.name,
				tt.args.srvAddress,
				tt.args.srvPort,
				types.TaskRunners{
					/*
					 * Set Remote Tasks here
					 */
					ttasks.NewGenGx,
					ttasks.NewGenGyx,
					ttasks.NewConfirm,
				}); err != nil {
				t.Error(err)
			}
		})
	}
}
