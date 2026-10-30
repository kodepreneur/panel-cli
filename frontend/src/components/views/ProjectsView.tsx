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
  onRollback: (id: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onViewLogs: (slug: string) => void;
}

export const ProjectsView: React.FC<ProjectsViewProps> = ({
  projects,
  onCreateProject,
  onDeploy,
  onRollback,
  onDelete,
  onViewLogs,
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [filterType, setFilterType] = useState<string>('all');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form states
  const [name, setName] = useState('');
  const [domain, setDomain] = useState('');
  const [projectType, setProjectType] = useState('laravel');
  const [repoUrl, setRepoUrl] = useState('');
  const [branch] = useState('main');
  const [enableSSL, setEnableSSL] = useState(true);

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
                        <a
                          href={`http://${project.domain}`}
                          target="_blank"
                          rel="noreferrer"
                          className="text-slate-500 hover:text-blue-600 font-mono text-xs flex items-center gap-1 mt-0.5"
                        >
                          <Globe className="w-3 h-3 text-slate-400" />
                          <span>{project.domain}</span>
                          <ExternalLink className="w-2.5 h-2.5 text-slate-400" />
                        </a>
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
    </div>
  );
};
