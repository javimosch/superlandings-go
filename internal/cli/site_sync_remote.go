package cli

func handleRemoteSiteSync(target, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}

	payload := map[string]interface{}{"site_slug": siteSlug}
	result, err := client.SyncSite(siteSlug, payload)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}

	writeJSON(map[string]interface{}{
		"version": "1.0",
		"sync":    result,
	})
}
