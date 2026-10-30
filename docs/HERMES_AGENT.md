# Hermes Agent Integration Guide

Hermes Agent connects to the **AI-Native VPS Management Panel** using the Model Context Protocol (MCP) or direct REST API / CLI integration.

---

## 1. Authentication & Credentials

1. Open the VPS Web Dashboard and navigate to **AI Agents**.
2. Click **Provision New Agent**.
3. Set the name to `Hermes Autonomous DevOps` and select type `Hermes Agent`.
4. Select the permissions required (e.g., `server.read`, `project.read`, `project.create`, `project.deploy`, `ssl.manage`).
5. Copy the generated secret key (formatted as `vps_live_<random_hex>`).

---

## 2. MCP Server Configuration (`hermes.json`)

Add the VPS Panel MCP server to your Hermes configuration file:

```json
{
  "name": "DevOps Supervisor",
  "model": "deepseek-coder",
  "mcpServers": {
    "vps_panel": {
      "transport": "sse",
      "url": "https://vps.yourdomain.com/api/v1/mcp/call",
      "headers": {
        "Authorization": "Bearer vps_live_8f0a2c3d4e5f67890abcdef123456789"
      }
    }
  }
}
```

Or for local execution via `stdio`:

```json
{
  "mcpServers": {
    "vps_panel_local": {
      "command": "/usr/local/bin/vps-mcp",
      "args": ["--mode=stdio"],
      "env": {
        "VPSD_SOCKET_PATH": "/run/vpsd/vpsd.sock"
      }
    }
  }
}
```

---

## 3. Discovered Tools

Hermes Agent automatically discovers and can execute the following tools:

| Tool Name | Description |
|---|---|
| `server_get_status` | Query host specs, OS version, architecture, and uptime |
| `server_get_resources` | Fetch real-time CPU, RAM, disk, and load metrics |
| `project_list` | List all hosted web projects and their active releases |
| `project_create` | Provision a new website, create folders, and configure Nginx vhost |
| `project_deploy` | Trigger an atomic zero-downtime release deployment |
| `project_rollback` | Roll back an application to the previous release |
| `ssl_install` | Automatically request and configure Let's Encrypt certificate |
| `service_restart` | Restart host services (`nginx`, `php8.4-fpm`, `mysql`, `redis`) |
| `firewall_get_status` | Inspect UFW firewall rules |

---

## 4. Human-in-the-Loop Safety Handling

When Hermes initiates a sensitive action (such as `project.delete` or `server.reboot`), the panel returns:

```json
{
  "status": "pending_approval",
  "approval_id": "appr_7c01a9df",
  "message": "Action 'project.delete' requires human administrator approval. Approval request submitted.",
  "expires_at": "2026-10-31T02:05:30Z"
}
```

Hermes will pause and inform the human administrator:
> *"I have prepared the deletion of project 'old-staging'. A human approval request (`appr_7c01a9df`) has been submitted to your dashboard for confirmation."*

Once the admin clicks **Approve** in the dashboard, the daemon executes the operation.

