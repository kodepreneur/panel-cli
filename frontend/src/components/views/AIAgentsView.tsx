import React, { useState } from 'react';
import {
  Bot,
  Plus,
  Trash2,
  Key,
  ShieldAlert,
  CheckCircle2,
  XCircle,
  Copy,
  Check,
  Terminal,
  Code2,
  X,
} from 'lucide-react';
import { AIAgentItem, ApprovalRequestItem } from '../../types';

interface AIAgentsViewProps {
  agents: AIAgentItem[];
  pendingApprovals: ApprovalRequestItem[];
  onCreateAgent: (data: {
    name: string;
    agent_type: string;
    permissions: string[];
  }) => Promise<{ agent: AIAgentItem; api_key: string }>;
  onRevokeAgent: (id: string) => Promise<void>;
  onApproveAction: (approvalId: string) => Promise<void>;
  onRejectAction: (approvalId: string) => Promise<void>;
}

export const AIAgentsView: React.FC<AIAgentsViewProps> = ({
  agents,
  pendingApprovals,
  onCreateAgent,
  onRevokeAgent,
  onApproveAction,
  onRejectAction,
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newKeyGenerated, setNewKeyGenerated] = useState<string | null>(null);
  const [copiedKey, setCopiedKey] = useState(false);

  // Form states
  const [name, setName] = useState('');
  const [agentType, setAgentType] = useState('hermes');
  const [permissions, setPermissions] = useState<string[]>([
    'server.read',
    'project.read',
    'project.create',
    'project.deploy',
    'ssl.manage',
  ]);

  const availablePermissions = [
    { id: 'server.read', label: 'server.read (View host stats, metrics, and services)' },
    { id: 'project.read', label: 'project.read (View projects, releases, and logs)' },
    { id: 'project.create', label: 'project.create (Create new projects and virtual hosts)' },
    { id: 'project.deploy', label: 'project.deploy (Trigger atomic deployments)' },
    { id: 'project.delete', label: 'project.delete (Delete projects - requires approval)' },
    { id: 'ssl.manage', label: 'ssl.manage (Issue & renew Let\'s Encrypt certificates)' },
    { id: 'github.manage', label: 'github.manage (Pull git repositories & webhooks)' },
    { id: 'service.manage', label: 'service.manage (Restart system services)' },
    { id: 'server.reboot', label: 'server.reboot (Reboot VPS - requires approval)' },
  ];

  const handleTogglePermission = (id: string) => {
    if (permissions.includes(id)) {
      setPermissions(permissions.filter((p) => p !== id));
    } else {
      setPermissions([...permissions, id]);
    }
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) return;
    const res = await onCreateAgent({
      name,
      agent_type: agentType,
      permissions,
    });
    setNewKeyGenerated(res.api_key);
    setName('');
  };

  const handleCopyKey = () => {
    if (!newKeyGenerated) return;
    navigator.clipboard.writeText(newKeyGenerated);
    setCopiedKey(true);
    setTimeout(() => setCopiedKey(false), 2000);
  };

  return (
    <div className="space-y-6">
      {/* View Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 tracking-tight">AI Agent Access & Integration</h1>
          <p className="text-xs text-slate-500 mt-0.5">
            Model Context Protocol (MCP) endpoints, scoped API keys, and Human-in-the-Loop approval queue.
          </p>
        </div>
        <button
          onClick={() => {
            setNewKeyGenerated(null);
            setIsModalOpen(true);
          }}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded bg-blue-600 hover:bg-blue-700 text-white shadow-sm transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>Provision New Agent</span>
        </button>
      </div>

      {/* Human-in-the-Loop Pending Approvals Banner */}
      {pendingApprovals.length > 0 && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ShieldAlert className="w-5 h-5 text-amber-600" />
              <h2 className="text-sm font-bold text-amber-900">
                Pending AI Agent Approval Requests ({pendingApprovals.length})
              </h2>
            </div>
            <span className="text-xs text-amber-700">Administrative confirmation required</span>
          </div>
          <div className="mt-4 space-y-3">
            {pendingApprovals.map((req) => (
              <div
                key={req.id}
                className="bg-white border border-amber-200/80 rounded p-4 flex items-center justify-between shadow-sm"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-xs text-slate-900">
                      {req.agent?.name || 'Autonomous Agent'}
                    </span>
                    <span className="text-[10px] bg-slate-100 text-slate-600 px-1.5 py-0.5 rounded font-mono">
                      {req.agent?.agent_type || 'agent'}
                    </span>
                  </div>
                  <div className="mt-1 text-xs text-slate-600 flex items-center gap-2">
                    <span>Requested action:</span>
                    <span className="font-mono font-bold text-rose-700 bg-rose-50 px-2 py-0.5 rounded border border-rose-200 text-[11px]">
                      {req.action_type}
                    </span>
                    {req.resource_id && (
                      <span className="font-mono text-slate-500 text-[11px]">
                        target: {req.resource_id}
                      </span>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => onRejectAction(req.id)}
                    className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded border border-slate-300 text-slate-700 hover:bg-slate-50"
                  >
                    <XCircle className="w-3.5 h-3.5 text-slate-400" />
                    <span>Reject</span>
                  </button>
                  <button
                    onClick={() => onApproveAction(req.id)}
                    className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded bg-emerald-600 hover:bg-emerald-700 text-white shadow-sm"
                  >
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    <span>Approve & Execute</span>
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Agents Table */}
      <div className="bg-white border border-slate-200 rounded-lg overflow-hidden">
        <div className="px-5 py-3.5 border-b border-slate-200">
          <h2 className="text-sm font-semibold text-slate-900">Configured Agent Credentials</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider font-medium border-b border-slate-200">
              <tr>
                <th className="px-5 py-3">Agent Name & Type</th>
                <th className="px-5 py-3">API Key Prefix</th>
                <th className="px-5 py-3">Assigned Permissions</th>
                <th className="px-5 py-3">Status</th>
                <th className="px-5 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {agents.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-5 py-8 text-center text-slate-400">
                    No external AI agents connected yet.
                  </td>
                </tr>
              ) : (
                agents.map((agent) => (
                  <tr key={agent.id} className="hover:bg-slate-50/70 transition-colors">
                    <td className="px-5 py-3.5">
                      <div className="font-semibold text-slate-900 flex items-center gap-2">
                        <Bot className="w-4 h-4 text-blue-600" />
                        <span>{agent.name}</span>
                      </div>
                      <div className="text-slate-400 font-mono text-[11px] mt-0.5">
                        type: {agent.agent_type}
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <div className="font-mono text-slate-700 bg-slate-50 px-2 py-0.5 rounded border border-slate-200 inline-block">
                        {agent.api_key_prefix}
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <div className="flex flex-wrap gap-1 max-w-xs">
                        {(agent.permissions || []).slice(0, 3).map((perm) => (
                          <span
                            key={perm}
                            className="text-[10px] font-mono bg-blue-50 text-blue-700 px-1.5 py-0.5 rounded border border-blue-200"
                          >
                            {perm}
                          </span>
                        ))}
                        {(agent.permissions || []).length > 3 && (
                          <span className="text-[10px] font-mono text-slate-500">
                            +{(agent.permissions || []).length - 3} more
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-5 py-3.5">
                      <span className="inline-flex items-center gap-1 text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded text-[11px] font-medium border border-emerald-200">
                        <CheckCircle2 className="w-3 h-3" /> Active
                      </span>
                    </td>
                    <td className="px-5 py-3.5 text-right">
                      <button
                        onClick={() => onRevokeAgent(agent.id)}
                        title="Revoke Credentials"
                        className="text-slate-400 hover:text-rose-600 p-1.5 rounded hover:bg-rose-50 transition-colors"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Integration Guide Box */}
      <div className="bg-slate-900 text-slate-300 rounded-lg p-5 border border-slate-800 space-y-4">
        <div className="flex items-center justify-between border-b border-slate-800 pb-3">
          <div className="flex items-center gap-2">
            <Code2 className="w-4 h-4 text-blue-400" />
            <h3 className="text-sm font-semibold text-white">Agent Connection Configuration</h3>
          </div>
          <span className="text-xs text-slate-400 font-mono">Model Context Protocol (MCP)</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs font-mono">
          <div>
            <div className="text-slate-400 mb-1 flex items-center gap-1.5">
              <Bot className="w-3.5 h-3.5 text-blue-400" /> Hermes Agent Config (hermes.json):
            </div>
            <pre className="bg-slate-950 p-3 rounded border border-slate-800 text-slate-300 overflow-x-auto">
{`{
  "mcpServers": {
    "vps_panel": {
      "transport": "sse",
      "url": "https://vps.example.com/api/v1/mcp/sse",
      "headers": {
        "Authorization": "Bearer vps_live_YOUR_KEY"
      }
    }
  }
}`}
            </pre>
          </div>

          <div>
            <div className="text-slate-400 mb-1 flex items-center gap-1.5">
              <Terminal className="w-3.5 h-3.5 text-blue-400" /> OpenClaw Agent Config (openclaw.yaml):
            </div>
            <pre className="bg-slate-950 p-3 rounded border border-slate-800 text-slate-300 overflow-x-auto">
{`agent:
  tools:
    - name: "vps_control"
      type: "mcp"
      endpoint: "https://vps.example.com/api/v1/mcp/sse"
      auth:
        type: "bearer"
        token: "vps_live_YOUR_KEY"`}
            </pre>
          </div>
        </div>
      </div>

      {/* Provision Agent Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-lg border border-slate-200 shadow-xl max-w-lg w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Key className="w-4 h-4 text-blue-600" />
                <h3 className="text-sm font-bold text-slate-900">Provision AI Agent Credentials</h3>
              </div>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-slate-400 hover:text-slate-600"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {newKeyGenerated ? (
              <div className="p-6 space-y-4 text-xs">
                <div className="p-4 bg-emerald-50 border border-emerald-200 rounded text-emerald-800">
                  <div className="font-bold flex items-center gap-1.5 text-sm">
                    <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                    Agent Credentials Generated!
                  </div>
                  <p className="mt-1">
                    Store this key safely now. For security purposes, this secret key will never be displayed again.
                  </p>
                </div>

                <div>
                  <label className="block font-medium text-slate-700 mb-1">API Key</label>
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      readOnly
                      value={newKeyGenerated}
                      className="w-full px-3 py-2 bg-slate-50 border border-slate-300 rounded font-mono text-xs select-all text-slate-900"
                    />
                    <button
                      onClick={handleCopyKey}
                      className="px-3 py-2 bg-slate-900 text-white rounded hover:bg-slate-800 flex items-center gap-1 font-medium"
                    >
                      {copiedKey ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                      <span>{copiedKey ? 'Copied' : 'Copy'}</span>
                    </button>
                  </div>
                </div>

                <div className="pt-4 border-t border-slate-100 flex justify-end">
                  <button
                    onClick={() => {
                      setIsModalOpen(false);
                      setNewKeyGenerated(null);
                    }}
                    className="px-4 py-2 bg-slate-900 text-white rounded font-medium"
                  >
                    Done
                  </button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleCreateSubmit} className="p-6 space-y-4 text-xs">
                <div>
                  <label className="block font-medium text-slate-700 mb-1">Agent Name</label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Hermes Autonomous DevOps or OpenClaw Infra"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded focus:outline-none focus:border-blue-600"
                  />
                </div>

                <div>
                  <label className="block font-medium text-slate-700 mb-1">Agent Type</label>
                  <select
                    value={agentType}
                    onChange={(e) => setAgentType(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded focus:outline-none focus:border-blue-600 bg-white"
                  >
                    <option value="hermes">Hermes Agent</option>
                    <option value="openclaw">OpenClaw Agent</option>
                    <option value="openhands">OpenHands</option>
                    <option value="custom">Custom AI Agent</option>
                  </select>
                </div>

                <div>
                  <label className="block font-medium text-slate-700 mb-2">Scoped Permissions</label>
                  <div className="space-y-1.5 max-h-48 overflow-y-auto border border-slate-200 rounded p-2.5 bg-slate-50">
                    {availablePermissions.map((perm) => (
                      <label key={perm.id} className="flex items-center gap-2 cursor-pointer hover:bg-white p-1 rounded">
                        <input
                          type="checkbox"
                          checked={permissions.includes(perm.id)}
                          onChange={() => handleTogglePermission(perm.id)}
                          className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="font-mono text-[11px] text-slate-700">{perm.label}</span>
                      </label>
                    ))}
                  </div>
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
                    className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white font-medium shadow-sm"
                  >
                    Generate Credentials
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

