<?php

namespace App\Http\Middleware;

use App\Models\AiAgent;
use App\Models\AgentAuditLog;
use App\Models\ApprovalRequest;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class AuthenticateAgentApiKey
{
    /**
     * Handle an incoming request from an AI agent or human token
     */
    public function handle(Request $request, Closure $next, ?string $requiredPermission = null): Response
    {
        $startTime = microtime(true);
        $authHeader = $request->header('Authorization');

        if (!$authHeader || !str_starts_with($authHeader, 'Bearer ')) {
            return response()->json([
                'status' => 'error',
                'message' => 'Missing or malformed Authorization header.',
            ], 401);
        }

        $token = substr($authHeader, 7);

        // If it's an AI Agent key (vps_live_...)
        if (str_starts_with($token, 'vps_live_')) {
            $keyHash = hash('sha256', $token);
            $agent = AiAgent::where('api_key_hash', $keyHash)
                ->where('is_active', true)
                ->first();

            if (!$agent) {
                return response()->json([
                    'status' => 'error',
                    'message' => 'Invalid or inactive agent API key.',
                ], 401);
            }

            if ($agent->expires_at && $agent->expires_at->isPast()) {
                return response()->json([
                    'status' => 'error',
                    'message' => 'Agent credentials have expired.',
                ], 401);
            }

            // IP allowlist check
            if (!empty($agent->ip_allowlist)) {
                $clientIp = $request->ip();
                if (!in_array($clientIp, $agent->ip_allowlist)) {
                    return response()->json([
                        'status' => 'error',
                        'message' => "Client IP {$clientIp} is not in the agent's allowlist.",
                    ], 403);
                }
            }

            // Permission check
            if ($requiredPermission && !$agent->hasPermission($requiredPermission)) {
                return response()->json([
                    'status' => 'error',
                    'message' => "Agent does not have the required permission: {$requiredPermission}",
                ], 403);
            }

            // Sensitive action approval gate
            $sensitiveActions = [
                'project.delete',
                'server.reboot',
                'firewall.modify',
                'terminal.execute',
            ];

            if ($requiredPermission && in_array($requiredPermission, $sensitiveActions)) {
                $approvalId = $request->header('X-Approval-ID');
                $isPreApproved = false;

                if ($approvalId) {
                    $approval = ApprovalRequest::where('id', $approvalId)
                        ->where('agent_id', $agent->id)
                        ->where('status', 'approved')
                        ->first();
                    if ($approval) {
                        $isPreApproved = true;
                    }
                }

                // If agent has full root '*' permission and confirms admin intent
                if ($agent->hasPermission('*') && ($request->header('X-Panel-Admin') === 'true' || $request->header('X-Admin-Confirm') === 'true')) {
                    $isPreApproved = true;
                }

                if (!$isPreApproved) {
                    // Create pending approval request
                    $approvalReq = ApprovalRequest::create([
                        'agent_id' => $agent->id,
                        'action_type' => $requiredPermission,
                        'resource_type' => $request->route('id') ? 'project' : 'system',
                        'resource_id' => $request->route('id'),
                        'payload' => $request->all(),
                        'status' => 'pending',
                        'expires_at' => now()->addHours(24),
                    ]);

                    return response()->json([
                        'status' => 'pending_approval',
                        'approval_id' => $approvalReq->id,
                        'message' => "Action '{$requiredPermission}' requires human administrator approval. Approval request submitted.",
                        'expires_at' => $approvalReq->expires_at->toIso8601String(),
                    ], 202);
                }
            }

            // Track last used timestamp
            $agent->update(['last_used_at' => now()]);
            $request->attributes->set('agent', $agent);

            /** @var Response $response */
            $response = $next($request);

            // Audit log
            $execTimeMs = (int) round((microtime(true) - $startTime) * 1000);
            AgentAuditLog::create([
                'agent_id' => $agent->id,
                'endpoint_or_tool' => $request->path(),
                'method' => $request->method(),
                'ip_address' => $request->ip(),
                'request_params' => $request->except(['password', 'api_key', 'token']),
                'response_code' => $response->getStatusCode(),
                'execution_time_ms' => $execTimeMs,
            ]);

            return $response;
        }

        // Otherwise delegate to standard Sanctum token or proceed
        return $next($request);
    }
}

