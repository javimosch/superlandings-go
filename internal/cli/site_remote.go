package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

func handleRemoteSiteList(target string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.ListSites()
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "sites": json.RawMessage(data)})
}

func handleRemoteVersionList(target string, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.ListVersions(siteSlug)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "versions": json.RawMessage(data)})
}

func handleRemoteVersionCreate(target string, args []string, cmd *cobra.Command) {
	version, _ := cmd.Flags().GetString("version")
	comment, _ := cmd.Flags().GetString("comment")
	author, _ := cmd.Flags().GetString("author")

	if version == "" {
		fail(ExitMissingFlag, "--version is required")
	}

	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.CreateVersion(args[0], version, comment, author)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "result": json.RawMessage(data)})
}

func handleRemoteVersionSwitch(target string, siteSlug, version string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.SwitchVersion(siteSlug, version)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "result": json.RawMessage(data)})
}

func handleRemoteSiteWrite(target string, args []string, cmd *cobra.Command) {
	content, _ := cmd.Flags().GetString("content")
	if content == "" {
		fail(ExitMissingFlag, "--content is required")
	}

	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.WriteFile(args[0], args[1], args[2], content)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "result": json.RawMessage(data)})
}

func handleRemoteDNSList(target string, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.ListDNS(siteSlug)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "dns": json.RawMessage(data)})
}

func handleRemoteDNSSetup(target string, args []string, cmd *cobra.Command) {
	domain, _ := cmd.Flags().GetString("domain")
	ip, _ := cmd.Flags().GetString("ip")
	traefik, _ := cmd.Flags().GetBool("traefik")

	if domain == "" {
		fail(ExitMissingFlag, "--domain is required")
	}
	if ip == "" {
		fail(ExitMissingFlag, "--ip is required")
	}

	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.SetupDNS(args[0], domain, ip, traefik)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "result": json.RawMessage(data)})
}

func handleRemoteDNSRemove(target string, args []string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.RemoveDNS(args[0])
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	data, _ := json.Marshal(result)
	writeJSON(map[string]interface{}{"version": "1.0", "result": json.RawMessage(data)})
}

func handleRemoteSiteUpload(target string, siteSlug, assetPath string, data []byte) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.UploadAsset(siteSlug, assetPath, data)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}
