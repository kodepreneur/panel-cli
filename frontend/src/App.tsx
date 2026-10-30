import React, { useState, useEffect } from 'react';
import { Sidebar, NavItem } from './components/Sidebar';
import { Header } from './components/Header';
import { OverviewView } from './components/views/OverviewView';
import { ProjectsView } from './components/views/ProjectsView';
import { DomainsView } from './components/views/DomainsView';
import { AIAgentsView } from './components/views/AIAgentsView';
import { DeploymentsView } from './components/views/DeploymentsView';
import { ServicesView } from './components/views/ServicesView';
import { TerminalView } from './components/views/TerminalView';
import {
  ServerOverview,
  ProjectItem,
  ServiceItem,
  AIAgentItem,
  ApprovalRequestItem,
  DeploymentItem,
} from './types';
import { X, Copy, Check, Terminal } from 'lucide-react';

export const App: React.FC = () => {
  const [currentTab, setCurrentTab] = useState<NavItem>('overview');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [cliModalOpen, setCliModalOpen] = useState(false);
  const [copiedCLI, setCopiedCLI] = useState(false);

  // Core state
  const [overview, setOverview] = useState<ServerOverview | undefined>();
  const [projects, setProjects] = useState<ProjectItem[]>([
    {
      id: 'proj_01',
      name: 'Landing Page',
      slug: 'landing-page',
      domain: 'kodepreneur.com',
      project_type: 'laravel',
      branch: 'main',
      root_dir: '/var/www/vps-panel/projects/landing-page/current',
      status: 'active',
      ssl_enabled: true,
      created_at: new Date().toISOString(),
      releases: [
        {
          id: 'rel_01',
          project_id: 'proj_01',
          release_tag: '20261031_010000',
          commit_hash: 'a7f9c2d',
          status: 'active',
          created_at: new Date().toISOString(),
        },
      ],
    },
    {
      id: 'proj_02',
      name: 'API Service',
      slug: 'api-service',
      domain: 'api.kodepreneur.com',
      project_type: 'nodejs',
      upstream_port: 3000,
      branch: 'main',
      root_dir: '/var/www/vps-panel/projects/api-service/current',
      status: 'active',
      ssl_enabled: true,
      created_at: new Date().toISOString(),
      releases: [
        {
          id: 'rel_02',
          project_id: 'proj_02',
          release_tag: '20261031_011500',
          commit_hash: '3e4b10a',
          status: 'active',
          created_at: new Date().toISOString(),
        },
      ],
    },
  ]);

  const [services] = useState<ServiceItem[]>([
    { name: 'nginx', display_name: 'Nginx Web Server', active: true, running: true, status: 'active' },
    { name: 'php8.4-fpm', display_name: 'PHP 8.4 FPM', active: true, running: true, status: 'active' },
    { name: 'mysql', display_name: 'MySQL Database', active: true, running: true, status: 'active' },
    { name: 'redis', display_name: 'Redis Cache & Queue', active: true, running: true, status: 'active' },
    { name: 'ufw', display_name: 'UFW Firewall', active: true, running: true, status: 'active' },
  ]);

  const [agents, setAgents] = useState<AIAgentItem[]>([
    {
      id: 'ag_hermes_01',
      name: 'Hermes Autonomous DevOps',
      agent_type: 'hermes',
      api_key_prefix: 'vps_live_8f0a...',
      permissions: ['server.read', 'project.read', 'project.create', 'project.deploy', 'ssl.manage'],
      is_active: true,
      created_at: new Date().toISOString(),
    },
    {
      id: 'ag_openclaw_01',
      name: 'OpenClaw Infra Agent',
      agent_type: 'openclaw',
      api_key_prefix: 'vps_live_4b2c...',
      permissions: ['server.read', 'project.read', 'service.manage'],
      is_active: true,
      created_at: new Date().toISOString(),
    },
  ]);

  const [pendingApprovals, setPendingApprovals] = useState<ApprovalRequestItem[]>([
    {
      id: 'appr_01',
      agent_id: 'ag_hermes_01',
      agent: {
        id: 'ag_hermes_01',
        name: 'Hermes Autonomous DevOps',
        agent_type: 'hermes',
        api_key_prefix: 'vps_live_8f0a...',
        permissions: [],
        is_active: true,
        created_at: new Date().toISOString(),
      },
      action_type: 'project.delete',
      resource_type: 'project',
      resource_id: 'old-staging-app',
      status: 'pending',
      created_at: new Date().toISOString(),
    },
  ]);

  const [deployments, setDeployments] = useState<DeploymentItem[]>([
    {
      id: 'dep_01',
      project_id: 'proj_01',
      project_slug: 'landing-page',
      release_tag: '20261031_010000',
      trigger_type: 'webhook',
      triggered_by_type: 'github',
      status: 'success',
      build_logs: `[vps-deploy] git pull origin main\n[vps-deploy] composer install --no-dev --optimize-autoloader\n[vps-deploy] php artisan optimize\n[vps-deploy] Symlink updated: current -> releases/20261031_010000\n[vps-deploy] Deployment completed in 3.4s.`,
      duration_seconds: 3,
      created_at: new Date().toISOString(),
    },
    {
      id: 'dep_02',
      project_id: 'proj_02',
      project_slug: 'api-service',
      release_tag: '20261031_011500',
      trigger_type: 'agent',
      triggered_by_type: 'hermes',
      status: 'success',
      build_logs: `[vps-deploy] Triggered by Hermes AI Agent.\n[vps-deploy] npm install --production\n[vps-deploy] systemctl reload vps-api-service\n[vps-deploy] Release activated cleanly.`,
      duration_seconds: 2,
      created_at: new Date().toISOString(),
    },
  ]);

  const fetchMetrics = async () => {
    setIsRefreshing(true);
    try {
      const res = await fetch('/api/v1/server/resources');
      if (res.ok) {
        const json = await res.json();
        setOverview(json.data);
      }
    } catch (e) {
      // Mock metrics in dev mode
      setOverview({
        uptime_seconds: 72400,
        os: {
          name: 'Ubuntu',
          version: 'Ubuntu 24.04 LTS',
          kernel: '6.8.0-generic',
          arch: 'x86_64',
          hostname: 'vps-server.local',
        },
        cpu: {
          model: 'AMD EPYC 7763 64-Core Processor',
          cores: 4,
          usage_percent: 21.4,
        },
        memory: {
          total_bytes: 8589934592,
          used_bytes: 3435973836,
          free_bytes: 5153960756,
          usage_percent: 40.0,
          swap_total_bytes: 2147483648,
          swap_used_bytes: 0,
        },
        disks: [
          {
            mount: '/',
            filesystem: 'ext4',
            total_bytes: 107374182400,
            used_bytes: 23088947200,
            free_bytes: 84285235200,
            usage_percent: 21.5,
          },
        ],
        network: [
          {
            interface: 'eth0',
            rx_bytes: 104857600,
            tx_bytes: 52428800,
          },
        ],
        load_avg: [0.42, 0.55, 0.61],
        timestamp: new Date().toISOString(),
      });
    } finally {
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 10000);
    return () => clearInterval(interval);
  }, []);

  // Action handlers
  const handleCreateProject = async (data: {
    name: string;
    domain: string;
    project_type: string;
    repository_url?: string;
    branch?: string;
    enable_ssl: boolean;
  }) => {
    const slug = data.name.toLowerCase().replace(/[^a-z0-9]/g, '-');
    const newProj: ProjectItem = {
      id: `proj_${Date.now()}`,
      name: data.name,
      slug,
      domain: data.domain,
      project_type: data.project_type as any,
      branch: data.branch || 'main',
      root_dir: `/var/www/vps-panel/projects/${slug}/current`,
      status: 'active',
      ssl_enabled: data.enable_ssl,
      created_at: new Date().toISOString(),
      releases: [
        {
          id: `rel_${Date.now()}`,
          project_id: `proj_${Date.now()}`,
          release_tag: new Date().toISOString().replace(/[-:T.]/g, '').slice(0, 15),
          status: 'active',
          created_at: new Date().toISOString(),
        },
      ],
    };
    setProjects([newProj, ...projects]);
  };

  const handleDeploy = async (id: string) => {
    const proj = projects.find((p) => p.id === id);
    if (!proj) return;
    const releaseTag = new Date().toISOString().replace(/[-:T.]/g, '').slice(0, 15);
    const newDep: DeploymentItem = {
      id: `dep_${Date.now()}`,
      project_id: proj.id,
      project_slug: proj.slug,
      release_tag: releaseTag,
      trigger_type: 'manual',
      triggered_by_type: 'dashboard',
      status: 'success',
      build_logs: `[vps-deploy] Manual trigger for ${proj.slug}\n[vps-deploy] Activated release ${releaseTag} cleanly.`,
      duration_seconds: 3,
      created_at: new Date().toISOString(),
    };
    setDeployments([newDep, ...deployments]);
    alert(`Deployment triggered and activated for ${proj.name}!`);
  };

  const handleRollback = async (id: string) => {
    const proj = projects.find((p) => p.id === id);
    if (!proj) return;
    alert(`Project '${proj.name}' rolled back to previous active release.`);
  };

  const handleDeleteProject = async (id: string) => {
    if (!confirm('Are you sure you want to delete this project and remove its Nginx configuration?')) {
      return;
    }
    setProjects(projects.filter((p) => p.id !== id));
  };

  const handleInstallSSL = async (projectId: string) => {
    setProjects(
      projects.map((p) => (p.id === projectId ? { ...p, ssl_enabled: true } : p))
    );
  };

  const handleCreateAgent = async (data: {
    name: string;
    agent_type: string;
    permissions: string[];
  }) => {
    const rawKey = `vps_live_${Math.random().toString(36).substring(2)}${Math.random().toString(36).substring(2)}`;
    const newAgent: AIAgentItem = {
      id: `ag_${Date.now()}`,
      name: data.name,
      agent_type: data.agent_type as any,
      api_key_prefix: rawKey.substring(0, 15) + '...',
      permissions: data.permissions,
      is_active: true,
      created_at: new Date().toISOString(),
    };
    setAgents([newAgent, ...agents]);
    return { agent: newAgent, api_key: rawKey };
  };

  const handleRevokeAgent = async (id: string) => {
    setAgents(agents.filter((a) => a.id !== id));
  };

  const handleApproveAction = async (approvalId: string) => {
    setPendingApprovals(pendingApprovals.filter((a) => a.id !== approvalId));
    alert('Action approved and executed by privileged host daemon.');
  };

  const handleRejectAction = async (approvalId: string) => {
    setPendingApprovals(pendingApprovals.filter((a) => a.id !== approvalId));
    alert('Action rejected.');
  };

  const handleRestartService = async (name: string) => {
    alert(`Service '${name}' restarted successfully.`);
  };

  const copyCLIText = () => {
    navigator.clipboard.writeText('vpsctl server status --json');
    setCopiedCLI(true);
    setTimeout(() => setCopiedCLI(false), 2000);
  };

  return (
    <div className="flex h-screen bg-slate-50 text-slate-900 overflow-hidden">
      {/* Dark Sidebar */}
      <Sidebar
        currentTab={currentTab}
        onSelectTab={setCurrentTab}
        pendingApprovalsCount={pendingApprovals.length}
      />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Header
          overview={overview}
          onRefresh={fetchMetrics}
          isRefreshing={isRefreshing}
          onOpenCLIHint={() => setCliModalOpen(true)}
        />

        <main className="flex-1 overflow-y-auto p-8">
          {currentTab === 'overview' && (
            <OverviewView
              overview={overview}
              projects={projects}
              services={services}
              onAddProject={() => setCurrentTab('projects')}
              onDeployProject={handleDeploy}
              onSelectTab={setCurrentTab}
            />
          )}

          {currentTab === 'projects' && (
            <ProjectsView
              projects={projects}
              onCreateProject={handleCreateProject}
              onDeploy={handleDeploy}
              onRollback={handleRollback}
              onDelete={handleDeleteProject}
              onViewLogs={() => setCurrentTab('deployments')}
            />
          )}

          {currentTab === 'domains' && (
            <DomainsView projects={projects} onInstallSSL={handleInstallSSL} />
          )}

          {currentTab === 'agents' && (
            <AIAgentsView
              agents={agents}
              pendingApprovals={pendingApprovals}
              onCreateAgent={handleCreateAgent}
              onRevokeAgent={handleRevokeAgent}
              onApproveAction={handleApproveAction}
              onRejectAction={handleRejectAction}
            />
          )}

          {currentTab === 'deployments' && (
            <DeploymentsView
              deployments={deployments}
              selectedLogs={null}
            />
          )}

          {currentTab === 'services' && (
            <ServicesView
              services={services}
              onRestartService={handleRestartService}
            />
          )}

          {currentTab === 'terminal' && <TerminalView />}

          {(currentTab === 'github' ||
            currentTab === 'databases' ||
            currentTab === 'backups' ||
            currentTab === 'activity' ||
            currentTab === 'settings') && (
            <div className="bg-white border border-slate-200 rounded-lg p-8 text-center text-slate-500">
              <h2 className="text-base font-semibold text-slate-800 capitalize mb-2">
                {currentTab} Management
              </h2>
              <p className="text-xs max-w-md mx-auto text-slate-500">
                Full programmatic control available via REST API and the <code>vpsctl {currentTab}</code> CLI command.
              </p>
            </div>
          )}
        </main>
      </div>

      {/* CLI Hint Modal */}
      {cliModalOpen && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-lg border border-slate-200 shadow-xl max-w-lg w-full p-6">
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <div className="flex items-center gap-2">
                <Terminal className="w-4 h-4 text-blue-600" />
                <h3 className="text-sm font-bold text-slate-900">vpsctl CLI Quick Access</h3>
              </div>
              <button
                onClick={() => setCliModalOpen(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="mt-4 space-y-3 text-xs">
              <p className="text-slate-600">
                You can manage the server directly from any terminal or SSH session using the compiled <code>vpsctl</code> binary:
              </p>
              <div className="bg-slate-950 text-slate-200 p-3 rounded font-mono text-[11px] flex items-center justify-between">
                <span>vpsctl server status --json</span>
                <button
                  onClick={copyCLIText}
                  className="p-1 hover:text-white rounded"
                >
                  {copiedCLI ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                </button>
              </div>
              <div className="space-y-1 text-slate-600">
                <div><code>vpsctl project list</code> — List all hosted projects</div>
                <div><code>vpsctl project deploy &lt;slug&gt;</code> — Trigger zero-downtime release</div>
                <div><code>vpsctl ssl install &lt;domain&gt;</code> — Issue Let's Encrypt certificate</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
