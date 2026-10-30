package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func handleGithub(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"repositories"}
	}

	sub := args[0]

	switch sub {
	case "connect":
		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]string{
					"connect_url": "/dashboard/github/connect",
					"app_name":    "VPS Panel GitHub App",
				},
			})
			return
		}
		fmt.Println("To connect GitHub, visit the Web Dashboard at: /dashboard/github/connect")
		fmt.Println("Or configure personal access token in Settings.")

	case "repositories":
		repos := []map[string]string{
			{"name": "kodepreneur/landing-page", "private": "false", "default_branch": "main"},
			{"name": "kodepreneur/api-service", "private": "true", "default_branch": "main"},
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   repos,
			})
			return
		}

		fmt.Println("=== GITHUB REPOSITORIES ===")
		fmt.Printf("%-32s %-10s %s\n", "REPOSITORY", "VISIBILITY", "BRANCH")
		fmt.Println("--------------------------------------------------")
		for _, r := range repos {
			vis := "public"
			if r["private"] == "true" {
				vis = "private"
			}
			fmt.Printf("%-32s %-10s %s\n", r["name"], vis, r["default_branch"])
		}

	case "pull":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl github pull <project_slug>"), 1)
			return
		}
		slug := args[1]
		projDir := filepath.Join("/var/www/vps-panel/projects", slug, "current")

		if _, err := os.Stat(projDir); os.IsNotExist(err) {
			printError(cfg, fmt.Errorf("project '%s' current release directory not found at %s", slug, projDir), 1)
			return
		}

		cmd := exec.Command("git", "-C", projDir, "pull")
		out, err := cmd.CombinedOutput()
		outputStr := string(out)

		if err != nil {
			printError(cfg, fmt.Errorf("git pull failed: %s (%w)", outputStr, err), 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"slug":   slug,
					"output": outputStr,
				},
			})
			return
		}

		fmt.Printf("Git pull succeeded for '%s':\n%s\n", slug, outputStr)

	default:
		printError(cfg, fmt.Errorf("unknown github subcommand: %s", sub), 1)
	}
}

