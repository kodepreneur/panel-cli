import React, { useState, useEffect, useCallback } from 'react';
import { Sidebar, NavItem } from './components/Sidebar';
import { Header } from './components/Header';
import { OverviewView } from './components/views/OverviewView';
import { ProjectsView } from './components/views/ProjectsView';
import { DomainsView } from './components/views/DomainsView';
import { AIAgentsView } from './components/views/AIAgentsView';
import { DeploymentsView } from './components/views/DeploymentsView';
import { ServicesView } from './components/views/ServicesView';
import { TerminalView } from './components/views/TerminalView';
import { GitHubView } from './components/views/GitHubView';
import { ErrorBoundary } from './components/ErrorBoundary';
import {
  ServerOverview,
  ProjectItem,
  ServiceItem,
  AIAgentItem,
  ApprovalRequestItem,
  DeploymentItem,
  SystemUpdateStatus,
  SystemUpdateResult,
} from './types';
import { api } from './api/client';
import { X, Copy, Check, Terminal, GitPullRequest, RefreshCw, AlertCircle, CheckCircle2 } from 'lucide-react';

export const App: React.FC = () => {
  const [currentTab, setCurrentTab] = useState<NavItem>('overview');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [cliModalOpen, setCliModalOpen] = useState(false);
  const [copiedCLI, setCopiedCLI] = useState(false);

  // System Update Modal State
  const [systemUpdateModalOpen, setSystemUpdateModalOpen] = useState(false);
  const [systemUpdateStatus, setSystemUpdateStatus] = useState<SystemUpdateStatus | null>(null);
  const [isCheckingUpdate, setIsCheckingUpdate] = useState(false);
  const [isUpdatingSystem, setIsUpdatingSystem] = useState(false);
  const [systemUpdateResult, setSystemUpdateResult] = useState<SystemUpdateResult | null>(null);
  const [systemUpdateError, setSystemUpdateError] = useState<string | null>(null);

  // Live state fetched directly from control plane backend

  const [overview, setOverview] = useState<ServerOverview | undefined>();
  const [projects, setProjects] = useState<ProjectItem[]>([]);
  const [services, setServices] = useState<ServiceItem[]>([]);
  const [agents, setAgents] = useState<AIAgentItem[]>([]);
  const [pendingApprovals, setPendingApprovals] = useState<ApprovalRequestItem[]>([]);
  const [deployments, setDeployments] = useState<DeploymentItem[]>([]);

  const fetchResources = useCallback(async () => {
    try {
      const data = await api.getResources();
      setOverview(data);
    } catch (e) {
      console.error('Failed to fetch server resources:', e);
    }
  }, []);

  const fetchServicesData = useCallback(async () => {
    try {
      const data = await api.getServices();
      setServices(data);
    } catch (e) {
      console.error('Failed to fetch services:', e);
    }
  }, []);

  const fetchProjectsData = useCallback(async () => {
    try {
      const data = await api.getProjects();
      setProjects(data);

      // Extract real deployments across projects
      const allDeps: DeploymentItem[] = [];
      data.forEach((p: any) => {
        if (p.deployments && Array.isArray(p.deployments)) {
          p.deployments.forEach((d: any) => {
            allDeps.push({
              ...d,
              project_slug: p.slug,
            });
          });
        } else if (p.releases && Array.isArray(p.releases)) {
          p.releases.forEach((r: any) => {
            allDeps.push({
              id: `dep_${r.id}`,
              project_id: p.id,
              project_slug: p.slug,
              release_tag: r.release_tag,
              trigger_type: 'cli',
              triggered_by_type: 'vpsctl',
              status: r.status === 'active' ? 'success' : 'failed',
              build_logs: `[vps-deploy] Project: ${p.name} (${p.slug})\n[vps-deploy] Release: ${r.release_tag}\n[vps-deploy] Status: ${r.status}\n[vps-deploy] Path: ${r.path || p.root_dir}`,
              duration_seconds: 3,
              created_at: r.created_at,
            });
          });
        }
      });
      allDeps.sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      setDeployments(allDeps);
    } catch (e) {
      console.error('Failed to fetch projects:', e);
    }
  }, []);

  const fetchAgentsData = useCallback(async () => {
    try {
      const data = await api.getAgents();
      setAgents(data);
    } catch (e) {
      console.error('Failed to fetch agents:', e);
    }
  }, []);

  const fetchApprovalsData = useCallback(async () => {
    try {
      const data = await api.getPendingApprovals();
      setPendingApprovals(data);
    } catch (e) {
      console.error('Failed to fetch approvals:', e);
    }
  }, []);

  const refreshAll = useCallback(async () => {
    setIsRefreshing(true);
    try {
      await Promise.allSettled([
        fetchResources(),
        fetchServicesData(),
        fetchProjectsData(),
        fetchAgentsData(),
        fetchApprovalsData(),
      ]);
    } finally {
      setIsRefreshing(false);
    }
  }, [
    fetchResources,
    fetchServicesData,
    fetchProjectsData,
    fetchAgentsData,
    fetchApprovalsData,
  ]);

  useEffect(() => {
    refreshAll();
    const interval = setInterval(() => {
      fetchResources();
      fetchApprovalsData();
    }, 8000);
    return () => clearInterval(interval);
  }, [refreshAll, fetchResources, fetchApprovalsData]);

  // Action handlers connected to backend REST API
  const handleCreateProject = async (data: {
    name: string;
    domain: string;
    project_type: string;
    repository_url?: string;
    branch?: string;
    enable_ssl: boolean;
  }) => {
    await api.createProject(data);
    await fetchProjectsData();
  };

  const handleDeploy = async (id: string) => {
    const proj = projects.find((p) => p.id === id);
    if (!proj) return;
    try {
      await api.deployProject(id);
      await fetchProjectsData();
      alert(`Deployment triggered and activated for ${proj.name}!`);
    } catch (err: any) {
      alert(`Deployment failed: ${err.message}`);
    }
  };

  const handleUpdateFromRepo = async (
    id: string,
    options: { repository_url?: string; branch?: string; run_build: boolean }
  ) => {
    const res = await api.updateProjectFromRepo(id, options);
    await fetchProjectsData();
    return {
      logs: res.logs,
      commit_hash: res.release?.commit_hash,
      release_tag: res.release?.release_tag,
    };
  };

  const handleOpenSystemUpdate = async () => {
    setSystemUpdateModalOpen(true);
    setSystemUpdateError(null);
    setSystemUpdateResult(null);
    setIsCheckingUpdate(true);
    try {
      const status = await api.getSystemUpdateStatus();
      setSystemUpdateStatus(status);
    } catch (err: any) {
      setSystemUpdateError(err.message || 'Failed to check system update status');
    } finally {
      setIsCheckingUpdate(false);
    }
  };

  const handleExecuteSystemUpdate = async () => {
    setIsUpdatingSystem(true);
    setSystemUpdateError(null);
    try {
      const res = await api.updateSystem(systemUpdateStatus?.current_branch || 'main');
      setSystemUpdateResult(res);
      // Refresh status after update
      const status = await api.getSystemUpdateStatus();
      setSystemUpdateStatus(status);
      await refreshAll();
    } catch (err: any) {
      setSystemUpdateError(err.message || 'Failed to execute system update');
    } finally {
      setIsUpdatingSystem(false);
    }
  };

  const handleRollback = async (id: string) => {
    const proj = projects.find((p) => p.id === id);
    if (!proj) return;
    try {
      await api.rollbackProject(id);
      await fetchProjectsData();
      alert(`Project '${proj.name}' rolled back to previous active release.`);
    } catch (err: any) {
      alert(`Rollback failed: ${err.message}`);
    }
  };

  const handleDeleteProject = async (id: string) => {
    const proj = projects.find((p) => p.id === id);
    if (!confirm(`Are you sure you want to delete ${proj?.name || 'this project'}?`)) {
      return;
    }
    try {
      await api.deleteProject(id);
      await fetchProjectsData();
    } catch (err: any) {
      alert(`Delete failed: ${err.message}`);
    }
  };

  const handleInstallSSL = async (projectId: string) => {
    try {
      await api.installSSL(projectId);
      await fetchProjectsData();
      alert('SSL Certificate successfully issued and virtual host configured.');
    } catch (err: any) {
      alert(`SSL install failed: ${err.message}`);
    }
  };

  const handleCreateAgent = async (data: {
    name: string;
    agent_type: string;
    permissions: string[];
  }) => {
    const res = await api.createAgent(data);
    await fetchAgentsData();
    return res;
  };

  const handleRevokeAgent = async (id: string) => {
    try {
      await api.revokeAgent(id);
      await fetchAgentsData();
    } catch (err: any) {
      alert(`Revoke failed: ${err.message}`);
    }
  };

  const handleApproveAction = async (approvalId: string) => {
    try {
      await api.approveAction(approvalId);
      await fetchApprovalsData();
      await fetchProjectsData();
      alert('Action approved and executed by privileged host daemon.');
    } catch (err: any) {
      alert(`Approval execution failed: ${err.message}`);
    }
  };

  const handleRejectAction = async (approvalId: string) => {
    try {
      await api.rejectAction(approvalId);
      await fetchApprovalsData();
      alert('Action rejected.');
    } catch (err: any) {
      alert(`Rejection failed: ${err.message}`);
    }
  };

  const handleRestartService = async (name: string) => {
    try {
      await api.restartService(name);
      await fetchServicesData();
      alert(`Service '${name}' restarted successfully.`);
    } catch (err: any) {
      alert(`Service restart failed: ${err.message}`);
    }
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
          onRefresh={refreshAll}
          isRefreshing={isRefreshing}
          onOpenCLIHint={() => setCliModalOpen(true)}
          onOpenSystemUpdate={handleOpenSystemUpdate}
        />

        <main className="flex-1 overflow-y-auto p-8">
          <ErrorBoundary fallbackTitle="View Rendering Error">
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
                onUpdateFromRepo={handleUpdateFromRepo}
                onRollback={handleRollback}
                onDelete={handleDeleteProject}
                onViewLogs={() => setCurrentTab('deployments')}
              />
            )}

            {currentTab === 'domains' && (
              <DomainsView projects={projects} onInstallSSL={handleInstallSSL} />
            )}

            {currentTab === 'github' && (
              <GitHubView
                projects={projects}
                onUpdateFromRepo={handleUpdateFromRepo}
                onViewLogs={() => setCurrentTab('deployments')}
              />
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

            {(currentTab === 'databases' ||
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
          </ErrorBoundary>
        </main>
      </div>

      {/* System Update Modal */}
      {systemUpdateModalOpen && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-lg border border-slate-200 shadow-xl max-w-lg w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <GitPullRequest className="w-4 h-4 text-emerald-600" />
                <div>
                  <h3 className="text-sm font-bold text-slate-900">VPS Panel System Update</h3>
                  <p className="text-[11px] text-slate-500">Update control plane and binaries from git repository</p>
                </div>
              </div>
              <button
                onClick={() => setSystemUpdateModalOpen(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="p-6 space-y-4 text-xs">
              {systemUpdateError && (
                <div className="bg-rose-50 border border-rose-200 text-rose-700 px-3.5 py-2.5 rounded text-xs flex items-start gap-2">
                  <AlertCircle className="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
                  <div>
                    <div className="font-semibold">Update check failed</div>
                    <div className="text-[11px] mt-0.5">{systemUpdateError}</div>
                  </div>
                </div>
              )}

              {/* Status details card */}
              <div className="bg-slate-50 border border-slate-200 rounded-lg p-4 space-y-2.5">
                <div className="flex items-center justify-between pb-2 border-b border-slate-200">
                  <span className="text-slate-500 font-medium">Panel Version</span>
                  <span className="font-mono font-semibold text-slate-800">
                    v{systemUpdateStatus?.current_version || '1.0.0'}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-slate-500 font-medium">Current Commit</span>
                  <span className="font-mono bg-white px-2 py-0.5 rounded border border-slate-200 text-slate-700">
                    {systemUpdateStatus?.current_commit || 'HEAD'}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-slate-500 font-medium">Active Branch</span>
                  <span className="font-mono text-blue-600 font-medium">
                    {systemUpdateStatus?.current_branch || 'main'}
                  </span>
                </div>

                <div className="flex items-center justify-between pt-1">
                  <span className="text-slate-500 font-medium">Remote Status</span>
                  {isCheckingUpdate ? (
                    <span className="flex items-center gap-1.5 text-slate-500">
                      <RefreshCw className="w-3 h-3 animate-spin" /> Checking remote...
                    </span>
                  ) : systemUpdateStatus?.update_available ? (
                    <span className="inline-flex items-center gap-1 text-amber-700 bg-amber-50 px-2 py-0.5 rounded text-[11px] border border-amber-200 font-semibold">
                      Update Available ({systemUpdateStatus.behind_count} commits)
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded text-[11px] border border-emerald-200 font-medium">
                      <CheckCircle2 className="w-3 h-3" /> Up to Date
                    </span>
                  )}
                </div>

                {systemUpdateStatus?.latest_message && (
                  <div className="pt-2 border-t border-slate-200 text-[11px] text-slate-600">
                    <span className="text-slate-400 font-medium">Latest Remote Commit:</span>{' '}
                    <span className="italic">{systemUpdateStatus.latest_message}</span>
                  </div>
                )}
              </div>

              {/* Execution result banner */}
              {systemUpdateResult && (
                <div className="bg-emerald-50 border border-emerald-200 text-emerald-800 px-3.5 py-2.5 rounded text-xs flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                    <div>
                      <div className="font-semibold">Panel Updated Successfully!</div>
                      <div className="text-[11px] text-emerald-700 font-mono mt-0.5">
                        Commit: {systemUpdateResult.new_commit}
                      </div>
                    </div>
                  </div>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 font-bold border border-emerald-300">
                    Updated
                  </span>
                </div>
              )}

              {/* Execution logs output */}
              {systemUpdateResult?.logs && (
                <div className="mt-2">
                  <div className="text-[11px] font-semibold text-slate-600 mb-1 flex items-center gap-1.5">
                    <Terminal className="w-3.5 h-3.5 text-slate-500" />
                    <span>Update Execution Logs</span>
                  </div>
                  <pre className="bg-slate-950 text-slate-200 font-mono text-[11px] p-3 rounded border border-slate-800 max-h-48 overflow-y-auto leading-relaxed whitespace-pre-wrap select-text">
                    {systemUpdateResult.logs}
                  </pre>
                </div>
              )}

              {/* Modal footer */}
              <div className="pt-4 border-t border-slate-100 flex items-center justify-between">
                <button
                  type="button"
                  onClick={handleOpenSystemUpdate}
                  disabled={isCheckingUpdate || isUpdatingSystem}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded border border-slate-200 text-slate-600 hover:bg-slate-50 font-medium text-xs disabled:opacity-50"
                >
                  <RefreshCw className={`w-3 h-3 ${isCheckingUpdate ? 'animate-spin' : ''}`} />
                  <span>Check for Updates</span>
                </button>

                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => setSystemUpdateModalOpen(false)}
                    className="px-3.5 py-1.5 rounded border border-slate-300 text-slate-700 hover:bg-slate-50 font-medium text-xs"
                  >
                    Close
                  </button>
                  <button
                    type="button"
                    onClick={handleExecuteSystemUpdate}
                    disabled={isUpdatingSystem}
                    className="flex items-center gap-1.5 px-3.5 py-1.5 rounded bg-emerald-600 hover:bg-emerald-700 text-white font-medium text-xs shadow-sm disabled:opacity-50 transition-colors"
                  >
                    {isUpdatingSystem ? (
                      <>
                        <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                        <span>Updating System...</span>
                      </>
                    ) : (
                      <>
                        <GitPullRequest className="w-3.5 h-3.5" />
                        <span>Update Panel Now</span>
                      </>
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

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
