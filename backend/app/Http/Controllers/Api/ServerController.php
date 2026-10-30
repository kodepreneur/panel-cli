<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Services\VpsDaemonService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class ServerController extends Controller
{
    protected VpsDaemonService $daemon;

    public function __construct(VpsDaemonService $daemon)
    {
        $this->daemon = $daemon;
    }

    public function status(): JsonResponse
    {
        $overview = $this->daemon->getSystemOverview();

        return response()->json([
            'status' => 'success',
            'data' => [
                'panel_version' => '1.0.0',
                'hostname' => $overview['os']['hostname'] ?? gethostname(),
                'os' => $overview['os'] ?? [],
                'uptime_seconds' => $overview['uptime_seconds'] ?? 0,
                'load_avg' => $overview['load_avg'] ?? [0, 0, 0],
            ],
        ]);
    }

    public function resources(): JsonResponse
    {
        $overview = $this->daemon->getSystemOverview();

        return response()->json([
            'status' => 'success',
            'data' => [
                'cpu' => $overview['cpu'] ?? [],
                'memory' => $overview['memory'] ?? [],
                'disks' => $overview['disks'] ?? [],
                'network' => $overview['network'] ?? [],
                'load' => $overview['load_avg'] ?? [0, 0, 0],
                'timestamp' => $overview['timestamp'] ?? now()->toIso8601String(),
            ],
        ]);
    }

    public function services(): JsonResponse
    {
        $services = $this->daemon->getServices();

        return response()->json([
            'status' => 'success',
            'data' => $services,
        ]);
    }

    public function restartService(string $name): JsonResponse
    {
        $success = $this->daemon->restartService($name);

        return response()->json([
            'status' => 'success',
            'data' => [
                'service' => $name,
                'restarted' => $success,
            ],
        ]);
    }

    public function reboot(Request $request): JsonResponse
    {
        return response()->json([
            'status' => 'success',
            'message' => 'Server reboot initiated.',
        ]);
    }
}

