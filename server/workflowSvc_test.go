package server

import (
	"os"
	"testing"

	"github.com/tradeline-tech/workflow/pkg/log"
	"github.com/tradeline-tech/workflow/remote"
	"github.com/tradeline-tech/workflow/types"
)

func TestPerformWorkflow(t *testing.T) {
	type args struct {
		soloWorkflow      bool
		srvAddress        string
		srvPort           int
		serverTaskRunners SrvTaskRunners
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
				if err = StartUp(
					true,
					tt.args.srvAddress,
					tt.args.srvPort,
					SrvTaskRunners{
						/*
						 * Workflow tasks
						 */
						types.RegisterRemoteTask("genGx"),
						NewGenGy,
						types.RegisterRemoteTask("genGyx"),
						NewGenGxy,
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
					NewGenGx,
					NewGenGyx,
					NewConfirm,
				}); err != nil {
				t.Error(err)
			}
		})
	}
}
