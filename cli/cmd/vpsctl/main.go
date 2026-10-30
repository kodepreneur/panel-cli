package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kodepreneur/vps-panel/pkg/client"
)

var version = "1.0.0"

type Config struct {
	SocketPath string
	APIURL     string
	APIToken   string
	JSONOutput bool
}

func main() {
	cfg := parseGlobalFlags()

	nonFlagArgs := filterFlags(os.Args[1:])
	if len(nonFlagArgs) == 0 {
		printHelp()
		os.Exit(0)
	}

	cmd := nonFlagArgs[0]
	args := nonFlagArgs[1:]

	// Handle version
	if cmd == "version" || cmd == "-v" {

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{"version": version})
		} else {
			fmt.Printf("vpsctl version %s\n", version)
		}
		return
	}

	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		printHelp()
		return
	}

	switch cmd {
	case "server":
		handleServer(cfg, args)
	case "project":
		handleProject(cfg, args)
	case "update":
		handleServer(cfg, append([]string{"update"}, args...))
	case "ssl":
		handleSSL(cfg, args)
	case "service":
		handleService(cfg, args)
	case "database":
		handleDatabase(cfg, args)
	case "firewall":
		handleFirewall(cfg, args)
	case "github":
		handleGithub(cfg, args)
	case "agent":
		handleAgent(cfg, args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'vpsctl help' for usage.\n", cmd)
		os.Exit(1)
	}
}

func parseGlobalFlags() Config {
	sock := os.Getenv("VPSD_SOCKET_PATH")
	if sock == "" {
		sock = "/run/vpsd/vpsd.sock"
	}
	// Fallback to /tmp/vpsd.sock if /run/vpsd does not exist and /tmp exists
	if _, err := os.Stat("/run/vpsd/vpsd.sock"); err != nil {
		if _, err2 := os.Stat("/tmp/vpsd.sock"); err2 == nil {
			sock = "/tmp/vpsd.sock"
		}
	}

	cfg := Config{
		SocketPath: sock,
		APIURL:     os.Getenv("VPS_API_URL"),
		APIToken:   os.Getenv("VPS_API_TOKEN"),
		JSONOutput: false,
	}

	for _, arg := range os.Args {
		if arg == "--json" {
			cfg.JSONOutput = true
		}
		if strings.HasPrefix(arg, "--socket=") {
			cfg.SocketPath = strings.TrimPrefix(arg, "--socket=")
		}
		if strings.HasPrefix(arg, "--api-url=") {
			cfg.APIURL = strings.TrimPrefix(arg, "--api-url=")
		}
		if strings.HasPrefix(arg, "--api-token=") {
			cfg.APIToken = strings.TrimPrefix(arg, "--api-token=")
		}
	}

	return cfg
}

func filterFlags(args []string) []string {
	var result []string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			result = append(result, arg)
		}
	}
	return result
}

func getRPCClient(cfg Config) *client.Client {
	return client.New(cfg.SocketPath)
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"status":"error","error":"%s"}`+"\n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printError(cfg Config, err error, code int) {
	if cfg.JSONOutput {
		printJSON(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
			"code":   code,
		})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
	os.Exit(code)
}

func printHelp() {
	fmt.Printf(`vpsctl - AI-Native VPS Management Panel CLI (v%s)

Usage:
  vpsctl <command> <subcommand> [arguments] [--flags]

Commands:
  server
    status                 Display overall server health, OS info and uptime
    resources              Display real-time CPU, RAM, Disk, and Load metrics
    services               List tracked system services
    update [--check]       Check or pull updates for VPS Panel from repository

  project
    list                   List all hosted websites and applications
    create <name> <domain> <type>
                           Provision a new project (laravel, nodejs, static, etc.)
    update <slug>          Update project from its git repository with atomic release
    deploy <slug>          Trigger atomic release deployment
    rollback <slug> [tag]  Roll back project to previous or specified release
    logs <slug>            Display recent project deployment logs
    delete <slug>          Remove a project and its virtual host

  update                   Check or pull updates for VPS Panel from repository


  ssl
    status <domain>        Check SSL certificate status and expiration
    install <domain>       Issue Let's Encrypt certificate for a domain
    renew <domain>         Force renew SSL certificate

  service
    list                   List system services status
    restart <service>      Restart a system service (e.g. nginx, mysql, redis)
    reload <service>       Reload a system service configuration

  database
    list                   List provisioned databases
    create <db> <user> <pass> [engine]
                           Create a new database and user (mysql or postgres)

  firewall
    status                 Display UFW firewall status and active rules
    allow <port> [proto]   Allow traffic through firewall (e.g. 80, 443, 22)
    delete <rule_num>      Delete firewall rule by number

  github
    connect                Connect GitHub account / App
    repositories           Browse accessible GitHub repositories
    pull <slug>            Pull latest commit from connected git repo

  agent
    list                   List AI agents and their credentials
    create <name> <type>   Generate API credentials for an AI agent
    revoke <agent_id>      Revoke agent credentials immediately
    approvals              List or resolve pending human approval requests

Flags:
  --json                   Output machine-readable JSON (ideal for AI Agents)
  --socket=<path>          Path to vpsd Unix socket (default: /run/vpsd/vpsd.sock)
  --api-url=<url>          Remote Panel API URL
  --api-token=<token>      API Bearer Token
`, version)
}
