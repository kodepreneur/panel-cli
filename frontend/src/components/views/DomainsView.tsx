import React, { useState } from 'react';
import { Globe, RefreshCw, CheckCircle2, AlertTriangle } from 'lucide-react';
import { ProjectItem } from '../../types';

interface DomainsViewProps {
  projects: ProjectItem[];
  onInstallSSL: (projectId: string) => Promise<void>;
}

export const DomainsView: React.FC<DomainsViewProps> = ({ projects, onInstallSSL }) => {
  const [loadingMap, setLoadingMap] = useState<Record<string, boolean>>({});

  const handleInstall = async (projectId: string) => {
    setLoadingMap((prev) => ({ ...prev, [projectId]: true }));
    try {
      await onInstallSSL(projectId);
    } finally {
      setLoadingMap((prev) => ({ ...prev, [projectId]: false }));
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">Domains & SSL Certificates</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Automated Let's Encrypt certificate issuance, HTTP-01 challenge routing, and auto-renewals.
          </p>
        </div>
      </div>

      <div className="bg-white border border-slate-200 rounded-lg overflow-hidden">
        <div className="px-5 py-3.5 border-b border-slate-200 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-slate-900">Configured Domains</h2>
          <span className="text-xs text-slate-400 font-mono">Let's Encrypt ACME v2</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider font-medium border-b border-slate-200">
              <tr>
                <th className="px-5 py-3">Domain Name</th>
                <th className="px-5 py-3">Associated Project</th>
                <th className="px-5 py-3">Certificate Issuer</th>
                <th className="px-5 py-3">SSL Status</th>
                <th className="px-5 py-3">Auto Renewal</th>
                <th className="px-5 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {projects.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-5 py-8 text-center text-slate-400">
                    No domains configured. Create a project to link a domain.
                  </td>
                </tr>
              ) : (
                projects.map((project) => {
                  const isLoading = loadingMap[project.id];
                  return (
                    <tr key={project.id} className="hover:bg-slate-50/70 transition-colors">
                      <td className="px-5 py-3.5 font-semibold text-slate-900 font-mono flex items-center gap-2">
                        <Globe className="w-3.5 h-3.5 text-slate-400" />
                        <span>{project.domain}</span>
                      </td>
                      <td className="px-5 py-3.5 text-slate-600">{project.name}</td>
                      <td className="px-5 py-3.5 text-slate-600">
                        {project.ssl_enabled ? "Let's Encrypt Authority" : 'None'}
                      </td>
                      <td className="px-5 py-3.5">
                        {project.ssl_enabled ? (
                          <span className="inline-flex items-center gap-1 text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded text-[11px] font-medium border border-emerald-200">
                            <CheckCircle2 className="w-3 h-3" /> Valid (89 days left)
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-amber-700 bg-amber-50 px-2 py-0.5 rounded text-[11px] font-medium border border-amber-200">
                            <AlertTriangle className="w-3 h-3" /> Insecure (No SSL)
                          </span>
                        )}
                      </td>
                      <td className="px-5 py-3.5">
                        <span className="text-emerald-600 font-mono text-[11px]">Active (cron)</span>
                      </td>
                      <td className="px-5 py-3.5 text-right">
                        <button
                          onClick={() => handleInstall(project.id)}
                          disabled={isLoading}
                          className="flex items-center gap-1 px-3 py-1 text-xs font-medium rounded border border-slate-300 hover:bg-slate-50 text-slate-700 transition-colors ml-auto disabled:opacity-50"
                        >
                          <RefreshCw className={`w-3 h-3 ${isLoading ? 'animate-spin' : ''}`} />
                          <span>{project.ssl_enabled ? 'Renew SSL' : 'Install SSL'}</span>
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
