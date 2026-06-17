package cli

import (
	"os"
	"path/filepath"

	"github.com/javimosch/superlandings-go/internal/config"
)

func handleRemoteSiteSync(target, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}

	// Read local site directory and batch-write all files
	cfg, _ := config.Load()
	versionDir := filepath.Join(cfg.SitesDir, siteSlug, "v1")

	var files []map[string]string
	filepath.Walk(versionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(versionDir, path)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		files = append(files, map[string]string{"file": rel, "content": string(content)})
		return nil
	})

	if len(files) == 0 {
		fail(ExitInvalidInput, "no files found in local version directory")
	}

	_, err = client.WriteBatch(siteSlug, "v1", files)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}

	writeJSON(map[string]interface{}{
		"version": "1.0",
		"success": true,
		"files_synced": len(files),
	})
}
