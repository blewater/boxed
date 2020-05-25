package msg

func TaskUpdateToRemote(s2r TaskCommunicator_RunTasksServer, taskName, msgText string, errIn error) error {
	msg := &ServerMsg{}

	if errIn != nil {
		msg.ErrorMsg = errIn.Error()
		return s2r.Send(msg)
	}

	msg.TaskInProgress = taskName
	msg.TaskOutput = msgText

	return s2r.Send(msg)
}
