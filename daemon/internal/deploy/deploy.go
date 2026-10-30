package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

// UpdateFromRepo pulls the latest codebase from a git repository into a timestamped release,
// performs build steps, links shared assets, and atomically activates the release.
func (m *Manager) UpdateFromRepo(req protocol.UpdateFromRepoRequest) (*protocol.UpdateFromRepoResult, error) {
	if req.ProjectSlug == "" {
		return nil, fmt.Errorf("project_slug is required")
	}

	slug := req.ProjectSlug
	projDir := m.GetProjectDir(slug)
	repoURL := req.RepoURL
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	var logBuf strings.Builder
	logBuf.WriteString(fmt.Sprintf("[%s] Starting project update from repository for '%s'...\n", time.Now().Format(time.RFC3339), slug))

	// Resolve repo URL if not provided
	if repoURL == "" {
		currentGit := filepath.Join(projDir, "current", ".git")
		if _, err := os.Stat(currentGit); err == nil {
			cmd := exec.Command("git", "-C", filepath.Join(projDir, "current"), "config", "--get", "remote.origin.url")
			out, err := cmd.Output()
			if err == nil {
				repoURL = strings.TrimSpace(string(out))
			}
		}
	}
	if repoURL == "" {
		return nil, fmt.Errorf("no repository URL provided or configured for project '%s'", slug)
	}

	releaseTag := req.ReleaseTag
	if releaseTag == "" {
		releaseTag = time.Now().Format("20060102_150405")
	}

	releasesDir := filepath.Join(projDir, "releases")
	targetReleaseDir := filepath.Join(releasesDir, releaseTag)

	if err := os.MkdirAll(releasesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create releases directory: %w", err)
	}
	_ = os.RemoveAll(targetReleaseDir)

	logBuf.WriteString(fmt.Sprintf("Cloning repository '%s' (branch: %s)...\n", repoURL, branch))

	// Clone repository with depth 1
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, repoURL, targetReleaseDir)
	cloneOut, err := cloneCmd.CombinedOutput()
	if err != nil {
		// Fallback without --branch in case branch is master or default
		logBuf.WriteString(fmt.Sprintf("Direct branch clone warning: %s. Trying default branch clone...\n", string(cloneOut)))
		_ = os.RemoveAll(targetReleaseDir)
		fallbackCmd := exec.Command("git", "clone", "--depth", "1", repoURL, targetReleaseDir)
		fbOut, fbErr := fallbackCmd.CombinedOutput()
		if fbErr != nil {
			return nil, fmt.Errorf("git clone failed: %s (%w)", string(fbOut), fbErr)
		}
		logBuf.WriteString(string(fbOut) + "\n")
	} else {
		logBuf.WriteString(string(cloneOut) + "\n")
	}

	// Read commit metadata
	commitHash := ""
	if out, err := exec.Command("git", "-C", targetReleaseDir, "rev-parse", "--short", "HEAD").Output(); err == nil {
		commitHash = strings.TrimSpace(string(out))
	}
	commitMsg := ""
	if out, err := exec.Command("git", "-C", targetReleaseDir, "log", "-1", "--pretty=%s").Output(); err == nil {
		commitMsg = strings.TrimSpace(string(out))
	}
	commitAuthor := ""
	if out, err := exec.Command("git", "-C", targetReleaseDir, "log", "-1", "--pretty=%an").Output(); err == nil {
		commitAuthor = strings.TrimSpace(string(out))
	}

	logBuf.WriteString(fmt.Sprintf("Commit: %s (%s) by %s\n", commitHash, commitMsg, commitAuthor))

	// Ensure shared directory & resources
	sharedDir := filepath.Join(projDir, "shared")
	_ = os.MkdirAll(sharedDir, 0755)

	sharedEnv := filepath.Join(sharedDir, ".env")
	if _, err := os.Stat(sharedEnv); os.IsNotExist(err) {
		// If cloned repo has .env.example, copy it to shared/.env
		repoEnvEx := filepath.Join(targetReleaseDir, ".env.example")
		if exData, err := os.ReadFile(repoEnvEx); err == nil {
			_ = os.WriteFile(sharedEnv, exData, 0600)
		} else {
			_ = os.WriteFile(sharedEnv, []byte("# Generated environment variables\nAPP_ENV=production\n"), 0600)
		}
	}

	// Symlink shared .env
	relEnv := filepath.Join(targetReleaseDir, ".env")
	_ = os.Remove(relEnv)
	_ = os.Symlink(sharedEnv, relEnv)

	// Shared storage for PHP / Laravel frameworks
	sharedStorage := filepath.Join(sharedDir, "storage")
	_ = os.MkdirAll(sharedStorage, 0775)
	relStorage := filepath.Join(targetReleaseDir, "storage")
	_ = os.RemoveAll(relStorage)
	_ = os.Symlink(sharedStorage, relStorage)

	// Optional build steps
	if req.RunBuild {
		// Composer
		if _, err := os.Stat(filepath.Join(targetReleaseDir, "composer.json")); err == nil {
			if _, lookErr := exec.LookPath("composer"); lookErr == nil {
				logBuf.WriteString("Running composer install...\n")
				compCmd := exec.Command("composer", "install", "--no-dev", "--no-interaction", "--optimize-autoloader")
				compCmd.Dir = targetReleaseDir
				compOut, _ := compCmd.CombinedOutput()
				logBuf.WriteString(string(compOut) + "\n")
			}
		}

		// NPM
		if _, err := os.Stat(filepath.Join(targetReleaseDir, "package.json")); err == nil {
			if _, lookErr := exec.LookPath("npm"); lookErr == nil {
				logBuf.WriteString("Running npm install & build...\n")
				npmCmd := exec.Command("npm", "install", "--no-audit", "--no-fund")
				npmCmd.Dir = targetReleaseDir
				npmOut, _ := npmCmd.CombinedOutput()
				logBuf.WriteString(string(npmOut) + "\n")

				buildCmd := exec.Command("npm", "run", "build")
				buildCmd.Dir = targetReleaseDir
				buildOut, _ := buildCmd.CombinedOutput()
				logBuf.WriteString(string(buildOut) + "\n")
			}
		}
	}

	// Atomically activate release
	actRes, err := m.ActivateRelease(slug, releaseTag)
	if err != nil {
		return nil, fmt.Errorf("failed to activate release '%s': %w", releaseTag, err)
	}

	// Cleanup old releases
	_, _ = m.CleanupOldReleases(slug, 5)

	logBuf.WriteString(fmt.Sprintf("[%s] Release '%s' successfully deployed and activated.\nSymlink: %s\n",
		time.Now().Format(time.RFC3339), releaseTag, actRes.CurrentSymlink))

	return &protocol.UpdateFromRepoResult{
		Success:        true,
		ProjectSlug:    slug,
		ReleaseTag:     releaseTag,
		CommitHash:     commitHash,
		CommitMessage:  commitMsg,
		CommitAuthor:   commitAuthor,
		Branch:         branch,
		Logs:           logBuf.String(),
		CurrentSymlink: actRes.CurrentSymlink,
		TargetRelease:  targetReleaseDir,
	}, nil
}

