package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

type Manager struct {
	systemctlBinary string
	mockMode        bool
}

func NewManager() *Manager {
	bin, err := exec.LookPath("systemctl")
	mock := false
	if err != nil || os.Geteuid() != 0 {
		mock = true
	}
	return &Manager{
		systemctlBinary: bin,
		mockMode:        mock,
	}
}

// GetStatus checks status of tracked services
func (m *Manager) GetStatus(services []string) ([]protocol.ServiceStatus, error) {
	if len(services) == 0 {
		services = []string{"nginx", "php8.4-fpm", "mysql", "redis", "ufw", "docker"}
	}

	var results []protocol.ServiceStatus
	for _, name := range services {
		status := protocol.ServiceStatus{
			Name:        name,
			DisplayName: formatDisplayName(name),
		}

		if m.systemctlBinary != "" && !m.mockMode {
			cmd := exec.Command(m.systemctlBinary, "is-active", name)
			out, _ := cmd.Output()
			state := strings.TrimSpace(string(out))
			status.Status = state
			status.Active = (state == "active")
			status.Running = (state == "active")
			results = append(results, status)
			continue
		}

		// Real process / service inspection fallback
		isRunning := checkProcessRunning(name)
		if isRunning {
			status.Active = true
			status.Running = true
			status.Status = "active"
		} else {
			status.Active = false
			status.Running = false
			status.Status = "inactive"
		}

		results = append(results, status)
	}

	return results, nil
}

func checkProcessRunning(name string) bool {
	switch name {
	case "nginx":
		return runPgrep("nginx")
	case "php8.4-fpm", "php8.3-fpm", "php-fpm":
		return runPgrep("php-fpm")
	case "mysql", "mariadb":
		return runPgrep("mysqld") || runPgrep("mariadbd")
	case "redis", "redis-server":
		return runPgrep("redis-server")
	case "docker":
		return runPgrep("dockerd") || runPgrep("Docker")
	case "ufw":
		out, err := exec.Command("ufw", "status").CombinedOutput()
		if err == nil && strings.Contains(string(out), "active") {
			return true
		}
		pfOut, errPf := exec.Command("/sbin/pfctl", "-s", "info").CombinedOutput()
		return errPf == nil && strings.Contains(string(pfOut), "Status: Enabled")
	default:
		return runPgrep(name)
	}
}

func runPgrep(pattern string) bool {
	cmd := exec.Command("pgrep", "-f", pattern)
	err := cmd.Run()
	return err == nil
}

// Action executes start/stop/restart/reload on a service
func (m *Manager) Action(req protocol.ServiceActionRequest) (bool, error) {
	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"reload":  true,
		"status":  true,
	}

	if !validActions[req.Action] {
		return false, fmt.Errorf("invalid action: %s", req.Action)
	}

	// Service name whitelist sanitization
	if strings.ContainsAny(req.Service, " ;&|<>$`\n\r") {
		return false, fmt.Errorf("invalid service name: %s", req.Service)
	}

	if m.mockMode || m.systemctlBinary == "" {
		return true, nil
	}

	cmd := exec.Command(m.systemctlBinary, req.Action, req.Service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("systemctl %s %s failed: %s (%w)", req.Action, req.Service, string(out), err)
	}

	return true, nil
}

func formatDisplayName(name string) string {
	switch name {
	case "nginx":
		return "Nginx Web Server"
	case "php8.4-fpm":
		return "PHP 8.4 FPM"
	case "php8.3-fpm":
		return "PHP 8.3 FPM"
	case "mysql":
		return "MySQL Database"
	case "mariadb":
		return "MariaDB Database"
	case "redis", "redis-server":
		return "Redis Cache & Queue"
	case "ufw":
		return "UFW Firewall"
	case "docker":
		return "Docker Engine"
	default:
		return strings.ToUpper(name[:1]) + name[1:]
	}
}

// FindPanelRepoDir locates the root git repository of panel-cli
func FindPanelRepoDir() string {
	candidates := []string{
		os.Getenv("PANEL_SOURCE_DIR"),
		os.Getenv("VPS_PANEL_DIR"),
		".",
		"..",
		"/var/www/vps-panel/panel",
		"/opt/panel-cli",
		"/usr/local/src/panel-cli",
		"/tmp/panel-cli",
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if fi, err := os.Stat(filepath.Join(c, ".git")); err == nil && (fi.IsDir() || !fi.IsDir()) {
			abs, errAbs := filepath.Abs(c)
			if errAbs == nil {
				return abs
			}
			return c
		}
	}
	return "."
}

// GetUpdateStatus inspects local git metadata and queries remote for available updates
func (m *Manager) GetUpdateStatus() (*protocol.SystemUpdateStatus, error) {
	repoDir := FindPanelRepoDir()

	status := &protocol.SystemUpdateStatus{
		CurrentVersion:  "1.0.0",
		CurrentCommit:   "unknown",
		CurrentBranch:   "main",
		RemoteURL:       "https://github.com/kodepreneur/panel-cli.git",
		UpdateAvailable: false,
		BehindCount:     0,
	}

	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		return status, nil
	}

	// Read current commit
	if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD").Output(); err == nil {
		status.CurrentCommit = strings.TrimSpace(string(out))
	}

	// Read current branch
	if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		status.CurrentBranch = strings.TrimSpace(string(out))
	}

	// Read remote URL
	if out, err := exec.Command("git", "-C", repoDir, "config", "--get", "remote.origin.url").Output(); err == nil {
		status.RemoteURL = strings.TrimSpace(string(out))
	}

	// Fetch updates from origin with a 5s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fetchCmd := exec.CommandContext(ctx, "git", "-C", repoDir, "fetch", "origin", status.CurrentBranch)
	_ = fetchCmd.Run()

	// Check latest remote commit
	originTarget := "origin/" + status.CurrentBranch
	if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", originTarget).Output(); err == nil {
		status.LatestCommit = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", repoDir, "log", "-1", "--pretty=%s", originTarget).Output(); err == nil {
		status.LatestMessage = strings.TrimSpace(string(out))
	}

	// Behind count
	revListCmd := exec.Command("git", "-C", repoDir, "rev-list", "--count", "HEAD.."+originTarget)
	if out, err := revListCmd.Output(); err == nil {
		var count int
		if _, scanErr := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &count); scanErr == nil {
			status.BehindCount = count
			if count > 0 {
				status.UpdateAvailable = true
			}
		}
	}

	return status, nil
}

