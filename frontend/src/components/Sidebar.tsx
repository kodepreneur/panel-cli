import React from 'react';
import {
  LayoutDashboard,
  FolderGit2,
  Globe,
  Github,
  Rocket,
  Database,
  Archive,
  Cpu,
  Terminal,
  Bot,
  ScrollText,
  Settings,
  ShieldAlert,
} from 'lucide-react';

export type NavItem =
  | 'overview'
  | 'projects'
  | 'domains'
  | 'github'
  | 'deployments'
  | 'databases'
  | 'backups'
  | 'services'
  | 'terminal'
  | 'agents'
  | 'activity'
  | 'settings';

interface SidebarProps {
  currentTab: NavItem;
  onSelectTab: (tab: NavItem) => void;
  pendingApprovalsCount: number;
}

export const Sidebar: React.FC<SidebarProps> = ({
  currentTab,
  onSelectTab,
  pendingApprovalsCount,
}) => {
  const navItems: Array<{ id: NavItem; label: string; icon: React.FC<any>; badge?: number }> = [
    { id: 'overview', label: 'Overview', icon: LayoutDashboard },
    { id: 'projects', label: 'Projects', icon: FolderGit2 },
    { id: 'domains', label: 'Domains & SSL', icon: Globe },
    { id: 'github', label: 'GitHub', icon: Github },
    { id: 'deployments', label: 'Deployments', icon: Rocket },
    { id: 'databases', label: 'Databases', icon: Database },
    { id: 'backups', label: 'Backups', icon: Archive },
    { id: 'services', label: 'Services', icon: Cpu },
    { id: 'terminal', label: 'Terminal', icon: Terminal },
    {
      id: 'agents',
      label: 'AI Agents',
      icon: Bot,
      badge: pendingApprovalsCount > 0 ? pendingApprovalsCount : undefined,
    },
    { id: 'activity', label: 'Activity Logs', icon: ScrollText },
    { id: 'settings', label: 'Settings', icon: Settings },
  ];

  return (
    <aside className="w-64 bg-slate-900 text-slate-300 flex flex-col h-screen border-r border-slate-800 select-none">
      {/* Brand Header */}
      <div className="h-16 flex items-center px-6 gap-3 border-b border-slate-800">
        <div className="w-8 h-8 rounded bg-blue-600 flex items-center justify-center text-white font-bold text-sm shadow-sm">
          VP
        </div>
        <div>
          <div className="text-white font-semibold text-sm tracking-tight leading-tight">
            VPS Panel
          </div>
          <div className="text-xs text-slate-400 font-mono">v1.0.0 • AI-Native</div>
        </div>
      </div>

      {/* Nav List */}
      <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = currentTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onSelectTab(item.id)}
              className={`w-full flex items-center justify-between px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                isActive
                  ? 'bg-blue-600/15 text-blue-400 border border-blue-500/30'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <div className="flex items-center gap-3">
                <Icon className={`w-4 h-4 ${isActive ? 'text-blue-400' : 'text-slate-400'}`} />
                <span>{item.label}</span>
              </div>
              {item.badge !== undefined && (
                <span className="flex items-center gap-1 bg-amber-500/20 text-amber-400 text-xs px-2 py-0.5 rounded-full font-mono font-semibold border border-amber-500/30">
                  <ShieldAlert className="w-3 h-3" />
                  {item.badge}
                </span>
              )}
            </button>
          );
        })}
      </nav>

      {/* Footer Status */}
      <div className="p-4 border-t border-slate-800 bg-slate-950/40 text-xs text-slate-400">
        <div className="flex items-center justify-between">
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            vpsd connected
          </span>
          <span className="font-mono text-slate-400">0.42 load</span>
        </div>
      </div>
    </aside>
  );
};

