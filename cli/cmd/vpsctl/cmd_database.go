package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		databases := getDatabasesFromDB()

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   databases,
			})
			return
		}

		fmt.Println("=== DATABASES ===")
		if len(databases) == 0 {
			fmt.Println("No databases configured yet.")
			fmt.Println("Use 'vpsctl database create <db_name> <username> <password> [engine]' to create one.")
			return
		}

		fmt.Printf("%-24s %-12s %s\n", "NAME", "ENGINE", "CHARSET")
		fmt.Println(strings.Repeat("-", 48))
		for _, db := range databases {
			fmt.Printf("%-24s %-12s %s\n", db["name"], db["engine"], db["charset"])
		}

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

		// Record in database
		dbPath := findDBPath()
		if dbPath != "" {
			insertCmd := fmt.Sprintf("INSERT OR REPLACE INTO databases (id, engine, name, username, charset, collation, status, created_at, updated_at) VALUES ('db_%s', '%s', '%s', '%s', 'utf8mb4', 'utf8mb4_unicode_ci', 'active', datetime('now'), datetime('now'));",
				dbName, engine, dbName, username)
			_ = exec.Command("sqlite3", dbPath, insertCmd).Run()
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

func getDatabasesFromDB() []map[string]string {
	dbPaths := []string{
		"/var/www/vps-panel/panel/database/database.sqlite",
		"backend/database/database.sqlite",
		"../backend/database/database.sqlite",
	}
	var dbPath string
	for _, p := range dbPaths {
		if _, err := os.Stat(p); err == nil {
			dbPath = p
			break
		}
	}
	if dbPath == "" {
		return []map[string]string{}
	}

	out, err := exec.Command("sqlite3", dbPath, "SELECT name, engine, charset FROM databases;").CombinedOutput()
	if err != nil {
		return []map[string]string{}
	}

	var list []map[string]string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			list = append(list, map[string]string{
				"name":    parts[0],
				"engine":  parts[1],
				"charset": parts[2],
			})
		}
	}
	return list
}
