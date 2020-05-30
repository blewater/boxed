// Package common declares shared types for server and remotes.
package types

const WorkflowKey = "workflowKey"

type RemoteTaskRunnersByKey map[string]TaskRunnerNewFunc
