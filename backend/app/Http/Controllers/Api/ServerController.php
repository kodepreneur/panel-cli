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
                'uptime_seconds' => $overview['uptime_seconds'] ?? 0,
                'os' => $overview['os'] ?? [
                    'name' => PHP_OS_FAMILY,
                    'version' => PHP_OS_FAMILY,
                    'kernel' => php_uname('r'),
                    'arch' => php_uname('m'),
                    'hostname' => gethostname(),
                ],
                'cpu' => $overview['cpu'] ?? [],
                'memory' => $overview['memory'] ?? [],
                'disks' => $overview['disks'] ?? [],
                'network' => $overview['network'] ?? [],
                'load_avg' => $overview['load_avg'] ?? [0, 0, 0],
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

    public function terminal(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'command' => 'required|string|max:500',
        ]);

        $cmdStr = trim($validated['command']);

        // Check command prefix
        $allowedPrefixes = ['vpsctl', 'help', 'uptime', 'df', 'whoami', 'uname', 'top', 'ps'];
        $firstWord = explode(' ', $cmdStr)[0];
        if (!in_array($firstWord, $allowedPrefixes)) {
            return response()->json([
                'status' => 'error',
                'message' => "Command not permitted in sandboxed terminal: {$firstWord}",
                'output' => "Command '{$firstWord}' is not approved for execution.\nApproved commands: vpsctl, help, uptime, df, whoami, uname, ps",
                'exit_code' => 127,
            ], 403);
        }

        $vpsctlBin = base_path('../bin/vpsctl');
        if (!file_exists($vpsctlBin)) {
            $vpsctlBin = '/usr/local/bin/vpsctl';
        }

        $vpsdBin = base_path('../bin/vpsd');
        if (!file_exists($vpsdBin)) {
            $vpsdBin = '/usr/local/bin/vpsd';
        }

        // Ensure vpsd socket is active for vpsctl commands
        $sock = env('VPSD_SOCKET_PATH', file_exists('/tmp/vpsd.sock') ? '/tmp/vpsd.sock' : '/run/vpsd/vpsd.sock');
        if (!file_exists($sock) && file_exists($vpsdBin)) {
            $sock = '/tmp/vpsd.sock';
            shell_exec(escapeshellarg($vpsdBin) . " --socket=" . escapeshellarg($sock) . " > /dev/null 2>&1 &");
            usleep(100000); // 100ms
        }

        $startTime = microtime(true);

        if ($firstWord === 'help') {
            $output = shell_exec(escapeshellarg($vpsctlBin) . ' help 2>&1') ?? "vpsctl - AI-Native VPS Management Panel CLI\nType 'vpsctl help' for all commands.";
            $exitCode = 0;
        } elseif ($firstWord === 'vpsctl') {
            $args = substr($cmdStr, 6); // remove 'vpsctl'
            $fullCmd = escapeshellarg($vpsctlBin) . ' ' . $args . " --socket=" . escapeshellarg($sock) . " 2>&1";
            $output = shell_exec($fullCmd) ?? '';
            $exitCode = 0;
        } else {
            $output = shell_exec(escapeshellcmd($cmdStr) . ' 2>&1') ?? '';
            $exitCode = 0;
        }

        $durationMs = (int) round((microtime(true) - $startTime) * 1000);

        return response()->json([
            'status' => 'success',
            'command' => $cmdStr,
            'output' => trim($output),
            'duration_ms' => $durationMs,
            'timestamp' => now()->toIso8601String(),
        ]);
    }

    public function updateStatus(): JsonResponse
    {
        $status = $this->daemon->getSystemUpdateStatus();

        return response()->json([
            'status' => 'success',
            'data' => $status,
        ]);
    }

    public function updateSystem(Request $request): JsonResponse
    {
        $branch = $request->input('branch');
        $result = $this->daemon->updateSystem($branch);

        return response()->json([
            'status' => 'success',
            'message' => 'VPS Panel update completed.',
            'data' => $result,
        ]);
    }
}


