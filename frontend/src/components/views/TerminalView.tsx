import React, { useState } from 'react';
import { Send, ShieldCheck, Loader2 } from 'lucide-react';
import { api } from '../../api/client';

export const TerminalView: React.FC = () => {
  const [history, setHistory] = useState<Array<{ cmd: string; out: string; time: string }>>([
    {
      cmd: 'help',
      out: `=== vpsctl Sandboxed Terminal ===\nConnected to live control plane.\nType 'vpsctl server status', 'vpsctl server resources', 'vpsctl project list', or 'help' to begin.`,
      time: new Date().toLocaleTimeString(),
    },
  ]);
  const [command, setCommand] = useState('');
  const [isRunning, setIsRunning] = useState(false);

  const handleRun = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = command.trim();
    if (!trimmed || isRunning) return;

    setCommand('');
    setIsRunning(true);

    try {
      const res = await api.executeTerminal(trimmed);
      setHistory((prev) => [
        ...prev,
        {
          cmd: trimmed,
          out: res.output || '(No output returned)',
          time: new Date().toLocaleTimeString(),
        },
      ]);
    } catch (err: any) {
      setHistory((prev) => [
        ...prev,
        {
          cmd: trimmed,
          out: `Error: ${err.message || 'Execution failed'}`,
          time: new Date().toLocaleTimeString(),
        },
      ]);
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">Interactive Administration Terminal</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Execute live vpsctl administrative operations against the host with real-time audit logging and validation.
          </p>
        </div>
        <div className="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
          <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
          <span>Live vpsctl Shell</span>
        </div>
      </div>

      <div className="bg-slate-950 text-slate-100 rounded-lg border border-slate-800 font-mono text-xs flex flex-col h-[560px] overflow-hidden shadow-sm">
        {/* Terminal Header */}
        <div className="px-4 py-2.5 bg-slate-900 border-b border-slate-800 flex items-center justify-between text-slate-400">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500"></span>
            <span className="w-2.5 h-2.5 rounded-full bg-amber-500"></span>
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
            <span className="text-[11px] ml-2 text-slate-300 font-medium">live session @ host</span>
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
          {isRunning && (
            <div className="flex items-center gap-2 text-slate-400 pl-4 py-1">
              <Loader2 className="w-3.5 h-3.5 animate-spin text-blue-400" />
              <span className="text-[11px] text-slate-400">Executing command on host...</span>
            </div>
          )}
        </div>

        {/* Terminal Input Form */}
        <form onSubmit={handleRun} className="p-3 bg-slate-900 border-t border-slate-800 flex items-center gap-2">
          <span className="text-blue-400 font-bold pl-2">$</span>
          <input
            type="text"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            disabled={isRunning}
            placeholder="e.g. vpsctl server resources, vpsctl server status, help..."
            className="flex-1 bg-transparent text-slate-100 focus:outline-none font-mono text-xs disabled:opacity-50"
            autoFocus
          />
          <button
            type="submit"
            disabled={isRunning || !command.trim()}
            className="p-1.5 text-slate-400 hover:text-white rounded hover:bg-slate-800 disabled:opacity-40"
          >
            <Send className="w-3.5 h-3.5" />
          </button>
        </form>
      </div>
    </div>
  );
};
