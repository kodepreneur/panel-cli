<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\AgentAuditLog;
use App\Models\AiAgent;
use App\Models\ApprovalRequest;
use App\Services\VpsDaemonService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class AgentController extends Controller
{
    protected VpsDaemonService $daemon;

    public function __construct(VpsDaemonService $daemon)
    {
        $this->daemon = $daemon;
    }

    public function index(): JsonResponse
    {
        $agents = AiAgent::withCount('approvals')
            ->latest()
            ->get();

        return response()->json([
            'status' => 'success',
            'data' => $agents,
        ]);
    }

    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'name' => 'required|string|max:100',
            'agent_type' => 'required|string|in:hermes,openclaw,openhands,custom',
            'permissions' => 'required|array',
            'allowed_project_ids' => 'nullable|array',
            'ip_allowlist' => 'nullable|array',
            'expires_at' => 'nullable|date',
        ]);

        $rawKey = 'vps_live_' . bin2hex(random_bytes(24));
        $keyHash = hash('sha256', $rawKey);
        $keyPrefix = substr($rawKey, 0, 15) . '...';

        $agent = AiAgent::create([
            'name' => $validated['name'],
            'agent_type' => $validated['agent_type'],
            'api_key_hash' => $keyHash,
            'api_key_prefix' => $keyPrefix,
            'permissions' => $validated['permissions'],
            'allowed_project_ids' => $validated['allowed_project_ids'] ?? null,
            'ip_allowlist' => $validated['ip_allowlist'] ?? null,
            'is_active' => true,
            'expires_at' => $validated['expires_at'] ?? null,
        ]);

        return response()->json([
            'status' => 'success',
            'message' => 'AI Agent provisioned successfully.',
            'data' => [
                'agent' => $agent,
                'api_key' => $rawKey, // Plaintext shown only once upon creation
            ],
        ], 201);
    }

    public function destroy(string $id): JsonResponse
    {
        $agent = AiAgent::findOrFail($id);
        $agent->delete();

        return response()->json([
            'status' => 'success',
            'message' => "AI Agent '{$agent->name}' credentials revoked successfully.",
        ]);
    }

    public function pendingApprovals(): JsonResponse
    {
        $approvals = ApprovalRequest::with('agent')
            ->where('status', 'pending')
            ->latest()
            ->get();

        return response()->json([
            'status' => 'success',
            'data' => $approvals,
        ]);
    }

    public function approve(Request $request, string $id): JsonResponse
    {
        $approval = ApprovalRequest::with('agent')->findOrFail($id);
        if ($approval->status !== 'pending') {
            return response()->json([
                'status' => 'error',
                'message' => "Approval request is already {$approval->status}.",
            ], 422);
        }

        $approval->update([
            'status' => 'approved',
            'approved_at' => now(),
        ]);

        // Execute the approved sensitive action
        $actionResult = match ($approval->action_type) {
            'project.delete' => $approval->resource_id ? $this->daemon->removeVhost($approval->resource_id) : true,
            'server.reboot' => true,
            default => true,
        };

        return response()->json([
            'status' => 'success',
            'message' => "Action '{$approval->action_type}' approved and executed.",
            'data' => [
                'approval' => $approval,
                'action_result' => $actionResult,
            ],
        ]);
    }

    public function reject(Request $request, string $id): JsonResponse
    {
        $approval = ApprovalRequest::findOrFail($id);
        $approval->update([
            'status' => 'rejected',
            'rejection_reason' => $request->input('reason', 'Rejected by administrator'),
        ]);

        return response()->json([
            'status' => 'success',
            'message' => 'Action rejected.',
            'data' => $approval,
        ]);
    }

    public function auditLogs(Request $request): JsonResponse
    {
        $logs = AgentAuditLog::with('agent')
            ->latest()
            ->paginate($request->integer('per_page', 50));

        return response()->json([
            'status' => 'success',
            'data' => $logs,
        ]);
    }
}

