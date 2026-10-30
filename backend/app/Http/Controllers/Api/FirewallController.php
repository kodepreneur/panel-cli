<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Services\VpsDaemonService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class FirewallController extends Controller
{
    protected VpsDaemonService $daemon;

    public function __construct(VpsDaemonService $daemon)
    {
        $this->daemon = $daemon;
    }

    public function status(): JsonResponse
    {
        $status = $this->daemon->getFirewallStatus();

        return response()->json([
            'status' => 'success',
            'data' => $status,
        ]);
    }

    public function addRule(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'port' => 'required|string',
            'proto' => 'nullable|string|in:tcp,udp,any',
            'action' => 'nullable|string|in:allow,deny',
        ]);

        $success = $this->daemon->addFirewallRule(
            $validated['port'],
            $validated['proto'] ?? 'tcp',
            $validated['action'] ?? 'allow'
        );

        return response()->json([
            'status' => 'success',
            'message' => 'Firewall rule added successfully.',
            'data' => [
                'rule' => $validated,
                'applied' => $success,
            ],
        ]);
    }

    public function deleteRule(int $ruleNumber): JsonResponse
    {
        $success = $this->daemon->deleteFirewallRule($ruleNumber);

        return response()->json([
            'status' => 'success',
            'message' => "Firewall rule #{$ruleNumber} deleted.",
            'data' => [
                'rule_number' => $ruleNumber,
                'deleted' => $success,
            ],
        ]);
    }
}

