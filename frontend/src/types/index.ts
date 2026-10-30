export interface ServerOverview {
  uptime_seconds: number;
  os: {
    name: string;
    version: string;
    kernel: string;
    arch: string;
    hostname: string;
  };
  cpu: {
    model: string;
    cores: number;
    usage_percent: number;
  };
  memory: {
    total_bytes: number;
    used_bytes: number;
    free_bytes: number;
    usage_percent: number;
    swap_total_bytes: number;
    swap_used_bytes: number;
  };
  disks: Array<{
    mount: string;
    filesystem: string;
    total_bytes: number;
    used_bytes: number;
    free_bytes?: number;
    usage_percent: number;
  }>;
  network: Array<{
    interface: string;
    rx_bytes: number;
    tx_bytes: number;
  }>;
  load_avg: [number, number, number];
  timestamp: string;
}

export interface ProjectItem {
  id: string;
  name: string;
  slug: string;
  domain: string;
  project_type: 'laravel' | 'php_native' | 'nodejs' | 'nextjs' | 'react_vite' | 'python' | 'docker' | 'static';
  repository_url?: string;
  branch: string;
  root_dir: string;
  public_dir?: string;
  php_version?: string;
  upstream_port?: number;
  status: 'active' | 'building' | 'error';
  ssl_enabled: boolean;
  releases?: ReleaseItem[];
  created_at: string;
}

export interface ReleaseItem {
  id: string;
  project_id: string;
  release_tag: string;
  commit_hash?: string;
  commit_message?: string;
  branch?: string;
  status: string;
  created_at: string;
}

export interface DeploymentItem {
  id: string;
  project_id: string;
  project_slug: string;
  release_tag: string;
  trigger_type: 'manual' | 'webhook' | 'agent' | 'cli';
  triggered_by_type: string;
  status: 'queued' | 'building' | 'activating' | 'success' | 'failed';
  build_logs: string;
  duration_seconds: number;
  created_at: string;
}

export interface AIAgentItem {
  id: string;
  name: string;
  agent_type: 'hermes' | 'openclaw' | 'openhands' | 'custom';
  api_key_prefix: string;
  permissions: string[];
  allowed_project_ids?: string[];
  is_active: boolean;
  created_at: string;
  last_used_at?: string;
}

export interface ApprovalRequestItem {
  id: string;
  agent_id: string;
  agent?: AIAgentItem;
  action_type: string;
  resource_type: string;
  resource_id?: string;
  payload?: any;
  status: 'pending' | 'approved' | 'rejected';
  created_at: string;
}

export interface ServiceItem {
  name: string;
  display_name: string;
  active: boolean;
  running: boolean;
  status: string;
}

export interface DatabaseItem {
  id: string;
  name: string;
  username: string;
  engine: string;
  status: string;
}
