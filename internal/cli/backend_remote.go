package cli

func handleRemoteBackendStatus(target string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.GetStatus()
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "status": result})
}