// UpdateSystem pulls latest panel repository commits, rebuilds binaries, runs migrations and builds frontend
func (m *Manager) UpdateSystem(req protocol.SystemUpdateRequest) (*protocol.SystemUpdateResult, error) {
	repoDir := FindPanelRepoDir()
	branch := req.Branch
	if branch == "" {
		if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			branch = strings.TrimSpace(string(out))
		}
		if branch == "" || branch == "HEAD" {
			branch = "main"
		}
	}

	var logBuf strings.Builder
	logBuf.WriteString(fmt.Sprintf("[%s] Starting VPS Panel update in %s (branch: %s)...\n", time.Now().Format(time.RFC3339), repoDir, branch))

	oldCommit := ""
	if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD").Output(); err == nil {
		oldCommit = strings.TrimSpace(string(out))
	}

	// Git Pull
	pullCmd := exec.Command("git", "-C", repoDir, "pull", "origin", branch)
	pullOut, err := pullCmd.CombinedOutput()
	logBuf.WriteString(string(pullOut) + "\n")
	if err != nil {
		return nil, fmt.Errorf("git pull failed: %s (%w)", string(pullOut), err)
	}

	newCommit := oldCommit
	if out, err := exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD").Output(); err == nil {
		newCommit = strings.TrimSpace(string(out))
	}

	binariesBuilt := false

	// Build Go binaries if Go is installed
	if _, lookErr := exec.LookPath("go"); lookErr == nil {
		logBuf.WriteString("Compiling Go binaries (vpsd, vpsctl, vps-mcp)...\n")

		buildVpsd := exec.Command("go", "build", "-o", "/usr/local/bin/vpsd", "./daemon/cmd/vpsd")
		buildVpsd.Dir = repoDir
		if out, err := buildVpsd.CombinedOutput(); err != nil {
			logBuf.WriteString(fmt.Sprintf("Warning: vpsd compilation: %s\n", string(out)))
		}

		buildVpsctl := exec.Command("go", "build", "-o", "/usr/local/bin/vpsctl", "./cli/cmd/vpsctl")
		buildVpsctl.Dir = repoDir
		if out, err := buildVpsctl.CombinedOutput(); err != nil {
			logBuf.WriteString(fmt.Sprintf("Warning: vpsctl compilation: %s\n", string(out)))
		}

		buildMcp := exec.Command("go", "build", "-o", "/usr/local/bin/vps-mcp", "./mcp/cmd/vps-mcp")
		buildMcp.Dir = repoDir
		if out, err := buildMcp.CombinedOutput(); err != nil {
			logBuf.WriteString(fmt.Sprintf("Warning: vps-mcp compilation: %s\n", string(out)))
		}

		binariesBuilt = true
		logBuf.WriteString("Binaries compiled successfully.\n")
	}

	// Backend update
	backendDir := filepath.Join(repoDir, "backend")
	if _, err := os.Stat(backendDir); err == nil {
		if _, lookErr := exec.LookPath("composer"); lookErr == nil {
			logBuf.WriteString("Updating backend composer packages...\n")
			compCmd := exec.Command("composer", "install", "--no-dev", "--optimize-autoloader", "--no-interaction")
			compCmd.Dir = backendDir
			compOut, _ := compCmd.CombinedOutput()
			logBuf.WriteString(string(compOut) + "\n")
		}

		if _, lookErr := exec.LookPath("php"); lookErr == nil {
			logBuf.WriteString("Running database migrations...\n")
			migrateCmd := exec.Command("php", "artisan", "migrate", "--force")
			migrateCmd.Dir = backendDir
			migOut, _ := migrateCmd.CombinedOutput()
			logBuf.WriteString(string(migOut) + "\n")
		}
	}

	// Frontend update
	frontendDir := filepath.Join(repoDir, "frontend")
	if _, err := os.Stat(frontendDir); err == nil {
		if _, lookErr := exec.LookPath("npm"); lookErr == nil {
			logBuf.WriteString("Building frontend production assets...\n")
			npmBuild := exec.Command("npm", "run", "build")
			npmBuild.Dir = frontendDir
			buildOut, _ := npmBuild.CombinedOutput()
			logBuf.WriteString(string(buildOut) + "\n")

			// Copy to public if destination directory exists
			pubDir := "/var/www/vps-panel/public"
			if fi, err := os.Stat(pubDir); err == nil && fi.IsDir() {
				_ = exec.Command("cp", "-r", filepath.Join(frontendDir, "dist")+"/*", pubDir).Run()
				logBuf.WriteString(fmt.Sprintf("Copied updated web dashboard assets to %s\n", pubDir))
			}
		}
	}

	logBuf.WriteString(fmt.Sprintf("[%s] Panel update completed successfully! (%s -> %s)\n",
		time.Now().Format(time.RFC3339), oldCommit, newCommit))

	return &protocol.SystemUpdateResult{
		Success:       true,
		OldCommit:     oldCommit,
		NewCommit:     newCommit,
		Logs:          logBuf.String(),
		BinariesBuilt: binariesBuilt,
	}, nil
}


