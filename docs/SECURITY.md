# Security Architecture & Threat Model

Security is an architectural invariant of the **AI-Native VPS Management Panel**. This document outlines the defense-in-depth safeguards, privilege boundaries, and cryptographic standards implemented across the stack.

---

## 1. Privilege Separation (Dual-Plane Architecture)

```text
[ Web Interface / Laravel Control Plane (www-data:www-data) ]
                         │
                         │ Strictly typed JSON-RPC 2.0
                         │ Local Unix Domain Socket (/run/vpsd/vpsd.sock)
                         ▼
[ Privileged Execution Daemon: vpsd (root:panel-rpc) ]
                         │
                         ▼
[ Host Services: Nginx, Certbot, UFW, Systemd ]
```

- **Unprivileged Web Layer**: The web application and PHP-FPM process run as unprivileged user `www-data` with supplementary group `panel-rpc`. The web server has **no sudoers privileges** and cannot execute arbitrary shell commands as root.
- **Strictly Scoped Daemon RPC**: All host mutations (Nginx configuration generation, cert issuance, systemd restarts) are handled by `vpsd`.
- **Zero Raw Shell Concatenation**: Commands inside `vpsd` are executed using parameterized process execution (`exec.Command(binary, args...)`) with strict regex validation (`^[a-zA-Z0-9_-]+$`).

---

## 2. AI Agent Sandbox & Access Control

1. **API Key Security**:
   - Agent API keys use the `vps_live_` prefix followed by 24 cryptographically random bytes (48 hex characters).
   - Only the SHA-256 hash of the API key is stored in the database (`api_key_hash`). Plaintext keys are never stored or displayed after creation.

2. **Fine-Grained RBAC**:
   - Agents are assigned only the minimal permissions necessary (`server.read`, `project.deploy`, `ssl.manage`).
   - Project scope isolation: Agents can be restricted to specific project IDs, preventing cross-tenant interference.

3. **Human-in-the-Loop Approval Workflow**:
   - Destructive actions (`project.delete`, `server.reboot`, `firewall.modify`) are intercepted by middleware.
   - The action is blocked, an immutable `ApprovalRequest` record is generated, and a `202 Accepted (pending_approval)` response is returned.
   - Execution is held until an administrator confirms the action via the Web Dashboard or `vpsctl agent approvals`.

4. **Audit Logging**:
   - Every agent API request and MCP tool call is recorded in `agent_audit_logs` with timestamp, client IP, requested endpoint, sanitized payload parameters, HTTP response status, and duration in milliseconds.

---

## 3. Atomic Symlink Releases & Path Traversal Prevention

- Hosted projects are organized using the atomic symlink release pattern:
  `/var/www/vps-panel/projects/<slug>/releases/<timestamp>`
- Deployments flip the `current` symlink atomically via POSIX rename (`os.Rename`), guaranteeing zero downtime and immediate one-step rollback.
- Slugs are validated against strict alphanumeric rules to completely eliminate directory traversal (`../`) attacks.
- Project root directories are locked to `/var/www/vps-panel/projects/`.

---

## 4. Webhook Integrity

- Inbound GitHub webhooks are verified using HMAC-SHA256 signatures (`X-Hub-Signature-256`) against the configured secret key using constant-time hash comparison (`hash_equals`).

