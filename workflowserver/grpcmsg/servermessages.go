package grpcmsg

import (
	"context"
	"fmt"

	"github.com/tradeline-tech/protos/namespace"

	"github.com/tradeline-tech/argo/models/workflow"
	"github.com/tradeline-tech/argo/pkg/logger"
)

//
// Collection of all server 2 Cli gRPC message functions
//

// SendCliErrMsg sends error message to the Cli
// errText2Cli: is the text message we will send to the Cli
// existingErrToRet to include in the response
func SendCliErrMsg(ctx context.Context,
	errText2Cli string,
	stream namespace.TaskCommunicator_RunTasksServer, existingErrToRet error) error {
	errSend := stream.Send(&namespace.ServerMsg{
		ErrorMsg: errText2Cli,
	})

	if errSend != nil {
		logger.Error(ctx, errSend, "sending server error failed")

		if existingErrToRet == nil {
			return errSend
		}

		return fmt.Errorf("workflow: %v, errSend:%v", existingErrToRet, errSend)
	}

	if existingErrToRet != nil {
		return existingErrToRet
	}

	return nil
}

// SendCliTasks sends tasks that need to be executed by the cli
func SendCliTasks(ctx context.Context, orgWorkflow *workflow.DBWorkflowType,
	stream namespace.TaskCommunicator_RunTasksServer) (sentCliTasks bool, err error) {
	if orgWorkflow.LastTaskIndexCompleted+1 >= len(orgWorkflow.Tasks) ||
		orgWorkflow.Tasks[orgWorkflow.LastTaskIndexCompleted+1].IsServer {
		return false, nil
	}

	cliTaskNames := make([]string, 0, 2)

	for i := orgWorkflow.LastTaskIndexCompleted + 1; i < len(orgWorkflow.Tasks); i++ {
		task := orgWorkflow.Tasks[i]

		if task.IsServer {
			break
		}

		cliTaskNames = append(cliTaskNames, task.Name)
	}

	if len(cliTaskNames) == 0 {
		return false, nil
	}

	errSend := stream.Send(&namespace.ServerMsg{
		ClientTasks: cliTaskNames,
	})

	if errSend != nil {
		logger.Error(ctx, errSend, "failed to send cli tasks", orgWorkflow.Token)

		return false, errSend
	}

	return true, nil
}

// SendCliCompletionMsg sends to cli that all work is done
func SendCliCompletionMsg(ctx context.Context,
	workflowServer *workflow.DBWorkflowType, stream namespace.TaskCommunicator_RunTasksServer) error {
	logger.Info(ctx, "The workflow completed sending completion message to Cli...")

	errSend := stream.Send(&namespace.ServerMsg{
		ServerTaskInProgress: "workflow_completed",
		ServerTaskOutPut:     fmt.Sprintf("The workflow completed execution for %d tasks.\n", len(workflowServer.Tasks)),
	})

	if errSend != nil {
		logger.Error(ctx, errSend, "failed to communicate cli tasks", workflowServer.Token)

		return errSend
	}

	return nil
}
