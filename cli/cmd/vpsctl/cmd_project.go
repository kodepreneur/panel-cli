package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

func handleProject(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	sub := args[0]
	client := getRPCClient(cfg)

	switch sub {
	case "list":
		// Check local projects directory
		baseDir := "/var/www/vps-panel/projects"
		if envDir := os.Getenv("VPSD_PROJECTS_DIR"); envDir != "" {
			baseDir = envDir
		} else if _, err := os.Stat(baseDir); err != nil {
			if _, err2 := os.Stat("storage/projects"); err2 == nil {
				baseDir = "storage/projects"
			}
		}

		entries, err := os.ReadDir(baseDir)
		var projects []map[string]interface{}

		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					slug := entry.Name()
					currentLink := filepath.Join(baseDir, slug, "current")
					target, _ := os.Readlink(currentLink)
					activeRelease := filepath.Base(target)
					if activeRelease == "." || activeRelease == "" {
						activeRelease = "none"
					}

					projects = append(projects, map[string]interface{}{
						"slug":           slug,
						"path":           filepath.Join(baseDir, slug),
						"active_release": activeRelease,
						"status":         "active",
					})
				}
			}
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   projects,
			})
			return
		}

		fmt.Println("=== HOSTED PROJECTS ===")
		if len(projects) == 0 {
			fmt.Println("No projects provisioned yet. Use 'vpsctl project create' to create one.")
			return
		}

		fmt.Printf("%-24s %-20s %s\n", "PROJECT SLUG", "CURRENT RELEASE", "STATUS")
		fmt.Println(strings.Repeat("-", 60))
		for _, p := range projects {
			fmt.Printf("%-24s %-20s %s\n", p["slug"], p["active_release"], p["status"])
		}

	case "create":
		if len(args) < 4 {
			printError(cfg, fmt.Errorf("usage: vpsctl project create <name> <domain> <type> (e.g. laravel, nodejs, static)"), 1)
			return
		}
		name := args[1]
		domain := args[2]
		pType := strings.ToLower(args[3])
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

		// Prepare release folder
		releaseTag := time.Now().Format("20060102_150405")
		prepRes, err := client.PrepareRelease(slug, releaseTag)
		if err != nil {
			printError(cfg, fmt.Errorf("failed to prepare release directory: %w", err), 1)
			return
		}

		// Activate initial release
		_, err = client.ActivateRelease(slug, releaseTag)
		if err != nil {
			printError(cfg, fmt.Errorf("failed to activate initial release: %w", err), 1)
			return
		}

		// Apply Nginx Virtual Host
		vhostCfg := protocol.VhostConfig{
			Slug:          slug,
			PrimaryDomain: domain,
			ProjectType:   pType,
			RootPath:      filepath.Join("/var/www/vps-panel/projects", slug, "current"),
			SSLEnabled:    false,
		}
		if pType == "laravel" {
			vhostCfg.PublicDir = "public"
			vhostCfg.PhpVersion = "8.4"
		} else if pType == "nodejs" || pType == "nextjs" {
			vhostCfg.UpstreamPort = 3000
		} else if pType == "react_vite" || pType == "static" {
			vhostCfg.PublicDir = "dist"
		}

		vRes, err := client.ApplyVhost(vhostCfg)
		if err != nil {
			printError(cfg, fmt.Errorf("failed to apply vhost: %w", err), 1)
			return
		}

		resData := map[string]interface{}{
			"slug":         slug,
			"name":         name,
			"domain":       domain,
			"type":         pType,
			"release_tag":  releaseTag,
			"release_path": prepRes.ReleasePath,
			"vhost":        vRes,
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   resData,
			})
			return
		}

		fmt.Printf("Project '%s' created successfully!\n", name)
		fmt.Printf("  Domain:       %s\n", domain)
		fmt.Printf("  Slug:         %s\n", slug)
		fmt.Printf("  Type:         %s\n", pType)
		fmt.Printf("  Directory:    %s\n", prepRes.ReleasePath)
		fmt.Printf("  Nginx Config: %s\n", vRes.ConfigPath)
		fmt.Println("\nRun 'vpsctl ssl install " + domain + "' to enable HTTPS.")

	case "deploy":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl project deploy <slug>"), 1)
			return
		}
		slug := args[1]
		releaseTag := time.Now().Format("20060102_150405")

		prepRes, err := client.PrepareRelease(slug, releaseTag)
		if err != nil {
			printError(cfg, fmt.Errorf("failed to prepare release: %w", err), 1)
			return
		}

		// Activate release atomically
		actRes, err := client.ActivateRelease(slug, releaseTag)
		if err != nil {
			printError(cfg, fmt.Errorf("failed to activate release: %w", err), 1)
			return
		}

		// Cleanup older releases (keep 5)
		removed, _ := getRPCClient(cfg).CleanupOldReleases(slug, 5)

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"slug":             slug,
					"release_tag":      releaseTag,
					"release_path":     prepRes.ReleasePath,
					"current_symlink":  actRes.CurrentSymlink,
					"pruned_releases":  removed,
				},
			})
			return
		}

		fmt.Printf("Deployment succeeded for '%s'!\n", slug)
		fmt.Printf("  Active Release:  %s\n", releaseTag)
		fmt.Printf("  Release Path:    %s\n", prepRes.ReleasePath)
		if len(removed) > 0 {
			fmt.Printf("  Pruned Releases: %s\n", strings.Join(removed, ", "))
		}

	case "rollback":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl project rollback <slug> [target_release_tag]"), 1)
			return
		}
		slug := args[1]
		targetTag := ""
		if len(args) >= 3 {
			targetTag = args[2]
		}

		_, err := client.RollbackRelease(slug, targetTag)
		if err != nil {
			printError(cfg, fmt.Errorf("rollback failed: %w", err), 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"slug":       slug,
					"rolled_to":  targetTag,
				},
			})
			return
		}

		if targetTag != "" {
			fmt.Printf("Successfully rolled back '%s' to release '%s'!\n", slug, targetTag)
		} else {
			fmt.Printf("Successfully rolled back '%s' to previous release!\n", slug)
		}

	case "delete":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl project delete <slug>"), 1)
			return
		}
		slug := args[1]

		// Remove vhost
		_, _ = client.RemoveVhost(slug)

		// Remove project directory
		projDir := filepath.Join("/var/www/vps-panel/projects", slug)
		_ = os.RemoveAll(projDir)

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"slug":    slug,
					"deleted": true,
				},
			})
			return
		}

		fmt.Printf("Project '%s' and its virtual host deleted.\n", slug)

	default:
		printError(cfg, fmt.Errorf("unknown project subcommand: %s", sub), 1)
	}
}
