package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

func handleFirewall(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"status"}
	}

	sub := args[0]
	client := getRPCClient(cfg)

	switch sub {
	case "status":
		status, err := client.GetFirewallStatus()
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   status,
			})
			return
		}

		fmt.Println("=== UFW FIREWALL STATUS ===")
		stateStr := "INACTIVE"
		if status.Enabled {
			stateStr = "ACTIVE"
		}
		fmt.Printf("Firewall: %s\n\n", stateStr)

		fmt.Printf("%-6s %-16s %-10s %s\n", "RULE", "PORT/PROTO", "ACTION", "FROM")
		fmt.Println(strings.Repeat("-", 48))
		for _, r := range status.Rules {
			fmt.Printf("[%-3d] %-16s %-10s %s\n", r.Number, r.Port, r.Action, r.From)
		}

	case "allow":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl firewall allow <port> [proto] (e.g. 80, 443, 22)"), 1)
			return
		}

		port := args[1]
		proto := "tcp"
		if len(args) >= 3 {
			proto = args[2]
		}

		success, err := client.AddFirewallRule(protocol.AddFirewallRuleRequest{
			Port:   port,
			Proto:  proto,
			Action: "allow",
		})
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"port":    port,
					"proto":   proto,
					"action":  "allow",
					"success": success,
				},
			})
			return
		}

		fmt.Printf("Firewall rule added: ALLOW %s/%s\n", port, proto)

	case "delete":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl firewall delete <rule_number>"), 1)
			return
		}

		num, err := strconv.Atoi(args[1])
		if err != nil {
			printError(cfg, fmt.Errorf("invalid rule number: %s", args[1]), 1)
			return
		}

		success, err := client.DeleteFirewallRule(num)
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"rule_number": num,
					"deleted":     success,
				},
			})
			return
		}

		fmt.Printf("Firewall rule #%d deleted.\n", num)

	default:
		printError(cfg, fmt.Errorf("unknown firewall subcommand: %s", sub), 1)
	}
}

