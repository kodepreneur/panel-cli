# OpenClaw Agent Integration Guide

OpenClaw connects to the VPS Panel through either Model Context Protocol (MCP) tool bindings or scripted CLI automation via `vpsctl --json`.

---

## 1. OpenClaw YAML Configuration (`openclaw.yaml`)

Configure your OpenClaw agent manifest with the remote VPS MCP endpoint:

```yaml
version: "1.0"
agent:
  name: "OpenClaw-Infra"
  description: "Autonomous cloud server administrator"
  capabilities:
    - autonomous_execution
    - tool_use

tools:
  - name: "vps_control"
    type: "mcp"
    endpoint: "https://vps.yourdomain.com/api/v1/mcp/call"
    auth:
      type: "bearer"
      token: "${VPS_AGENT_KEY}" # vps_live_...

policies:
  require_confirmation_for:
    - "project_delete"
    - "firewall_modify"
    - "server_reboot"
```

---

## 2. CLI Mode Integration (`vpsctl`)

If OpenClaw has direct shell or container access to the server, it can invoke `vpsctl` directly with the `--json` flag:

```bash
# Query resources
vpsctl server resources --json

# Deploy a project
vpsctl project deploy my-api --json

# Roll back
vpsctl project rollback my-api --json
```

All commands return structured JSON adhering to the standard envelope:

```json
{
  "status": "success",
  "data": { ... }
}
```

Exit codes:
- `0`: Execution succeeded.
- `1`: Validation or parameter error.
- `2`: Pending human approval required.
- `3`: Fatal daemon communication failure.

