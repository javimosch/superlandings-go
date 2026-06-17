package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
)

// VersioningService handles auto-versioning, snapshots, pruning, and etag locking.
type VersioningService struct {
	cfg     *config.Config
	repo    *db.SiteVersionRepository
	siteRepo *db.SiteRepository
}

func NewVersioningService(cfg *config.Config) *VersioningService {
	return &VersioningService{
		cfg:      cfg,
		repo:     db.NewSiteVersionRepository(),
		siteRepo: db.NewSiteRepository(),
	}
}

// versionTimestamp generates a version label like "v1-20260617143005".
func versionTimestamp(base string) string {
	return base + "-" + time.Now().Format("20060102150405")
}

// AutoSave copies the current active version to a new directory, applies a single file write,
// and sets the new version as active. Returns the new version info.
func (s *VersioningService) AutoSave(siteSlug, filePath, content string) (*db.SiteVersion, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	active, err := s.repo.GetActiveVersion(site.ID)
	if err != nil {
		return nil, fmt.Errorf("no active version: %w", err)
	}

	// Copy active version directory to new timestamped dir
	baseVer := active.Version
	if idx := strings.LastIndex(baseVer, "-"); idx > 0 {
		if len(baseVer)-idx == 15 { // looks like a timestamp suffix
			baseVer = baseVer[:idx]
		}
	}
	newVer := versionTimestamp(baseVer)
	newDir := filepath.Join(s.cfg.SitesDir, siteSlug, newVer)

	// Resolve full path of active version
	activeFullPath := active.Path
	if !filepath.IsAbs(activeFullPath) {
		activeFullPath = strings.TrimPrefix(activeFullPath, "sites/")
		activeFullPath = filepath.Join(s.cfg.SitesDir, activeFullPath)
	}

	if err := copyDir(activeFullPath, newDir); err != nil {
		return nil, fmt.Errorf("copy version dir: %w", err)
	}

	// Apply the write to the new directory
	fullPath := filepath.Join(newDir, filePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("create parent dirs: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	// Create version record
	version := &db.SiteVersion{
		ID:       generateID(),
		SiteID:   site.ID,
		Version:  newVer,
		Path:     newDir,
		Comment:  fmt.Sprintf("auto: wrote %s", filePath),
		IsActive: true,
		Orphaned: false,
	}
	if err := s.repo.Create(version); err != nil {
		os.RemoveAll(newDir)
		return nil, fmt.Errorf("create version record: %w", err)
	}

	// Deactivate old and activate new
	if err := s.repo.SetActiveVersion(site.ID, version); err != nil {
		return nil, fmt.Errorf("set active version: %w", err)
	}

	return version, nil
}

// Snapshot creates a named, non-orphaned checkpoint of the active version.
func (s *VersioningService) Snapshot(siteSlug, name string) (*db.SiteVersion, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	active, err := s.repo.GetActiveVersion(site.ID)
	if err != nil {
		return nil, fmt.Errorf("no active version: %w", err)
	}

	newVer := name
	newDir := filepath.Join(s.cfg.SitesDir, siteSlug, newVer)

	// Check if snapshot name already exists
	if _, err := s.repo.GetBySiteAndVersion(site.ID, newVer); err == nil {
		newVer = versionTimestamp(name)
		newDir = filepath.Join(s.cfg.SitesDir, siteSlug, newVer)
	}

	// Resolve full path of active version
	activeFullPath := active.Path
	if !filepath.IsAbs(activeFullPath) {
		activeFullPath = strings.TrimPrefix(activeFullPath, "sites/")
		activeFullPath = filepath.Join(s.cfg.SitesDir, activeFullPath)
	}

	if err := copyDir(activeFullPath, newDir); err != nil {
		return nil, fmt.Errorf("copy version dir: %w", err)
	}

	version := &db.SiteVersion{
		ID:       generateID(),
		SiteID:   site.ID,
		Version:  newVer,
		Path:     newDir,
		Comment:  fmt.Sprintf("snapshot: %s", name),
		IsActive: true,
		Orphaned: false,
	}
	if err := s.repo.Create(version); err != nil {
		os.RemoveAll(newDir)
		return nil, fmt.Errorf("create snapshot record: %w", err)
	}

	if err := s.repo.SetActiveVersion(site.ID, version); err != nil {
		return nil, fmt.Errorf("set active version: %w", err)
	}

	return version, nil
}

// Rollback switches to a previous version and marks all later versions as orphaned.
func (s *VersioningService) Rollback(siteSlug, targetVersion string) (*db.SiteVersion, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	target, err := s.repo.GetBySiteAndVersion(site.ID, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("target version not found: %w", err)
	}

	// Mark all non-active versions after target as orphaned
	s.repo.MarkOrphanedAfter(site.ID, target.CreatedAt)

	// Set target as active (this also clears orphaned flag on target)
	if err := s.repo.SetActiveVersion(site.ID, target); err != nil {
		return nil, fmt.Errorf("set active: %w", err)
	}

	return target, nil
}

// PruneOrphaned deletes all orphaned version directories and their DB records.
func (s *VersioningService) PruneOrphaned(siteSlug string) (int, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return 0, fmt.Errorf("site not found: %w", err)
	}

	orphaned, err := s.repo.GetOrphanedVersions(site.ID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, v := range orphaned {
		os.RemoveAll(v.Path)
		if err := s.repo.DeleteVersion(v.ID); err == nil {
			count++
		}
	}
	return count, nil
}

// FileEtag returns a SHA256 hash of a file in the active version (for optimistic locking).
func (s *VersioningService) FileEtag(siteSlug, filePath string) (string, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return "", fmt.Errorf("site not found: %w", err)
	}

	active, err := s.repo.GetActiveVersion(site.ID)
	if err != nil {
		return "", fmt.Errorf("no active version: %w", err)
	}

	fullPath := filepath.Join(active.Path, filePath)
	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // new file, no etag needed
		}
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CheckEtag returns true if the etag matches the current file content.
// Used by write handlers to detect conflicts.
func (s *VersioningService) CheckEtag(siteSlug, filePath, etag string) (bool, error) {
	if etag == "" {
		return true, nil // no etag provided = skip check
	}
	current, err := s.FileEtag(siteSlug, filePath)
	if err != nil {
		return false, err
	}
	// New file: etag is empty string, but caller may have old etag "" → conflict
	return current == etag, nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
