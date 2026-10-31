import React from 'react';
import {
  Cpu,
  HardDrive,
  Activity,
  Layers,
  ArrowUpRight,
  ShieldCheck,
  Clock,
  Plus,
} from 'lucide-react';
import { ServerOverview, ProjectItem, ServiceItem } from '../../types';

interface OverviewViewProps {
  overview?: ServerOverview;
  projects: ProjectItem[];
  services: ServiceItem[];
  onAddProject: () => void;
  onDeployProject: (id: string) => void;
  onSelectTab: (tab: any) => void;
}

export const OverviewView: React.FC<OverviewViewProps> = ({
  overview,
  projects,
  services,
  onAddProject,
  onDeployProject,
  onSelectTab,
}) => {
  const cpuUsage = Number(overview?.cpu?.usage_percent ?? 0);
  const memTotalBytes = Number(overview?.memory?.total_bytes ?? 0);
  const memUsedBytes = Number(overview?.memory?.used_bytes ?? 0);
  const memUsedGB = memTotalBytes > 0 ? (memUsedBytes / (1024 * 1024 * 1024)).toFixed(1) : '0.0';
  const memTotalGB = memTotalBytes > 0 ? (memTotalBytes / (1024 * 1024 * 1024)).toFixed(1) : '0.0';
  const memUsagePct = Number(overview?.memory?.usage_percent ?? 0);

  const disk = overview?.disks?.[0];
  const diskTotalBytes = Number(disk?.total_bytes ?? 0);
  const diskUsedBytes = Number(disk?.used_bytes ?? 0);
  const diskUsagePct = Number(disk?.usage_percent ?? 0);
  const diskUsedGB = diskTotalBytes > 0 ? (diskUsedBytes / (1024 * 1024 * 1024)).toFixed(1) : '0.0';
  const diskTotalGB = diskTotalBytes > 0 ? (diskTotalBytes / (1024 * 1024 * 1024)).toFixed(1) : '0.0';

  const sslCount = projects.filter((p) => p.ssl_enabled).length;

  return (
    <div className="space-y-6">
      {/* Page Title & Top CTA */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">Infrastructure Overview</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Real-time host metrics, active virtual hosts, and AI agent execution status.
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => onSelectTab('agents')}
            className="px-3 py-1.5 text-xs font-medium rounded border border-slate-300 text-slate-700 bg-white hover:bg-slate-50 transition-colors"
          >
            Manage AI Agents
          </button>
          <button
            onClick={onAddProject}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded bg-blue-600 hover:bg-blue-700 text-white shadow-sm transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Add Project</span>
          </button>
        </div>
      </div>

      {/* Core Resource Metrics Strip */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {/* CPU Metric Card */}
        <div className="bg-white border border-slate-200 rounded-lg p-4">
          <div className="flex items-center justify-between text-slate-500 text-xs font-medium">
            <span className="flex items-center gap-1.5">
              <Cpu className="w-3.5 h-3.5 text-slate-400" /> CPU Usage
            </span>
            <span className="font-mono text-slate-400">{overview?.cpu?.cores ?? '--'} cores</span>
          </div>
          <div className="mt-2 flex items-baseline justify-between">
            <span className="text-2xl font-bold font-mono text-slate-900">{overview?.cpu ? cpuUsage.toFixed(1) + '%' : '--'}</span>
            <span className="text-xs text-slate-400 font-mono">
              load: {overview?.load_avg && overview.load_avg[0] != null && !isNaN(Number(overview.load_avg[0])) ? Number(overview.load_avg[0]).toFixed(2) : '--'}
            </span>
          </div>
          <div className="w-full bg-slate-100 h-1.5 rounded-full mt-3 overflow-hidden">
            <div
              className={`h-full rounded-full ${cpuUsage > 80 ? 'bg-rose-500' : 'bg-blue-600'}`}
              style={{ width: `${Math.min(cpuUsage, 100)}%` }}
            ></div>
          </div>
        </div>

        {/* Memory Metric Card */}
        <div className="bg-white border border-slate-200 rounded-lg p-4">
          <div className="flex items-center justify-between text-slate-500 text-xs font-medium">
            <span className="flex items-center gap-1.5">
              <Activity className="w-3.5 h-3.5 text-slate-400" /> RAM Utilization
            </span>
            <span className="font-mono text-slate-400">{memUsagePct.toFixed(0)}%</span>
          </div>
          <div className="mt-2 flex items-baseline justify-between">
            <span className="text-2xl font-bold font-mono text-slate-900">
              {memUsedGB} <span className="text-xs text-slate-500 font-normal">/ {memTotalGB} GB</span>
            </span>
          </div>
          <div className="w-full bg-slate-100 h-1.5 rounded-full mt-3 overflow-hidden">
            <div
              className={`h-full rounded-full ${memUsagePct > 85 ? 'bg-rose-500' : 'bg-blue-600'}`}
              style={{ width: `${Math.min(memUsagePct, 100)}%` }}
            ></div>
          </div>
        </div>

        {/* Disk Space Metric Card */}
        <div className="bg-white border border-slate-200 rounded-lg p-4">
          <div className="flex items-center justify-between text-slate-500 text-xs font-medium">
            <span className="flex items-center gap-1.5">
              <HardDrive className="w-3.5 h-3.5 text-slate-400" /> Disk Space (/)
            </span>
            <span className="font-mono text-slate-400">{diskUsagePct.toFixed(0)}%</span>
          </div>
          <div className="mt-2 flex items-baseline justify-between">
            <span className="text-2xl font-bold font-mono text-slate-900">
              {diskUsedGB} <span className="text-xs text-slate-500 font-normal">/ {diskTotalGB} GB</span>
            </span>
          </div>
          <div className="w-full bg-slate-100 h-1.5 rounded-full mt-3 overflow-hidden">
            <div
              className={`h-full rounded-full ${diskUsagePct > 90 ? 'bg-rose-500' : 'bg-emerald-600'}`}
              style={{ width: `${Math.min(diskUsagePct, 100)}%` }}
            ></div>
          </div>
        </div>

        {/* Web & Security Metric Card */}
        <div className="bg-white border border-slate-200 rounded-lg p-4">
          <div className="flex items-center justify-between text-slate-500 text-xs font-medium">
            <span className="flex items-center gap-1.5">
              <Layers className="w-3.5 h-3.5 text-slate-400" /> Web & SSL
            </span>
            <span className="font-mono text-emerald-600 flex items-center gap-1">
              <ShieldCheck className="w-3.5 h-3.5" /> active
            </span>
          </div>
          <div className="mt-2 flex items-baseline justify-between">
            <span className="text-2xl font-bold font-mono text-slate-900">
              {projects.length} <span className="text-xs text-slate-500 font-normal">sites</span>
            </span>
            <span className="text-xs font-mono text-emerald-600">{sslCount} HTTPS</span>
          </div>
          <div className="w-full bg-slate-100 h-1.5 rounded-full mt-3 overflow-hidden">
            <div className="h-full rounded-full bg-emerald-500" style={{ width: '100%' }}></div>
          </div>
        </div>
      </div>

      {/* Main Grid: Projects Table & Side Infrastructure Status */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active Projects Table (2 Cols) */}
        <div className="lg:col-span-2 bg-white border border-slate-200 rounded-lg overflow-hidden">
          <div className="px-5 py-3.5 border-b border-slate-200 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-900">Active Websites & Projects</h2>
            <button
              onClick={() => onSelectTab('projects')}
              className="text-xs text-blue-600 hover:text-blue-700 font-medium flex items-center gap-1"
            >
              View all ({projects.length}) <ArrowUpRight className="w-3 h-3" />
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider font-medium border-b border-slate-200">
                <tr>
                  <th className="px-5 py-3">Project / Domain</th>
                  <th className="px-5 py-3">Runtime</th>
                  <th className="px-5 py-3">Latest Release</th>
                  <th className="px-5 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {projects.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-5 py-8 text-center text-slate-400">
                      No projects configured. Click "Add Project" to launch your first website.
                    </td>
                  </tr>
                ) : (
                  projects.slice(0, 5).map((project) => {
                    const activeRelease = project.releases?.[0]?.release_tag || 'initial';
                    return (
                      <tr key={project.id} className="hover:bg-slate-50/80 transition-colors">
                        <td className="px-5 py-3">
                          <div className="font-semibold text-slate-900">{project.name}</div>
                          <div className="text-slate-500 font-mono text-[11px] flex items-center gap-1">
                            {project.ssl_enabled ? (
                              <span className="text-emerald-600">https://</span>
                            ) : (
                              <span className="text-slate-400">http://</span>
                            )}
                            {project.domain}
                          </div>
                        </td>
                        <td className="px-5 py-3">
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono font-medium bg-slate-100 text-slate-700 border border-slate-200">
                            {project.project_type}
                          </span>
                        </td>
                        <td className="px-5 py-3">
                          <span className="font-mono text-slate-600 text-[11px] bg-slate-50 px-1.5 py-0.5 rounded border border-slate-200">
                            {activeRelease}
                          </span>
                        </td>
                        <td className="px-5 py-3 text-right">
                          <button
                            onClick={() => onDeployProject(project.id)}
                            className="text-xs text-blue-600 hover:text-blue-800 font-medium px-2 py-1 rounded hover:bg-blue-50"
                          >
                            Deploy
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

        {/* Host Services Status (1 Col) */}
        <div className="bg-white border border-slate-200 rounded-lg p-5 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <h2 className="text-sm font-semibold text-slate-900">System Services</h2>
              <button
                onClick={() => onSelectTab('services')}
                className="text-xs text-blue-600 hover:text-blue-700 font-medium"
              >
                Manage
              </button>
            </div>

            <div className="mt-4 space-y-3">
              {services.map((s) => {
                const isActive = s.status === 'active' || s.running;
                return (
                  <div key={s.name} className="flex items-center justify-between text-xs">
                    <div className="flex items-center gap-2">
                      <span
                        className={`w-2 h-2 rounded-full ${
                          isActive ? 'bg-emerald-500' : 'bg-slate-300'
                        }`}
                      />
                      <span className="text-slate-800 font-medium">{s.display_name}</span>
                    </div>
                    <span
                      className={`font-mono px-2 py-0.5 rounded text-[10px] uppercase font-semibold border ${
                        isActive
                          ? 'text-emerald-700 bg-emerald-50 border-emerald-200'
                          : 'text-slate-500 bg-slate-100 border-slate-200'
                      }`}
                    >
                      {s.status}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="mt-6 pt-4 border-t border-slate-100 text-xs text-slate-500">
            <div className="flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5 text-slate-400" />
              <span>
                System Uptime: {Math.floor((overview?.uptime_seconds || 72000) / 3600)} hours
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

