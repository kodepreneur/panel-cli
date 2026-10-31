import React from 'react';
import { RefreshCw, Server, ShieldCheck, Terminal as TerminalIcon, GitPullRequest } from 'lucide-react';
import { ServerOverview } from '../types';

interface HeaderProps {
  overview?: ServerOverview;
  onRefresh: () => void;
  isRefreshing: boolean;
  onOpenCLIHint: () => void;
  onOpenSystemUpdate?: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  overview,
  onRefresh,
  isRefreshing,
  onOpenCLIHint,
  onOpenSystemUpdate,
}) => {
  return (
    <header className="h-16 bg-white border-b border-slate-200 px-8 flex items-center justify-between">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-sm text-slate-700 font-medium">
          <Server className="w-4 h-4 text-slate-400" />
          <span className="font-semibold text-slate-900">
            {overview?.os?.hostname || 'vps-server.local'}
          </span>
          <span className="text-slate-300">|</span>
          <span className="text-slate-500 font-normal">
            {overview?.os?.version || 'Ubuntu 24.04 LTS'}
          </span>
          <span className="text-xs px-2 py-0.5 rounded bg-slate-100 text-slate-600 font-mono">
            {overview?.os?.arch || 'x86_64'}
          </span>
        </div>
      </div>

      <div className="flex items-center gap-3">
        {onOpenSystemUpdate && (
          <button
            onClick={onOpenSystemUpdate}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded border border-slate-200 text-slate-700 bg-white hover:bg-slate-50 transition-colors shadow-sm"
            title="Check for panel updates or update from repository"
          >
            <GitPullRequest className="w-3.5 h-3.5 text-emerald-600" />
            <span>Update Panel</span>
          </button>
        )}

        <button
          onClick={onOpenCLIHint}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono font-medium rounded border border-slate-200 text-slate-700 bg-slate-50 hover:bg-slate-100 transition-colors"
        >
          <TerminalIcon className="w-3.5 h-3.5 text-slate-500" />
          <span>vpsctl cli ready</span>
        </button>

        <div className="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
          <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
          <span>Daemon Active</span>
        </div>

        <button
          onClick={onRefresh}
          disabled={isRefreshing}
          className="p-2 text-slate-500 hover:text-slate-800 rounded hover:bg-slate-100 transition-colors"
          title="Refresh Metrics"
        >
          <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
        </button>
      </div>
    </header>
  );
};


