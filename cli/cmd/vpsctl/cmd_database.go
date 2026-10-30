package main

import (
	"fmt"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

func handleDatabase(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	sub := args[0]
	client := getRPCClient(cfg)

	switch sub {
	case "list":
		// Mock/read-only list when local socket
		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": []map[string]string{
					{"name": "vps_panel", "engine": "mysql", "charset": "utf8mb4"},
				},
			})
			return
		}

		fmt.Println("=== DATABASES ===")
		fmt.Printf("%-20s %-12s %s\n", "NAME", "ENGINE", "CHARSET")
		fmt.Println("----------------------------------------")
		fmt.Printf("%-20s %-12s %s\n", "vps_panel", "mysql", "utf8mb4")

	case "create":
		if len(args) < 4 {
			printError(cfg, fmt.Errorf("usage: vpsctl database create <db_name> <username> <password> [engine]"), 1)
			return
		}

		dbName := args[1]
		username := args[2]
		password := args[3]
		engine := "mysql"
		if len(args) >= 5 {
			engine = args[4]
		}

		req := protocol.CreateDatabaseRequest{
			Engine:   engine,
			Database: dbName,
			Username: username,
			Password: password,
		}

		res, err := client.CreateDatabase(req)
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if !res.Success {
			printError(cfg, fmt.Errorf("database creation failed: %s", res.Error), 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   res,
			})
			return
		}

		fmt.Printf("Database '%s' created successfully!\n", dbName)
		fmt.Printf("  Engine:   %s\n", engine)
		fmt.Printf("  User:     %s\n", username)

	default:
		printError(cfg, fmt.Errorf("unknown database subcommand: %s", sub), 1)
	}
}

