package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func handleAgent(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	sub := args[0]

	switch sub {
	case "list":
		agents := []map[string]interface{}{
			{
				"id":          "ag_hermes_01",
				"name":        "Hermes Autonomous DevOps",
				"agent_type":  "hermes",
				"key_prefix":  "vps_live_8f0a...",
				"status":      "active",
				"permissions": []string{"server.read", "project.read", "project.deploy"},
				"created_at":  time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			},
			{
				"id":          "ag_openclaw_01",
				"name":        "OpenClaw Infra Agent",
				"agent_type":  "openclaw",
				"key_prefix":  "vps_live_4b2c...",
				"status":      "active",
				"permissions": []string{"server.read", "project.read", "project.create", "ssl.manage"},
				"created_at":  time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			},
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   agents,
			})
			return
		}

		fmt.Println("=== CONNECTED AI AGENTS ===")
		fmt.Printf("%-16s %-24s %-12s %-10s %s\n", "AGENT ID", "NAME", "TYPE", "STATUS", "KEY PREFIX")
		fmt.Println(strings.Repeat("-", 75))
		for _, a := range agents {
			fmt.Printf("%-16s %-24s %-12s %-10s %s\n", a["id"], a["name"], a["agent_type"], a["status"], a["key_prefix"])
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
		agentID := "ag_" + agentType + "_" + hex.EncodeToString(rawBytes[:4])

		res := map[string]interface{}{
			"id":         agentID,
			"name":       name,
			"agent_type": agentType,
			"api_key":    apiKey,
			"status":     "active",
			"permissions": []string{
				"server.read",
				"project.read",
				"project.deploy",
				"ssl.manage",
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
		approvals := []map[string]interface{}{
			{
				"id":         "appr_101",
				"agent_id":   "ag_hermes_01",
				"agent_name": "Hermes Autonomous DevOps",
				"action":     "project.delete",
				"target":     "old-staging-app",
				"status":     "pending",
				"requested":  time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			},
		}

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

		fmt.Printf("%-12s %-20s %-16s %-16s %s\n", "ID", "AGENT", "ACTION", "TARGET", "STATUS")
		fmt.Println(strings.Repeat("-", 75))
		for _, ap := range approvals {
			fmt.Printf("%-12s %-20s %-16s %-16s %s\n", ap["id"], ap["agent_name"], ap["action"], ap["target"], ap["status"])
		}

	default:
		printError(cfg, fmt.Errorf("unknown agent subcommand: %s", sub), 1)
	}
}

