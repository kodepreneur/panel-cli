package firewall

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

type Manager struct {
	ufwBinary string
	mockMode  bool
}

func NewManager() *Manager {
	bin, err := exec.LookPath("ufw")
	mock := false
	if err != nil || os.Geteuid() != 0 {
		mock = true
	}
	return &Manager{
		ufwBinary: bin,
		mockMode:  mock,
	}
}

var ufwRuleRegex = regexp.MustCompile(`^\[\s*(\d+)\]\s+([^\s]+)\s+([A-Z]+)\s+IN\s+([^\s]+)`)

func (m *Manager) GetStatus() (*protocol.FirewallStatusResult, error) {
	if m.ufwBinary != "" && !m.mockMode {
		cmd := exec.Command(m.ufwBinary, "status", "numbered")
		out, err := cmd.CombinedOutput()
		if err == nil {
			result := &protocol.FirewallStatusResult{
				Rules: []protocol.FirewallRule{},
			}
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "Status: active") {
					result.Enabled = true
				}
				matches := ufwRuleRegex.FindStringSubmatch(line)
				if len(matches) == 5 {
					num, _ := strconv.Atoi(matches[1])
					proto := "any"
					if strings.Contains(matches[2], "/tcp") {
						proto = "tcp"
					} else if strings.Contains(matches[2], "/udp") {
						proto = "udp"
					}
					result.Rules = append(result.Rules, protocol.FirewallRule{
						Number: num,
						Port:   matches[2],
						Proto:  proto,
						Action: matches[3],
						From:   matches[4],
					})
				}
			}
			return result, nil
		}
	}

	// Check pfctl on macOS / BSD
	pfOut, pfErr := exec.Command("/sbin/pfctl", "-s", "info").CombinedOutput()
	pfEnabled := pfErr == nil && strings.Contains(string(pfOut), "Status: Enabled")

	return &protocol.FirewallStatusResult{
		Enabled:         pfEnabled,
		DefaultIncoming: "deny",
		Rules:           []protocol.FirewallRule{},
	}, nil
}

func (m *Manager) AddRule(req protocol.AddFirewallRuleRequest) (bool, error) {
	if req.Port == "" {
		return false, fmt.Errorf("port is required")
	}

	// Validate port string format (e.g., "80", "80/tcp", "8080")
	if strings.ContainsAny(req.Port, " ;&|<>$`\n\r") {
		return false, fmt.Errorf("invalid port specification: %s", req.Port)
	}

	action := strings.ToLower(req.Action)
	if action != "allow" && action != "deny" {
		action = "allow"
	}

	if m.mockMode || m.ufwBinary == "" {
		return true, nil
	}

	args := []string{action}
	if req.Proto != "" && req.Proto != "any" {
		args = append(args, fmt.Sprintf("%s/%s", req.Port, req.Proto))
	} else {
		args = append(args, req.Port)
	}

	if req.FromIP != "" {
		args = append(args, "from", req.FromIP)
	}

	cmd := exec.Command(m.ufwBinary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("ufw add rule failed: %s (%w)", string(out), err)
	}

	return true, nil
}

func (m *Manager) DeleteRule(ruleNumber int) (bool, error) {
	if ruleNumber <= 0 {
		return false, fmt.Errorf("invalid rule number: %d", ruleNumber)
	}

	if m.mockMode || m.ufwBinary == "" {
		return true, nil
	}

	cmd := exec.Command(m.ufwBinary, "--force", "delete", strconv.Itoa(ruleNumber))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("ufw delete rule failed: %s (%w)", string(out), err)
	}

	return true, nil
}

