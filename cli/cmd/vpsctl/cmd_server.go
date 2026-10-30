package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

func handleServer(cfg Config, args []string) {
	if len(args) == 0 {
		args = []string{"status"}
	}

	sub := args[0]
	client := getRPCClient(cfg)

	switch sub {
	case "status":
		overview, err := client.GetSystemOverview()
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   overview,
			})
			return
		}

		fmt.Println("=== SERVER OVERVIEW ===")
		fmt.Printf("Hostname:    %s\n", overview.OS.Hostname)
		fmt.Printf("OS:          %s (%s, %s)\n", overview.OS.Version, overview.OS.Kernel, overview.OS.Arch)
		fmt.Printf("Uptime:      %d hours %d mins\n", overview.UptimeSeconds/3600, (overview.UptimeSeconds%3600)/60)
		fmt.Printf("Load Avg:    %.2f, %.2f, %.2f\n", overview.LoadAvg[0], overview.LoadAvg[1], overview.LoadAvg[2])
		fmt.Printf("CPU:         %s (%d cores) - %.1f%% used\n", overview.CPU.Model, overview.CPU.Cores, overview.CPU.UsagePercent)
		fmt.Printf("Memory:      %.1f GB / %.1f GB (%.1f%% used)\n",
			float64(overview.Memory.UsedBytes)/(1024*1024*1024),
			float64(overview.Memory.TotalBytes)/(1024*1024*1024),
			overview.Memory.UsagePercent)
		if len(overview.Disks) > 0 {
			d := overview.Disks[0]
			fmt.Printf("Disk (/):    %.1f GB / %.1f GB (%.1f%% used)\n",
				float64(d.UsedBytes)/(1024*1024*1024),
				float64(d.TotalBytes)/(1024*1024*1024),
				d.UsagePercent)
		}

	case "resources":
		overview, err := client.GetSystemOverview()
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"cpu":     overview.CPU,
					"memory":  overview.Memory,
					"disks":   overview.Disks,
					"network": overview.Network,
					"load":    overview.LoadAvg,
				},
			})
			return
		}

		fmt.Println("=== SYSTEM RESOURCES ===")
		fmt.Printf("CPU Usage:    [%.1f%%] %s (%d cores)\n", overview.CPU.UsagePercent, overview.CPU.Model, overview.CPU.Cores)
		fmt.Printf("RAM Usage:    [%.1f%%] %.2f GB used of %.2f GB\n",
			overview.Memory.UsagePercent,
			float64(overview.Memory.UsedBytes)/(1024*1024*1024),
			float64(overview.Memory.TotalBytes)/(1024*1024*1024))
		fmt.Printf("Swap Usage:   %.2f GB used of %.2f GB\n",
			float64(overview.Memory.SwapUsedBytes)/(1024*1024*1024),
			float64(overview.Memory.SwapTotalBytes)/(1024*1024*1024))
		for _, d := range overview.Disks {
			fmt.Printf("Disk (%s):    [%.1f%%] %.2f GB used of %.2f GB\n",
				d.Mount, d.UsagePercent,
				float64(d.UsedBytes)/(1024*1024*1024),
				float64(d.TotalBytes)/(1024*1024*1024))
		}

	case "services":
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

	case "update":
		checkOnly := false
		branch := ""

		for _, a := range os.Args {
			if a == "--check" {
				checkOnly = true
			}
			if strings.HasPrefix(a, "--branch=") {
				branch = strings.TrimPrefix(a, "--branch=")
			}
		}

		if checkOnly {
			status, err := client.GetSystemUpdateStatus()
			if err != nil {
				printError(cfg, fmt.Errorf("failed to check update status: %w", err), 1)
				return
			}

			if cfg.JSONOutput {
				printJSON(map[string]interface{}{
					"status": "success",
					"data":   status,
				})
				return
			}

			fmt.Println("=== VPS PANEL UPDATE STATUS ===")
			fmt.Printf("Current Version:  %s\n", status.CurrentVersion)
			fmt.Printf("Current Commit:   %s\n", status.CurrentCommit)
			fmt.Printf("Active Branch:    %s\n", status.CurrentBranch)
			fmt.Printf("Repository URL:   %s\n", status.RemoteURL)
			if status.UpdateAvailable {
				fmt.Printf("Update Status:    UPDATE AVAILABLE (%d commits behind origin/%s)\n", status.BehindCount, status.CurrentBranch)
				fmt.Printf("Latest Commit:    %s (\"%s\")\n", status.LatestCommit, status.LatestMessage)
				fmt.Println("\nRun 'vpsctl server update' or 'vpsctl update' to apply updates.")
			} else {
				fmt.Println("Update Status:    Up to date with remote repository.")
			}
			return
		}

		fmt.Println("Pulling latest panel updates from repository and compiling binaries...")
		result, err := client.UpdateSystem(protocol.SystemUpdateRequest{Branch: branch})
		if err != nil {
			printError(cfg, fmt.Errorf("panel update failed: %w", err), 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   result,
			})
			return
		}

		fmt.Println("=== VPS PANEL UPDATE COMPLETE ===")
		fmt.Printf("Previous Commit:  %s\n", result.OldCommit)
		fmt.Printf("Updated Commit:   %s\n", result.NewCommit)
		binStatus := "No"
		if result.BinariesBuilt {
			binStatus = "Yes (vpsd, vpsctl, vps-mcp recompiled)"
		}
		fmt.Printf("Binaries Built:   %s\n", binStatus)
		fmt.Println("Status:           SUCCESS")
		if result.Logs != "" {
			fmt.Println("\n--- Update Execution Logs ---")
			fmt.Println(strings.TrimSpace(result.Logs))
		}


	default:
		printError(cfg, fmt.Errorf("unknown server subcommand: %s", sub), 1)
	}
}
