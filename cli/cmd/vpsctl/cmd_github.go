package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
		fmt.Println("Or configure personal access token via GITHUB_TOKEN in your environment.")

	case "repositories":
		repos := fetchGitHubRepositories()

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   repos,
			})
			return
		}

		fmt.Println("=== GITHUB REPOSITORIES ===")
		if len(repos) == 0 {
			fmt.Println("No GitHub repositories connected.")
			fmt.Println("Link a GitHub repository using 'vpsctl github connect' or set GITHUB_TOKEN.")
			return
		}

		fmt.Printf("%-36s %-12s %s\n", "REPOSITORY", "VISIBILITY", "BRANCH")
		fmt.Println(strings.Repeat("-", 60))
		for _, r := range repos {
			vis := "public"
			if r["private"] == "true" {
				vis = "private"
			}
			fmt.Printf("%-36s %-12s %s\n", r["name"], vis, r["default_branch"])
		}

	case "pull":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl github pull <project_slug>"), 1)
			return
		}
		slug := args[1]
		baseDir := "/var/www/vps-panel/projects"
		if envDir := os.Getenv("VPSD_PROJECTS_DIR"); envDir != "" {
			baseDir = envDir
		} else if _, err := os.Stat("storage/projects"); err == nil {
			baseDir = "storage/projects"
		}

		projDir := filepath.Join(baseDir, slug, "current")

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

func fetchGitHubRepositories() []map[string]string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return []map[string]string{}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/user/repos?per_page=30&sort=updated", nil)
	if err != nil {
		return []map[string]string{}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return []map[string]string{}
	}
	defer resp.Body.Close()

	var ghRepos []struct {
		FullName      string `json:"full_name"`
		Private       bool   `json:"private"`
		DefaultBranch string `json:"default_branch"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghRepos); err != nil {
		return []map[string]string{}
	}

	var results []map[string]string
	for _, r := range ghRepos {
		vis := "false"
		if r.Private {
			vis = "true"
		}
		results = append(results, map[string]string{
			"name":           r.FullName,
			"private":        vis,
			"default_branch": r.DefaultBranch,
		})
	}
	return results
}
