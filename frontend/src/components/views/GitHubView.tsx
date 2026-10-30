import React, { useState } from 'react';
import {
  Github,
  GitBranch,
  GitPullRequest,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  Terminal,
  FolderGit2,
  Globe,
  X,
  ScrollText,
} from 'lucide-react';
import { ProjectItem } from '../../types';

interface GitHubViewProps {
  projects: ProjectItem[];
  onUpdateFromRepo: (
    id: string,
    options: { repository_url?: string; branch?: string; run_build: boolean }
  ) => Promise<{ logs: string; commit_hash?: string; release_tag?: string }>;
  onViewLogs: (slug: string) => void;
}

export const GitHubView: React.FC<GitHubViewProps> = ({
  projects,
  onUpdateFromRepo,
  onViewLogs,
}) => {
  const [selectedProject, setSelectedProject] = useState<ProjectItem | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [repoUrl, setRepoUrl] = useState('');
  const [branch, setBranch] = useState('main');
  const [runBuild, setRunBuild] = useState(true);
  const [isUpdating, setIsUpdating] = useState(false);
  const [logs, setLogs] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [successInfo, setSuccessInfo] = useState<{ hash?: string; tag?: string } | null>(null);

  const gitProjects = projects.filter((p) => p.repository_url || p.releases?.length);


  const openModal = (project: ProjectItem) => {
    setSelectedProject(project);
    setRepoUrl(project.repository_url || '');
    setBranch(project.branch || 'main');
    setRunBuild(true);
    setLogs(null);
    setError(null);
    setSuccessInfo(null);
    setIsModalOpen(true);
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject) return;

    setIsUpdating(true);
    setError(null);
    setLogs(null);

    try {
      const res = await onUpdateFromRepo(selectedProject.id, {
        repository_url: repoUrl || undefined,
        branch: branch || 'main',
        run_build: runBuild,
      });

      setLogs(res.logs);
      setSuccessInfo({
        hash: res.commit_hash,
        tag: res.release_tag,
      });
    } catch (err: any) {
      setError(err.message || 'Failed to update from repository.');
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* View Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight flex items-center gap-2">
            <Github className="w-5 h-5 text-slate-800" />
            <span>GitHub & Repository Integration</span>
          </h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Sync applications with remote Git repositories and execute atomic zero-downtime releases.
          </p>
        </div>
      </div>

      {/* Info Stats Bar */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white border border-slate-200 rounded-lg p-4">
          <div className="text-[11px] font-medium text-slate-400 uppercase tracking-wider">
            Connected Repositories
          </div>
          <div className="text-2xl font-bold text-slate-900 mt-1 font-mono">
            {gitProjects.length}
          </div>
          <div className="text-[11px] text-slate-500 mt-1 flex items-center gap-1">
            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" /> Ready for continuous deployment
          </div>
        </div>

        <div className="bg-white border border-slate-200 rounded-lg p-4">
          <div className="text-[11px] font-medium text-slate-400 uppercase tracking-wider">
            CLI Direct Update
          </div>
          <div className="text-xs font-mono font-semibold text-slate-800 mt-2 bg-slate-50 p-1.5 rounded border border-slate-200">
            vpsctl project update &lt;slug&gt;
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            Full zero-downtime release automation from shell
          </div>
        </div>

        <div className="bg-white border border-slate-200 rounded-lg p-4">
          <div className="text-[11px] font-medium text-slate-400 uppercase tracking-wider">
            Webhook Receiver
          </div>
          <div className="text-xs font-mono text-emerald-700 font-semibold mt-2 bg-emerald-50 p-1.5 rounded border border-emerald-200 flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            HMAC-SHA256 Active
          </div>
          <div className="text-[11px] text-slate-500 mt-1">
            Automated git push releases on branch update
          </div>
        </div>
      </div>

      {/* Git Projects Table */}
      <div className="bg-white border border-slate-200 rounded-lg overflow-hidden">
        <div className="px-5 py-3.5 border-b border-slate-200 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <FolderGit2 className="w-4 h-4 text-blue-600" />
            <h2 className="text-sm font-semibold text-slate-900">Repository Deployments</h2>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider font-medium border-b border-slate-200">
              <tr>
                <th className="px-5 py-3">Project & Domain</th>
                <th className="px-5 py-3">Git Repository</th>
                <th className="px-5 py-3">Branch</th>
                <th className="px-5 py-3">Active Release</th>
                <th className="px-5 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {gitProjects.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-5 py-12 text-center text-slate-400">
                    No Git-linked projects found. Link a repository in the Projects tab to start.
                  </td>
                </tr>
              ) : (
                gitProjects.map((project) => {
                  const activeRelease = project.releases?.[0]?.release_tag || 'none';
                  const commitHash = project.releases?.[0]?.commit_hash;

                  return (
                    <tr key={project.id} className="hover:bg-slate-50/70 transition-colors">
                      <td className="px-5 py-3.5">
                        <div className="font-semibold text-slate-900 text-sm">{project.name}</div>
                        <a
                          href={`http://${project.domain}`}
                          target="_blank"
                          rel="noreferrer"
                          className="text-slate-500 hover:text-blue-600 font-mono text-xs flex items-center gap-1 mt-0.5"
                        >
                          <Globe className="w-3 h-3 text-slate-400" />
                          <span>{project.domain}</span>
                        </a>
                      </td>

                      <td className="px-5 py-3.5">
                        {project.repository_url ? (
                          <div className="font-mono text-slate-700 text-xs flex items-center gap-1.5">
                            <Github className="w-3.5 h-3.5 text-slate-700 shrink-0" />
                            <span className="truncate max-w-xs">{project.repository_url}</span>
                          </div>
                        ) : (
                          <span className="text-slate-400 italic">No remote URL set</span>
                        )}
                      </td>

                      <td className="px-5 py-3.5">
                        <span className="inline-flex items-center gap-1 text-[11px] font-mono font-medium px-2 py-0.5 rounded bg-slate-100 text-slate-700 border border-slate-200">
                          <GitBranch className="w-3 h-3 text-blue-500" />
                          {project.branch || 'main'}
                        </span>
                      </td>

                      <td className="px-5 py-3.5">
                        <div className="space-y-0.5">
                          <div className="font-mono text-slate-800 text-xs font-semibold">
                            {activeRelease}
                          </div>
                          {commitHash && (
                            <div className="font-mono text-[11px] text-slate-500">
                              commit: <span className="text-blue-600">{commitHash}</span>
                            </div>
                          )}
                        </div>
                      </td>

                      <td className="px-5 py-3.5 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <button
                            onClick={() => onViewLogs(project.slug)}
                            title="View Deployment Logs"
                            className="p-1.5 text-slate-500 hover:text-slate-800 hover:bg-slate-100 rounded transition-colors"
                          >
                            <ScrollText className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => openModal(project)}
                            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm transition-colors"
                          >
                            <GitPullRequest className="w-3.5 h-3.5" />
                            <span>Update from Repo</span>
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

      {/* Update from Repo Modal */}
      {isModalOpen && selectedProject && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-lg border border-slate-200 shadow-xl max-w-lg w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <GitPullRequest className="w-4 h-4 text-emerald-600" />
                <div>
                  <h3 className="text-sm font-bold text-slate-900">Update from Repository</h3>
                  <p className="text-[11px] text-slate-500">Pull latest git commits and activate release</p>
                </div>
              </div>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleUpdate} className="p-6 space-y-4 text-xs">
              <div className="bg-slate-50 border border-slate-200 rounded p-3 flex items-center justify-between">
                <div>
                  <div className="font-semibold text-slate-800 text-xs">{selectedProject.name}</div>
                  <div className="text-[11px] text-slate-500 font-mono mt-0.5">{selectedProject.domain}</div>
                </div>
                <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-white border border-slate-200 text-slate-700">
                  {selectedProject.project_type}
                </span>
              </div>

              {error && (
                <div className="bg-rose-50 border border-rose-200 text-rose-700 px-3.5 py-2.5 rounded text-xs flex items-start gap-2">
                  <AlertCircle className="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
                  <div>
                    <div className="font-semibold">Update failed</div>
                    <div className="text-[11px] mt-0.5">{error}</div>
                  </div>
                </div>
              )}

              {successInfo && (
                <div className="bg-emerald-50 border border-emerald-200 text-emerald-800 px-3.5 py-2.5 rounded text-xs flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                    <div>
                      <div className="font-semibold">Release Deployed Successfully</div>
                      <div className="text-[11px] text-emerald-700 font-mono mt-0.5">
                        Tag: {successInfo.tag} {successInfo.hash ? `(${successInfo.hash})` : ''}
                      </div>
                    </div>
                  </div>
                  <span className="text-[10px] font-mono uppercase px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 font-bold border border-emerald-300">
                    Active
                  </span>
                </div>
              )}

              <div>
                <label className="block font-medium text-slate-700 mb-1">
                  Git Repository URL
                </label>
                <input
                  type="text"
                  required
                  placeholder="https://github.com/user/repository.git"
                  value={repoUrl}
                  onChange={(e) => setRepoUrl(e.target.value)}
                  disabled={isUpdating}
                  className="w-full px-3 py-2 border border-slate-300 rounded font-mono text-xs focus:outline-none focus:border-emerald-600 disabled:bg-slate-50"
                />
              </div>

              <div>
                <label className="block font-medium text-slate-700 mb-1">
                  Git Branch
                </label>
                <div className="relative">
                  <input
                    type="text"
                    required
                    placeholder="main"
                    value={branch}
                    onChange={(e) => setBranch(e.target.value)}
                    disabled={isUpdating}
                    className="w-full pl-7 pr-3 py-2 border border-slate-300 rounded font-mono text-xs focus:outline-none focus:border-emerald-600 disabled:bg-slate-50"
                  />
                  <GitBranch className="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2.5" />
                </div>
              </div>

              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="gh_update_run_build"
                  checked={runBuild}
                  onChange={(e) => setRunBuild(e.target.checked)}
                  disabled={isUpdating}
                  className="rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
                />
                <label htmlFor="gh_update_run_build" className="text-slate-700 font-medium">
                  Run package managers & build scripts (composer, npm run build)
                </label>
              </div>

              {logs && (
                <div className="mt-3">
                  <div className="text-[11px] font-semibold text-slate-600 mb-1 flex items-center gap-1.5">
                    <Terminal className="w-3.5 h-3.5 text-slate-500" />
                    <span>Deployment & Build Logs</span>
                  </div>
                  <pre className="bg-slate-950 text-slate-200 font-mono text-[11px] p-3 rounded border border-slate-800 max-h-48 overflow-y-auto leading-relaxed whitespace-pre-wrap select-text">
                    {logs}
                  </pre>
                </div>
              )}

              <div className="pt-4 border-t border-slate-100 flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 rounded border border-slate-300 text-slate-700 hover:bg-slate-50 font-medium"
                >
                  {successInfo ? 'Close' : 'Cancel'}
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
                      <span>{successInfo ? 'Re-deploy from Repo' : 'Update & Deploy Now'}</span>
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
