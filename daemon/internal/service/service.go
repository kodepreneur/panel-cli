package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

