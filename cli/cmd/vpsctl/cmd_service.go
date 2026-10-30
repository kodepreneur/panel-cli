package main

import (
	"fmt"
	"strings"
)

func handleService(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	sub := args[0]
	client := getRPCClient(cfg)

	switch sub {
	case "list":
		statuses, err := client.GetServices(nil)
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   statuses,
			})
			return
		}

		fmt.Println("=== SYSTEM SERVICES ===")
		fmt.Printf("%-20s %-12s %s\n", "SERVICE", "STATUS", "DISPLAY NAME")
		fmt.Println(strings.Repeat("-", 50))
		for _, s := range statuses {
			stateTag := "[INACTIVE]"
			if s.Active {
				stateTag = "[RUNNING]"
			}
			fmt.Printf("%-20s %-12s %s\n", s.Name, stateTag, s.DisplayName)
		}

	case "restart", "start", "stop", "reload":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl service %s <service_name>", sub), 1)
			return
		}
		serviceName := args[1]

		success, err := client.ServiceAction(serviceName, sub)
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"service": serviceName,
					"action":  sub,
					"success": success,
				},
			})
			return
		}

		fmt.Printf("Service '%s' %sed successfully.\n", serviceName, sub)

	default:
		printError(cfg, fmt.Errorf("unknown service subcommand: %s", sub), 1)
	}
}
