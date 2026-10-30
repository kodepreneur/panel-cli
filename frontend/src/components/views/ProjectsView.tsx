import React, { useState } from 'react';
import {
  FolderGit2,
  Plus,
  Rocket,
  RotateCcw,
  ScrollText,
  Trash2,
  Globe,
  CheckCircle2,
  X,
  ExternalLink,
  GitPullRequest,
  GitBranch,
  RefreshCw,
  Terminal,
  AlertCircle,
} from 'lucide-react';
import { ProjectItem } from '../../types';

interface ProjectsViewProps {
  projects: ProjectItem[];
  onCreateProject: (data: {
    name: string;
    domain: string;
    project_type: string;
    repository_url?: string;
    branch?: string;
    enable_ssl: boolean;
  }) => Promise<void>;
  onDeploy: (id: string) => Promise<void>;
  onUpdateFromRepo: (
    id: string,
    options: { repository_url?: string; branch?: string; run_build: boolean }
  ) => Promise<{ logs: string; commit_hash?: string; release_tag?: string }>;
  onRollback: (id: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onViewLogs: (slug: string) => void;
}

export const ProjectsView: React.FC<ProjectsViewProps> = ({
  projects,
  onCreateProject,
  onDeploy,
  onUpdateFromRepo,
  onRollback,
  onDelete,
  onViewLogs,
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [filterType, setFilterType] = useState<string>('all');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Update from Repo Modal State
  const [isUpdateModalOpen, setIsUpdateModalOpen] = useState(false);
  const [selectedProject, setSelectedProject] = useState<ProjectItem | null>(null);
  const [updateRepoUrl, setUpdateRepoUrl] = useState('');
  const [updateBranch, setUpdateBranch] = useState('main');
  const [updateRunBuild, setUpdateRunBuild] = useState(true);
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateLogs, setUpdateLogs] = useState<string | null>(null);
  const [updateError, setUpdateError] = useState<string | null>(null);
  const [updateCommitInfo, setUpdateCommitInfo] = useState<{ hash?: string; tag?: string } | null>(null);

  // Form states
  const [name, setName] = useState('');
  const [domain, setDomain] = useState('');
  const [projectType, setProjectType] = useState('laravel');
  const [repoUrl, setRepoUrl] = useState('');
  const [branch] = useState('main');
  const [enableSSL, setEnableSSL] = useState(true);

  const openUpdateModal = (project: ProjectItem) => {
    setSelectedProject(project);
    setUpdateRepoUrl(project.repository_url || '');
    setUpdateBranch(project.branch || 'main');
    setUpdateRunBuild(true);
    setUpdateLogs(null);
    setUpdateError(null);
    setUpdateCommitInfo(null);
    setIsUpdateModalOpen(true);
  };

  const handleUpdateFromRepo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject) return;
    if (!updateRepoUrl && !selectedProject.repository_url) {
      setUpdateError('Please provide a Git repository URL.');
      return;
    }

    setIsUpdating(true);
    setUpdateError(null);
    setUpdateLogs(null);

