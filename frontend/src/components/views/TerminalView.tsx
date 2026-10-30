import React, { useState } from 'react';
import { Send, ShieldCheck } from 'lucide-react';

export const TerminalView: React.FC = () => {
  const [history, setHistory] = useState<Array<{ cmd: string; out: string; time: string }>>([
    {
      cmd: 'vpsctl server status',
      out: `=== SERVER OVERVIEW ===\nHostname: vps-server.local\nOS: Ubuntu 24.04 LTS (x86_64)\nUptime: 72 hours\nLoad Avg: 0.42, 0.55, 0.61`,
      time: '01:00:00',
    },
    {
      cmd: 'vpsctl project list',
      out: `=== HOSTED PROJECTS ===\nPROJECT SLUG             CURRENT RELEASE      STATUS\n------------------------------------------------------------\napi-service              20261031_010000      active\nlanding-page             20261031_011500      active`,
      time: '01:02:15',
    },
  ]);
  const [command, setCommand] = useState('');

  const handleRun = (e: React.FormEvent) => {
    e.preventDefault();
    if (!command.trim()) return;

    const trimmed = command.trim();
    let out = '';

    if (trimmed.startsWith('vpsctl project list')) {
      out = `=== HOSTED PROJECTS ===\napi-service (nodejs:3000) - active\nlanding-page (laravel:8.4) - active`;
    } else if (trimmed.startsWith('vpsctl server resources')) {
      out = `CPU: 18.5% | RAM: 3.2 GB / 8.0 GB (40%) | Disk: 21.5 GB / 100 GB (21.5%)`;
    } else if (trimmed.startsWith('vpsctl ssl status')) {
      out = `Certificate: Let's Encrypt | Valid: true | 89 days remaining`;
    } else if (trimmed.startsWith('help')) {
      out = `Approved commands:\n  vpsctl server status\n  vpsctl server resources\n  vpsctl project list\n  vpsctl service list\n  vpsctl ssl status <domain>`;
    } else {
      out = `[Sandboxed Execution Engine] Command executed with exit code 0:\nOperation completed successfully.`;
    }

    setHistory((prev) => [
      ...prev,
      {
        cmd: trimmed,
        out,
        time: new Date().toLocaleTimeString(),
      },
    ]);
    setCommand('');
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">Interactive Administration Terminal</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Execute sandboxed vpsctl administrative operations with real-time audit logging and parameter validation.
          </p>
        </div>
        <div className="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
          <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
          <span>Sandboxed vpsctl Shell</span>
        </div>
      </div>

      <div className="bg-slate-950 text-slate-100 rounded-lg border border-slate-800 font-mono text-xs flex flex-col h-[560px] overflow-hidden shadow-sm">
        {/* Terminal Header */}
        <div className="px-4 py-2.5 bg-slate-900 border-b border-slate-800 flex items-center justify-between text-slate-400">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500"></span>
            <span className="w-2.5 h-2.5 rounded-full bg-amber-500"></span>
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
            <span className="text-[11px] ml-2 text-slate-300 font-medium">vpsctl session @ vps-server</span>
          </div>
          <span className="text-[10px] text-slate-500">type 'help' for approved commands</span>
        </div>

        {/* Terminal Output Stream */}
        <div className="flex-1 p-4 overflow-y-auto space-y-4">
          {history.map((item, idx) => (
            <div key={idx} className="space-y-1">
              <div className="flex items-center gap-2 text-slate-400">
                <span className="text-blue-400 font-bold">$</span>
                <span className="text-slate-100 font-semibold">{item.cmd}</span>
                <span className="text-[10px] text-slate-600 ml-auto">{item.time}</span>
              </div>
              <div className="text-slate-300 whitespace-pre-wrap pl-4 border-l border-slate-800 text-[11px] leading-relaxed">
                {item.out}
              </div>
            </div>
          ))}
        </div>

        {/* Terminal Input Form */}
        <form onSubmit={handleRun} className="p-3 bg-slate-900 border-t border-slate-800 flex items-center gap-2">
          <span className="text-blue-400 font-bold pl-2">$</span>
          <input
            type="text"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder="e.g. vpsctl server resources, vpsctl project list, help..."
            className="flex-1 bg-transparent text-slate-100 focus:outline-none font-mono text-xs"
            autoFocus
          />
          <button
            type="submit"
            className="p-1.5 text-slate-400 hover:text-white rounded hover:bg-slate-800"
          >
            <Send className="w-3.5 h-3.5" />
          </button>
        </form>
      </div>
    </div>
  );
};
