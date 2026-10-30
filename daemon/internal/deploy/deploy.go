package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

type Manager struct {
	baseDir string // default: /var/www/vps-panel/projects
}

func NewManager(baseDir string) *Manager {
	if baseDir == "" {
		baseDir = "/var/www/vps-panel/projects"
	}
	return &Manager{baseDir: baseDir}
}

// GetProjectDir returns the base path for a project
func (m *Manager) GetProjectDir(slug string) string {
	return filepath.Join(m.baseDir, slug)
}

// PrepareRelease creates directories for a new release and links shared resources
func (m *Manager) PrepareRelease(slug, releaseTag string) (*protocol.PrepareReleaseResult, error) {
	if slug == "" || releaseTag == "" {
		return nil, fmt.Errorf("slug and release_tag are required")
	}

	projDir := m.GetProjectDir(slug)
	releasesDir := filepath.Join(projDir, "releases")
	targetReleaseDir := filepath.Join(releasesDir, releaseTag)
	sharedDir := filepath.Join(projDir, "shared")

	// Create directories
	if err := os.MkdirAll(targetReleaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create release dir: %w", err)
	}
	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create shared dir: %w", err)
	}

	// Create shared/.env if it does not exist
	sharedEnv := filepath.Join(sharedDir, ".env")
	if _, err := os.Stat(sharedEnv); os.IsNotExist(err) {
		_ = os.WriteFile(sharedEnv, []byte("# Generated environment variables\nAPP_ENV=production\n"), 0600)
	}

	// Create shared/storage for frameworks like Laravel
	sharedStorage := filepath.Join(sharedDir, "storage")
	if err := os.MkdirAll(sharedStorage, 0775); err != nil {
		return nil, fmt.Errorf("failed to create shared storage: %w", err)
	}

	// Symlink shared items into release dir
	relEnv := filepath.Join(targetReleaseDir, ".env")
	_ = os.Remove(relEnv)
	_ = os.Symlink(sharedEnv, relEnv)

	relStorage := filepath.Join(targetReleaseDir, "storage")
	_ = os.Remove(relStorage)
	_ = os.Symlink(sharedStorage, relStorage)

	return &protocol.PrepareReleaseResult{
		ReleasePath: targetReleaseDir,
		SharedPath:  sharedDir,
	}, nil
}

// ActivateRelease updates the 'current' symlink atomically
func (m *Manager) ActivateRelease(slug, releaseTag string) (*protocol.ActivateReleaseResult, error) {
	projDir := m.GetProjectDir(slug)
	targetReleaseDir := filepath.Join(projDir, "releases", releaseTag)

	if _, err := os.Stat(targetReleaseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("release %s does not exist at %s", releaseTag, targetReleaseDir)
	}

	currentSymlink := filepath.Join(projDir, "current")
	tmpSymlink := filepath.Join(projDir, "current.tmp")

	_ = os.Remove(tmpSymlink)

	// Create relative or absolute symlink
	relPath, err := filepath.Rel(projDir, targetReleaseDir)
	if err != nil {
		relPath = targetReleaseDir
	}

	if err := os.Symlink(relPath, tmpSymlink); err != nil {
		return nil, fmt.Errorf("failed to create temporary symlink: %w", err)
	}

	// Atomic rename to replace 'current'
	if err := os.Rename(tmpSymlink, currentSymlink); err != nil {
		_ = os.Remove(tmpSymlink)
		return nil, fmt.Errorf("failed to atomically activate symlink: %w", err)
	}

	return &protocol.ActivateReleaseResult{
		Success:        true,
		CurrentSymlink: currentSymlink,
		TargetRelease:  targetReleaseDir,
	}, nil
}

// RollbackRelease switches 'current' to the specified target release or the previous release if empty
func (m *Manager) RollbackRelease(slug, targetReleaseTag string) error {
	if targetReleaseTag == "" {
		releasesDir := filepath.Join(m.GetProjectDir(slug), "releases")
		entries, err := os.ReadDir(releasesDir)
		if err != nil || len(entries) < 2 {
			return fmt.Errorf("no previous release found to roll back to in %s", releasesDir)
		}

		var dirNames []string
		for _, entry := range entries {
			if entry.IsDir() {
				dirNames = append(dirNames, entry.Name())
			}
		}
		sort.Strings(dirNames)

		if len(dirNames) < 2 {
			return fmt.Errorf("not enough releases to perform rollback (found %d)", len(dirNames))
		}

		// Second to last is the previous release
		targetReleaseTag = dirNames[len(dirNames)-2]
	}

	_, err := m.ActivateRelease(slug, targetReleaseTag)
	return err
}

// CleanupOldReleases deletes older releases, retaining only the most recent keepCount
func (m *Manager) CleanupOldReleases(slug string, keepCount int) ([]string, error) {
	if keepCount <= 0 {
		keepCount = 5
	}

	releasesDir := filepath.Join(m.GetProjectDir(slug), "releases")
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		return nil, nil // No releases yet
	}

	var dirNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		}
	}

	// Sort chronologically (assuming release tags are timestamps or lexicographically ordered)
	sort.Strings(dirNames)

	if len(dirNames) <= keepCount {
		return nil, nil
	}

	var removed []string
	toRemoveCount := len(dirNames) - keepCount
	for i := 0; i < toRemoveCount; i++ {
		targetDir := filepath.Join(releasesDir, dirNames[i])
		if err := os.RemoveAll(targetDir); err == nil {
			removed = append(removed, dirNames[i])
		}
	}

	return removed, nil
}
