import {
  ServerOverview,
  ProjectItem,
  ServiceItem,
  AIAgentItem,
  ApprovalRequestItem,
  DeploymentItem,
} from '../types';

const DEFAULT_TOKEN = 'vps_live_test_agent_secret_key_8f0a2c3d';

export function getApiToken(): string {
  return localStorage.getItem('vps_panel_token') || DEFAULT_TOKEN;
}

export function setApiToken(token: string): void {
  localStorage.setItem('vps_panel_token', token);
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = getApiToken();
  const headers = new Headers(options.headers || {});

  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  if (!headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  headers.set('X-Panel-Admin', 'true');

  const response = await fetch(endpoint, {
    ...options,
    headers,
  });

  const json = await response.json().catch(() => null);

  if (!response.ok) {
    const errorMsg = json?.message || `HTTP ${response.status}: ${response.statusText}`;
    throw new Error(errorMsg);
  }

  return json;
}

export const api = {
  // Server resources & status
  async getResources(): Promise<ServerOverview> {
    const res = await request<{ status: string; data: ServerOverview }>('/api/v1/server/resources');
    return res.data;
  },

  async getServices(): Promise<ServiceItem[]> {
    const res = await request<{ status: string; data: ServiceItem[] }>('/api/v1/server/services');
    return res.data;
  },

  async restartService(name: string): Promise<boolean> {
    const res = await request<{ status: string; data: { restarted: boolean } }>(
      `/api/v1/server/services/${encodeURIComponent(name)}/restart`,
      { method: 'POST' }
    );
    return res.data?.restarted ?? true;
  },

  // Projects
  async getProjects(): Promise<ProjectItem[]> {
    const res = await request<{ status: string; data: ProjectItem[] }>('/api/v1/projects');
    return res.data;
  },

  async createProject(data: {
    name: string;
    domain: string;
    project_type: string;
    repository_url?: string;
    branch?: string;
    enable_ssl: boolean;
  }): Promise<ProjectItem> {
    const res = await request<{ status: string; data: ProjectItem }>('/api/v1/projects', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return res.data;
  },

  async deleteProject(id: string): Promise<void> {
    await request(`/api/v1/projects/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  },

  async deployProject(id: string): Promise<{ deployment: DeploymentItem }> {
    const res = await request<{ status: string; data: { deployment: DeploymentItem } }>(
      `/api/v1/projects/${encodeURIComponent(id)}/deploy`,
      { method: 'POST' }
    );
    return res.data;
  },

  async rollbackProject(id: string, targetTag?: string): Promise<void> {
    await request(`/api/v1/projects/${encodeURIComponent(id)}/rollback`, {
      method: 'POST',
      body: JSON.stringify({ target_release_tag: targetTag }),
    });
  },

  async getProjectLogs(id: string): Promise<string> {
    const res = await request<{ status: string; data: { logs: string } }>(
      `/api/v1/projects/${encodeURIComponent(id)}/logs`
    );
    return res.data?.logs || '';
  },

  async installSSL(id: string, email?: string): Promise<void> {
    await request(`/api/v1/projects/${encodeURIComponent(id)}/ssl/install`, {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  },

  // AI Agents
  async getAgents(): Promise<AIAgentItem[]> {
    const res = await request<{ status: string; data: AIAgentItem[] }>('/api/v1/agents');
    return res.data;
  },

  async createAgent(data: {
    name: string;
    agent_type: string;
    permissions: string[];
  }): Promise<{ agent: AIAgentItem; api_key: string }> {
    const res = await request<{ status: string; data: { agent: AIAgentItem; api_key: string } }>(
      '/api/v1/agents',
      {
        method: 'POST',
        body: JSON.stringify(data),
      }
    );
    return res.data;
  },

  async revokeAgent(id: string): Promise<void> {
    await request(`/api/v1/agents/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  },

  // Approvals
  async getPendingApprovals(): Promise<ApprovalRequestItem[]> {
    const res = await request<{ status: string; data: ApprovalRequestItem[] }>(
      '/api/v1/agents/approvals/pending'
    );
    return res.data;
  },

  async approveAction(id: string): Promise<void> {
    await request(`/api/v1/agents/approvals/${encodeURIComponent(id)}/approve`, {
      method: 'POST',
    });
  },

  async rejectAction(id: string, reason?: string): Promise<void> {
    await request(`/api/v1/agents/approvals/${encodeURIComponent(id)}/reject`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    });
  },

  // Terminal Execution
  async executeTerminal(command: string): Promise<{ output: string; duration_ms: number; timestamp: string }> {
    const res = await request<{
      status: string;
      output: string;
      duration_ms: number;
      timestamp: string;
    }>('/api/v1/server/terminal', {
      method: 'POST',
      body: JSON.stringify({ command }),
    });
    return res;
  },
};