    try {
      const res = await onUpdateFromRepo(selectedProject.id, {
        repository_url: updateRepoUrl || undefined,
        branch: updateBranch || 'main',
        run_build: updateRunBuild,
      });

      setUpdateLogs(res.logs);
      setUpdateCommitInfo({
        hash: res.commit_hash,
        tag: res.release_tag,
      });
    } catch (err: any) {
      setUpdateError(err.message || 'Failed to update project from repository.');
    } finally {
      setIsUpdating(false);
    }
  };


  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !domain) return;
    setIsSubmitting(true);
    try {
      await onCreateProject({
        name,
        domain,
        project_type: projectType,
        repository_url: repoUrl || undefined,
        branch,
        enable_ssl: enableSSL,
      });
      setIsModalOpen(false);
      setName('');
      setDomain('');
      setRepoUrl('');
    } finally {
      setIsSubmitting(false);
    }
  };

  const filtered = filterType === 'all'
    ? projects
    : projects.filter((p) => p.project_type === filterType);

  return (
    <div className="space-y-6">
      {/* View Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">Websites & Application Projects</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Manage multi-tenant applications, virtual hosts, zero-downtime deployments, and SSL certificates.
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded bg-blue-600 hover:bg-blue-700 text-white shadow-sm transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>Add Project</span>
        </button>
      </div>

      {/* Filter Chips */}
      <div className="flex items-center gap-2 border-b border-slate-200 pb-3 text-xs overflow-x-auto">
        <span className="text-slate-400 font-medium mr-1">Runtimes:</span>
        {['all', 'laravel', 'nodejs', 'nextjs', 'react_vite', 'python', 'docker', 'static'].map((type) => (
          <button
            key={type}
            onClick={() => setFilterType(type)}
            className={`px-2.5 py-1 rounded-md font-mono transition-colors ${
              filterType === type
                ? 'bg-slate-900 text-white font-medium'
                : 'bg-white border border-slate-200 text-slate-600 hover:bg-slate-50'
            }`}
          >
            {type}
          </button>
        ))}
      </div>

      {/* Projects Table */}
      <div className="bg-white border border-slate-200 rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider font-medium border-b border-slate-200">
              <tr>
                <th className="px-5 py-3">Project & Domain</th>
                <th className="px-5 py-3">Runtime</th>
                <th className="px-5 py-3">SSL Status</th>
                <th className="px-5 py-3">Active Release</th>
                <th className="px-5 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filtered.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-5 py-12 text-center text-slate-400">
                    No projects found for the selected filter.
                  </td>
                </tr>
              ) : (
                filtered.map((project) => {
                  const activeRelease = project.releases?.[0]?.release_tag || 'none';
                  return (
                    <tr key={project.id} className="hover:bg-slate-50/70 transition-colors">
                      <td className="px-5 py-3.5">
                        <div className="font-semibold text-slate-900 text-sm">{project.name}</div>
                        <div className="flex items-center gap-2 mt-0.5">
                          <a
                            href={`http://${project.domain}`}
                            target="_blank"
                            rel="noreferrer"
                            className="text-slate-500 hover:text-blue-600 font-mono text-xs flex items-center gap-1"
                          >
                            <Globe className="w-3 h-3 text-slate-400" />
                            <span>{project.domain}</span>
                            <ExternalLink className="w-2.5 h-2.5 text-slate-400" />
                          </a>
                          {project.repository_url && (
                            <span
                              className="inline-flex items-center gap-1 text-[10px] font-mono text-slate-600 bg-slate-100 border border-slate-200 px-1.5 py-0.5 rounded"
                              title={project.repository_url}
                            >
                              <GitBranch className="w-2.5 h-2.5 text-blue-500" />
                              {project.branch || 'main'}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-5 py-3.5">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono font-medium bg-slate-100 text-slate-700 border border-slate-200">
                          {project.project_type}
                        </span>
                      </td>
                      <td className="px-5 py-3.5">
                        {project.ssl_enabled ? (
                          <span className="inline-flex items-center gap-1 text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded text-[11px] border border-emerald-200 font-medium">
                            <CheckCircle2 className="w-3 h-3" /> Let's Encrypt
                          </span>
                        ) : (
                          <span className="text-slate-400 text-xs">HTTP Only</span>
                        )}
                      </td>
                      <td className="px-5 py-3.5">
                        <div className="font-mono text-slate-700 text-xs bg-slate-50 px-2 py-0.5 rounded inline-block border border-slate-200">
                          {activeRelease}
                        </div>
                      </td>
                      <td className="px-5 py-3.5 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            onClick={() => openUpdateModal(project)}
                            title="Update from Repository"
                            className="p-1.5 text-slate-600 hover:text-emerald-600 hover:bg-emerald-50 rounded transition-colors"
                          >
                            <GitPullRequest className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => onDeploy(project.id)}
                            title="Deploy Release"
                            className="p-1.5 text-slate-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          >
                            <Rocket className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => onRollback(project.id)}
                            title="Rollback Release"
                            className="p-1.5 text-slate-600 hover:text-amber-600 hover:bg-amber-50 rounded transition-colors"
                          >
                            <RotateCcw className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => onViewLogs(project.slug)}
                            title="View Logs"
                            className="p-1.5 text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded transition-colors"
                          >
                            <ScrollText className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => onDelete(project.id)}
                            title="Delete Project"
                            className="p-1.5 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded transition-colors"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Add Project Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-lg border border-slate-200 shadow-xl max-w-lg w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <FolderGit2 className="w-4 h-4 text-blue-600" />
                <h3 className="text-sm font-bold text-slate-900">Provision New Project</h3>
              </div>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-4 text-xs">
              <div>
                <label className="block font-medium text-slate-700 mb-1">Project Name</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Landing Page or API Backend"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded focus:outline-none focus:border-blue-600"
                />
              </div>

              <div>
                <label className="block font-medium text-slate-700 mb-1">Domain Name</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. app.example.com"
                  value={domain}
                  onChange={(e) => setDomain(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded focus:outline-none focus:border-blue-600"
                />
              </div>

              <div>
                <label className="block font-medium text-slate-700 mb-1">Application Runtime</label>
                <select
                  value={projectType}
                  onChange={(e) => setProjectType(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded focus:outline-none focus:border-blue-600 bg-white"
                >
                  <option value="laravel">Laravel / PHP 8.4 (PHP-FPM)</option>
                  <option value="php_native">PHP Native</option>
                  <option value="nodejs">Node.js Daemon</option>
                  <option value="nextjs">Next.js SSR</option>
                  <option value="react_vite">React / Vite SPA</option>
                  <option value="python">Python (FastAPI / Flask)</option>
                  <option value="docker">Docker Container</option>
                  <option value="static">Static HTML</option>
                </select>
              </div>

              <div>
                <label className="block font-medium text-slate-700 mb-1">Git Repository URL (Optional)</label>
                <input
                  type="text"
                  placeholder="https://github.com/user/repo.git"
                  value={repoUrl}
                  onChange={(e) => setRepoUrl(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded focus:outline-none focus:border-blue-600 font-mono"
                />
              </div>

              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="enable_ssl"
                  checked={enableSSL}
                  onChange={(e) => setEnableSSL(e.target.checked)}
                  className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                />
                <label htmlFor="enable_ssl" className="text-slate-700 font-medium">
                  Automatically provision Let's Encrypt SSL certificate
                </label>
              </div>

              <div className="pt-4 border-t border-slate-100 flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 rounded border border-slate-300 text-slate-700 hover:bg-slate-50 font-medium"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white font-medium shadow-sm disabled:opacity-50"
                >
                  {isSubmitting ? 'Provisioning...' : 'Create & Configure'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Update from Repo Modal */}
      {isUpdateModalOpen && selectedProject && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-lg border border-slate-200 shadow-xl max-w-lg w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <GitPullRequest className="w-4 h-4 text-emerald-600" />
                <div>
                  <h3 className="text-sm font-bold text-slate-900">Update from Repository</h3>
                  <p className="text-[11px] text-slate-500">Atomic zero-downtime release pull & deploy</p>
                </div>
              </div>
              <button
                onClick={() => setIsUpdateModalOpen(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleUpdateFromRepo} className="p-6 space-y-4 text-xs">
              {/* Target Project Info Card */}
              <div className="bg-slate-50 border border-slate-200 rounded p-3 flex items-center justify-between">
                <div>
                  <div className="font-semibold text-slate-800 text-xs">{selectedProject.name}</div>
                  <div className="text-[11px] text-slate-500 font-mono mt-0.5">{selectedProject.domain}</div>
                </div>
                <div className="text-right">
                  <span className="inline-block text-[10px] font-mono font-medium px-2 py-0.5 rounded bg-white border border-slate-200 text-slate-700">
                    {selectedProject.project_type}
                  </span>
                </div>
              </div>

              {/* Error banner */}
              {updateError && (
                <div className="bg-rose-50 border border-rose-200 text-rose-700 px-3.5 py-2.5 rounded text-xs flex items-start gap-2">
                  <AlertCircle className="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
                  <div>
                    <div className="font-semibold">Update failed</div>
                    <div className="text-[11px] mt-0.5">{updateError}</div>
                  </div>
                </div>
              )}

              {/* Success Commit Banner */}
              {updateCommitInfo && (
                <div className="bg-emerald-50 border border-emerald-200 text-emerald-800 px-3.5 py-2.5 rounded text-xs flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                    <div>
                      <div className="font-semibold">Successfully Updated & Activated</div>
                      <div className="text-[11px] text-emerald-700 font-mono mt-0.5">
                        Release: {updateCommitInfo.tag} {updateCommitInfo.hash ? `(${updateCommitInfo.hash})` : ''}
                      </div>
                    </div>
                  </div>
                  <span className="text-[10px] font-mono uppercase px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 font-bold border border-emerald-300">
                    Active
                  </span>
                </div>
              )}

              {/* Repository URL Input */}
              <div>
                <label className="block font-medium text-slate-700 mb-1">
                  Git Repository URL
                </label>
                <input
                  type="text"
                  required
                  placeholder="https://github.com/user/repository.git"
                  value={updateRepoUrl}
                  onChange={(e) => setUpdateRepoUrl(e.target.value)}
                  disabled={isUpdating}
                  className="w-full px-3 py-2 border border-slate-300 rounded font-mono text-xs focus:outline-none focus:border-emerald-600 disabled:bg-slate-50"
                />
                <p className="text-[11px] text-slate-400 mt-1">
                  Public HTTPS or authenticated Git repository URL.
                </p>
              </div>

              {/* Branch Input */}
              <div>
                <label className="block font-medium text-slate-700 mb-1">
                  Git Branch
                </label>
                <div className="relative">
                  <input
                    type="text"
                    required
                    placeholder="main"
                    value={updateBranch}
                    onChange={(e) => setUpdateBranch(e.target.value)}
                    disabled={isUpdating}
                    className="w-full pl-7 pr-3 py-2 border border-slate-300 rounded font-mono text-xs focus:outline-none focus:border-emerald-600 disabled:bg-slate-50"
                  />
                  <GitBranch className="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2.5" />
                </div>
              </div>

              {/* Build script checkbox */}
              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="update_run_build"
                  checked={updateRunBuild}
                  onChange={(e) => setUpdateRunBuild(e.target.checked)}
                  disabled={isUpdating}
                  className="rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
                />
                <label htmlFor="update_run_build" className="text-slate-700 font-medium">
                  Run package managers & build scripts (composer, npm run build)
                </label>
              </div>

              {/* Terminal Logs Output */}
              {updateLogs && (
                <div className="mt-3">
                  <div className="text-[11px] font-semibold text-slate-600 mb-1 flex items-center gap-1.5">
                    <Terminal className="w-3.5 h-3.5 text-slate-500" />
                    <span>Deployment & Build Logs</span>
                  </div>
                  <pre className="bg-slate-950 text-slate-200 font-mono text-[11px] p-3 rounded border border-slate-800 max-h-48 overflow-y-auto leading-relaxed whitespace-pre-wrap select-text">
                    {updateLogs}
                  </pre>
                </div>
              )}

              {/* Modal Buttons */}
              <div className="pt-4 border-t border-slate-100 flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsUpdateModalOpen(false)}
                  className="px-4 py-2 rounded border border-slate-300 text-slate-700 hover:bg-slate-50 font-medium"
                >
                  {updateCommitInfo ? 'Close' : 'Cancel'}
                </button>
                <button
                  type="submit"
                  disabled={isUpdating}
                  className="flex items-center gap-1.5 px-4 py-2 rounded bg-emerald-600 hover:bg-emerald-700 text-white font-medium shadow-sm disabled:opacity-50 transition-colors"
                >
                  {isUpdating ? (
                    <>
                      <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                      <span>Pulling & Deploying...</span>
                    </>
                  ) : (
                    <>
                      <GitPullRequest className="w-3.5 h-3.5" />
                      <span>{updateCommitInfo ? 'Re-deploy from Repo' : 'Update & Deploy Now'}</span>
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  );
};
