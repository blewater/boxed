// Package common declares shared types for server and remotes.
package common

const ServerWorkflowCompletionText = "workflow_completed"

type RemoteTaskRunnersByKey map[string]TaskRunnerNewFunc
