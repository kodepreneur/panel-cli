import React, { useState } from 'react';
import { CheckCircle2, Clock, GitCommit, ScrollText, ArrowRight } from 'lucide-react';
import { DeploymentItem } from '../../types';

interface DeploymentsViewProps {
  deployments: DeploymentItem[];
  selectedLogs: string | null;
}

export const DeploymentsView: React.FC<DeploymentsViewProps> = ({
  deployments,
  selectedLogs,
}) => {
  const [activeLog, setActiveLog] = useState<string | null>(selectedLogs);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">Deployment Pipeline & History</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Zero-downtime release deployments triggered via GitHub Webhook, API, CLI, or AI Agents.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Deployments List (2 Cols) */}
        <div className="lg:col-span-2 bg-white border border-slate-200 rounded-lg overflow-hidden">
          <div className="px-5 py-3.5 border-b border-slate-200">
            <h2 className="text-sm font-semibold text-slate-900">Recent Executions</h2>
          </div>

          <div className="divide-y divide-slate-100">
            {deployments.map((dep) => (
              <div
                key={dep.id}
                onClick={() => setActiveLog(dep.build_logs)}
                className="p-4 hover:bg-slate-50 cursor-pointer transition-colors flex items-center justify-between text-xs"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-slate-900">{dep.project_slug}</span>
                    <span className="font-mono text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded text-[11px]">
                      {dep.release_tag}
                    </span>
                    <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200">
                      {dep.trigger_type}
                    </span>
                  </div>
                  <div className="mt-1 text-slate-500 flex items-center gap-3">
                    <span className="flex items-center gap-1 font-mono text-[11px]">
                      <GitCommit className="w-3 h-3 text-slate-400" /> main (head)
                    </span>
                    <span className="flex items-center gap-1 text-[11px]">
                      <Clock className="w-3 h-3 text-slate-400" /> {dep.duration_seconds}s duration
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <span className="inline-flex items-center gap-1 text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded text-[11px] font-medium border border-emerald-200">
                    <CheckCircle2 className="w-3 h-3" /> success
                  </span>
                  <ArrowRight className="w-4 h-4 text-slate-300" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Live Build Logs Console (1 Col) */}
        <div className="bg-slate-950 text-slate-200 rounded-lg border border-slate-800 p-4 font-mono text-xs flex flex-col h-[520px]">
          <div className="flex items-center justify-between pb-3 border-b border-slate-800 text-slate-400">
            <span className="flex items-center gap-1.5 font-semibold text-slate-200">
              <ScrollText className="w-3.5 h-3.5 text-blue-400" /> Build Log Output
            </span>
            <span className="text-[10px] text-emerald-400 flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
              live stream
            </span>
          </div>

          <div className="flex-1 overflow-y-auto mt-3 space-y-1 text-slate-300 text-[11px] leading-relaxed whitespace-pre-wrap select-text">
            {activeLog ||
`[vps-deploy] Initializing deployment runner...
[vps-deploy] Target path: /var/www/vps-panel/projects/api-service/releases/20261031_010000
[vps-deploy] Linking shared/.env -> releases/20261031_010000/.env
[vps-deploy] Linking shared/storage -> releases/20261031_010000/storage
[vps-deploy] Running dependency installation...
[vps-deploy] npm install --production --silent (completed in 2.1s)
[vps-deploy] Atomic symlink swap: current.tmp -> releases/20261031_010000
[vps-deploy] os.Rename("current.tmp", "current") - atomic POSIX switch OK
[vps-deploy] Pruning releases older than keep_count(5)...
[vps-deploy] Deployment completed successfully in 3.42s.`}
          </div>
        </div>
      </div>
    </div>
  );
};
