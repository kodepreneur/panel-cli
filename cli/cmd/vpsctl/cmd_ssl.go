package main

import (
	"fmt"
	"strings"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

func handleSSL(cfg Config, args []string) {
	if len(args) == 0 {
		printError(cfg, fmt.Errorf("usage: vpsctl ssl <status|install|renew> <domain>"), 1)
		return
	}

	sub := args[0]
	client := getRPCClient(cfg)

	switch sub {
	case "status":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl ssl status <domain>"), 1)
			return
		}
		domain := args[1]
		status, err := client.GetSSLStatus(domain)
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

		fmt.Printf("=== SSL CERTIFICATE STATUS: %s ===\n", domain)
		if !status.Exists {
			fmt.Println("No SSL certificate installed for this domain.")
			return
		}

		fmt.Printf("Valid:        %v\n", status.Valid)
		fmt.Printf("Issuer:       %s\n", status.Issuer)
		fmt.Printf("Domains:      %s\n", strings.Join(status.Domains, ", "))
		fmt.Printf("Expires:      %s (%d days left)\n", status.ExpiresAt.Format("2006-01-02 15:04:05 UTC"), status.DaysLeft)
		fmt.Printf("Cert Path:    %s\n", status.CertPath)

	case "install":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl ssl install <domain> [email]"), 1)
			return
		}
		domain := args[1]
		email := "admin@" + domain
		if len(args) >= 3 {
			email = args[2]
		}

		req := protocol.IssueSSLRequest{
			Domain:        domain,
			Email:         email,
			ChallengeType: "http-01",
		}

		res, err := client.IssueSSL(req)
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if !res.Success {
			printError(cfg, fmt.Errorf("SSL issuance failed: %s", res.Error), 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data":   res,
			})
			return
		}

		fmt.Printf("SSL certificate issued successfully for %s!\n", domain)
		fmt.Printf("  Issuer:    %s\n", res.Issuer)
		fmt.Printf("  Expires:   %s\n", res.ExpiresAt.Format("2006-01-02"))
		fmt.Printf("  Cert Path: %s\n", res.CertPath)
		fmt.Printf("  Key Path:  %s\n", res.KeyPath)

	case "renew":
		if len(args) < 2 {
			printError(cfg, fmt.Errorf("usage: vpsctl ssl renew <domain>"), 1)
			return
		}
		domain := args[1]

		var success bool
		err := client.Call("SSL.RenewCertificate", domain, &success)
		if err != nil {
			printError(cfg, err, 1)
			return
		}

		if cfg.JSONOutput {
			printJSON(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"domain":  domain,
					"renewed": success,
				},
			})
			return
		}

		fmt.Printf("SSL certificate renewal triggered for %s.\n", domain)

	default:
		printError(cfg, fmt.Errorf("unknown ssl subcommand: %s", sub), 1)
	}
}

