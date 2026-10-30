<?php

use App\Http\Controllers\Api\AgentController;
use App\Http\Controllers\Api\DatabaseController;
use App\Http\Controllers\Api\FirewallController;
use App\Http\Controllers\Api\GithubController;
use App\Http\Controllers\Api\ProjectController;
use App\Http\Controllers\Api\ServerController;
use App\Http\Controllers\Api\SslController;
use Illuminate\Support\Facades\Route;

Route::prefix('v1')->group(function () {
    // Health check
    Route::get('/health', function () {
        return response()->json([
            'status' => 'healthy',
            'panel' => 'AI-Native VPS Management Panel',
            'version' => '1.0.0',
            'timestamp' => now()->toIso8601String(),
        ]);
    });

    // Webhook from GitHub (public, HMAC-SHA256 verified)
    Route::post('/webhooks/github/{id}', [GithubController::class, 'webhook']);

    // Authenticated API routes (supports Sanctum or Agent API Key vps_live_...)
    Route::middleware(['agent.auth'])->group(function () {
        // Server Management
        Route::get('/server/status', [ServerController::class, 'status'])->middleware('agent.auth:server.read');
        Route::get('/server/resources', [ServerController::class, 'resources'])->middleware('agent.auth:server.read');
        Route::get('/server/services', [ServerController::class, 'services'])->middleware('agent.auth:server.read');
        Route::post('/server/services/{name}/restart', [ServerController::class, 'restartService'])->middleware('agent.auth:service.manage');
        Route::post('/server/reboot', [ServerController::class, 'reboot'])->middleware('agent.auth:server.reboot');
        Route::post('/server/terminal', [ServerController::class, 'terminal'])->middleware('agent.auth:server.read');

        // Firewall Management
        Route::get('/server/firewall', [FirewallController::class, 'status'])->middleware('agent.auth:server.read');
        Route::post('/server/firewall/rules', [FirewallController::class, 'addRule'])->middleware('agent.auth:firewall.modify');
        Route::delete('/server/firewall/rules/{ruleNumber}', [FirewallController::class, 'deleteRule'])->middleware('agent.auth:firewall.modify');

        // Project Lifecycle
        Route::get('/projects', [ProjectController::class, 'index'])->middleware('agent.auth:project.read');
        Route::post('/projects', [ProjectController::class, 'store'])->middleware('agent.auth:project.create');
        Route::get('/projects/{id}', [ProjectController::class, 'show'])->middleware('agent.auth:project.read');
        Route::delete('/projects/{id}', [ProjectController::class, 'destroy'])->middleware('agent.auth:project.delete');

        // Deployments & Releases
        Route::post('/projects/{id}/deploy', [ProjectController::class, 'deploy'])->middleware('agent.auth:project.deploy');
        Route::post('/projects/{id}/rollback', [ProjectController::class, 'rollback'])->middleware('agent.auth:project.deploy');
        Route::get('/projects/{id}/logs', [ProjectController::class, 'logs'])->middleware('agent.auth:project.read');

        // SSL Management
        Route::post('/projects/{id}/ssl/install', [SslController::class, 'install'])->middleware('agent.auth:ssl.manage');
        Route::get('/projects/{id}/ssl/status', [SslController::class, 'status'])->middleware('agent.auth:ssl.manage');

        // GitHub Integration
        Route::get('/github/repositories', [GithubController::class, 'repositories'])->middleware('agent.auth:github.manage');
        Route::post('/projects/{id}/github/pull', [GithubController::class, 'pull'])->middleware('agent.auth:github.manage');

        // Database Management
        Route::get('/databases', [DatabaseController::class, 'index'])->middleware('agent.auth:server.read');
        Route::post('/databases', [DatabaseController::class, 'store'])->middleware('agent.auth:server.admin');

        // AI Agents & Approvals Management
        Route::get('/agents', [AgentController::class, 'index'])->middleware('agent.auth:server.admin');
        Route::post('/agents', [AgentController::class, 'store'])->middleware('agent.auth:server.admin');
        Route::delete('/agents/{id}', [AgentController::class, 'destroy'])->middleware('agent.auth:server.admin');
        Route::get('/agents/approvals/pending', [AgentController::class, 'pendingApprovals'])->middleware('agent.auth:server.admin');
        Route::post('/agents/approvals/{id}/approve', [AgentController::class, 'approve'])->middleware('agent.auth:server.admin');
        Route::post('/agents/approvals/{id}/reject', [AgentController::class, 'reject'])->middleware('agent.auth:server.admin');
        Route::get('/agents/audit-logs', [AgentController::class, 'auditLogs'])->middleware('agent.auth:server.admin');
    });
});
