package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func handleAgent(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	sub := args[0]

	switch sub {
	case "list":
		agents := getAgentsFromDB()

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   agents,
			})
			return
		}

		fmt.Println("=== CONNECTED AI AGENTS ===")
		if len(agents) == 0 {
			fmt.Println("No connected AI agents found.")
			fmt.Println("Use 'vpsctl agent create <name> <type>' to provision an agent.")
			return
		}

		fmt.Printf("%-38s %-26s %-10s %-10s %s\n", "AGENT ID", "NAME", "TYPE", "STATUS", "KEY PREFIX")
		fmt.Println(strings.Repeat("-", 95))
		for _, a := range agents {
			fmt.Printf("%-38s %-26s %-10s %-10s %s\n", a["id"], a["name"], a["agent_type"], a["status"], a["key_prefix"])
		}

	case "create":
		if len(args) < 3 {
			printError(cfg, fmt.Errorf("usage: vpsctl agent create <name> <type> (e.g. hermes, openclaw, openhands, custom)"), 1)
			return
		}
		name := args[1]
		agentType := args[2]

		// Generate random 24-byte API key
		rawBytes := make([]byte, 24)
		_, _ = rand.Read(rawBytes)
		apiKey := "vps_live_" + hex.EncodeToString(rawBytes)
		keyPrefix := apiKey[:15] + "..."

		// If DB is available, insert directly
		dbPath := findDBPath()
		agentID := "ag_" + agentType + "_" + hex.EncodeToString(rawBytes[:4])
		if dbPath != "" {
			insertCmd := fmt.Sprintf("INSERT INTO ai_agents (id, name, agent_type, api_key_hash, api_key_prefix, permissions, is_active, created_at, updated_at) VALUES ('%s', '%s', '%s', '%s', '%s', '[\"*\"]', 1, datetime('now'), datetime('now'));",
				agentID, name, agentType, sha256Hex(apiKey), keyPrefix)
			_ = exec.Command("sqlite3", dbPath, insertCmd).Run()
		}

		res := map[string]interface{}{
			"id":         agentID,
			"name":       name,
			"agent_type": agentType,
			"api_key":    apiKey,
			"status":     "active",
			"permissions": []string{
				"*",
			},
			"created_at": time.Now().Format(time.RFC3339),
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   res,
			})
			return
		}

		fmt.Printf("AI Agent '%s' provisioned successfully!\n", name)
		fmt.Printf("  Agent ID:   %s\n", agentID)
		fmt.Printf("  Type:       %s\n", agentType)
		fmt.Printf("  API Key:    %s\n", apiKey)
		fmt.Println("\nIMPORTANT: Store this API key securely. It will not be shown again.")

	case "revoke":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl agent revoke <agent_id>"), 1)
			return
		}
		agentID := args[1]

		dbPath := findDBPath()
		if dbPath != "" {
			_ = exec.Command("sqlite3", dbPath, fmt.Sprintf("DELETE FROM ai_agents WHERE id='%s';", agentID)).Run()
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":      agentID,
					"revoked": true,
				},
			})
			return
		}

		fmt.Printf("Agent '%s' revoked successfully.\n", agentID)

	case "approvals":
		approvals := getPendingApprovalsFromDB()

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   approvals,
			})
			return
		}

		fmt.Println("=== PENDING AGENT APPROVALS ===")
		if len(approvals) == 0 {
			fmt.Println("No pending approval requests.")
			return
		}

		fmt.Printf("%-38s %-20s %-16s %-20s %s\n", "REQUEST ID", "AGENT", "ACTION", "TARGET", "STATUS")
		fmt.Println(strings.Repeat("-", 100))
		for _, appr := range approvals {
			fmt.Printf("%-38s %-20s %-16s %-20s %s\n", appr["id"], appr["agent_name"], appr["action"], appr["target"], appr["status"])
		}

	default:
		printError(cfg, fmt.Errorf("unknown agent subcommand: %s", sub), 1)
	}
}

func findDBPath() string {
	dbPaths := []string{
		"/var/www/vps-panel/panel/database/database.sqlite",
		"backend/database/database.sqlite",
		"../backend/database/database.sqlite",
	}
	for _, p := range dbPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func getAgentsFromDB() []map[string]interface{} {
	dbPath := findDBPath()
	if dbPath == "" {
		return []map[string]interface{}{}
	}

	out, err := exec.Command("sqlite3", dbPath, "SELECT id, name, agent_type, api_key_prefix, is_active FROM ai_agents;").CombinedOutput()
	if err != nil {
		return []map[string]interface{}{}
	}

	var list []map[string]interface{}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			status := "inactive"
			if parts[4] == "1" {
				status = "active"
			}
			list = append(list, map[string]interface{}{
				"id":         parts[0],
				"name":       parts[1],
				"agent_type": parts[2],
				"key_prefix": parts[3],
				"status":     status,
			})
		}
	}
	return list
}

func getPendingApprovalsFromDB() []map[string]interface{} {
	dbPath := findDBPath()
	if dbPath == "" {
		return []map[string]interface{}{}
	}

	out, err := exec.Command("sqlite3", dbPath, "SELECT a.id, a.agent_id, ag.name, a.action_type, a.resource_id, a.status, a.created_at FROM approval_requests a LEFT JOIN ai_agents ag ON a.agent_id = ag.id WHERE a.status='pending';").CombinedOutput()
	if err != nil {
		return []map[string]interface{}{}
	}

	var list []map[string]interface{}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 6 {
			agentName := parts[2]
			if agentName == "" {
				agentName = "AI Agent"
			}
			list = append(list, map[string]interface{}{
				"id":         parts[0],
				"agent_id":   parts[1],
				"agent_name": agentName,
				"action":     parts[3],
				"target":     parts[4],
				"status":     parts[5],
				"requested":  parts[6],
			})
		}
	}
	return list
}
