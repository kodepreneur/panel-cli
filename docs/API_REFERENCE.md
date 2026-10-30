# REST API Reference

Base URL: `http://<your-vps-ip>/api/v1`

All authenticated endpoints require either:
- Personal Access Token: `Authorization: Bearer <sanctum_token>`
- Agent API Key: `Authorization: Bearer vps_live_<agent_key>`

---

## Endpoints Summary

### Server Management
- `GET /api/v1/server/status`: Get host OS details, architecture, hostname, and uptime.
- `GET /api/v1/server/resources`: Get real-time CPU, RAM, disk mounts, network, and load average.
- `GET /api/v1/server/services`: List tracked systemd services and statuses.
- `POST /api/v1/server/services/{name}/restart`: Restart a service (`nginx`, `mysql`, `redis`, etc.).
- `POST /api/v1/server/reboot`: Initiate server reboot (triggers human approval if called by an agent).

### Project Management
- `GET /api/v1/projects`: List hosted projects.
- `POST /api/v1/projects`: Provision a new website/project.
  - Body:
    ```json
    {
      "name": "My App",
      "domain": "app.example.com",
      "project_type": "laravel",
      "enable_ssl": true
    }
    ```
- `GET /api/v1/projects/{id}`: Get project details and releases.
- `DELETE /api/v1/projects/{id}`: Delete a project and remove its Nginx configuration.
- `POST /api/v1/projects/{id}/deploy`: Trigger an atomic release deployment.
- `POST /api/v1/projects/{id}/update-from-repo`: Pull latest commits from Git repository, run build scripts, and atomically activate a new release.
- `POST /api/v1/projects/{id}/rollback`: Roll back to the previous release.
- `GET /api/v1/projects/{id}/logs`: Retrieve recent build output logs.

### Server & System Updates
- `GET /api/v1/server/update/status`: Check current VPS Panel version, git branch/commit, and remote updates.
- `POST /api/v1/server/update`: Update VPS Panel control plane and recompile binaries from git repository.


### SSL Management
- `POST /api/v1/projects/{id}/ssl/install`: Issue Let's Encrypt certificate via Certbot.
- `GET /api/v1/projects/{id}/ssl/status`: Check certificate validity and expiration date.

### AI Agents & Approvals
- `GET /api/v1/agents`: List configured AI agent credentials.
- `POST /api/v1/agents`: Provision a new AI agent key with scoped permissions.
- `DELETE /api/v1/agents/{id}`: Revoke agent credentials immediately.
- `GET /api/v1/agents/approvals/pending`: List actions awaiting human administrator approval.
- `POST /api/v1/agents/approvals/{id}/approve`: Approve and execute an agent action.
- `POST /api/v1/agents/approvals/{id}/reject`: Reject an agent action with reason.
- `GET /api/v1/agents/audit-logs`: Query the immutable audit log ledger.

