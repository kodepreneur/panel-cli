<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Project;
use App\Models\SslCertificate;
use App\Services\VpsDaemonService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class SslController extends Controller
{
    protected VpsDaemonService $daemon;

    public function __construct(VpsDaemonService $daemon)
    {
        $this->daemon = $daemon;
    }

    public function install(Request $request, string $id): JsonResponse
    {
        $project = Project::findOrFail($id);
        $email = $request->input('email', "admin@{$project->domain}");

        $res = $this->daemon->issueSSL($project->domain, $email, $project->aliases ?? []);

        if (!($res['success'] ?? false)) {
            return response()->json([
                'status' => 'error',
                'message' => 'SSL issuance failed.',
                'logs' => $res['logs'] ?? '',
                'error' => $res['error'] ?? 'Unknown error',
            ], 422);
        }

        $cert = SslCertificate::updateOrCreate(
            ['project_id' => $project->id],
            [
                'primary_domain' => $project->domain,
                'san_domains' => $project->aliases ?? [],
                'issuer' => $res['issuer'] ?? "Let's Encrypt",
                'cert_path' => $res['cert_path'] ?? '',
                'key_path' => $res['key_path'] ?? '',
                'expires_at' => $res['expires_at'] ?? now()->addDays(90),
                'last_renewed_at' => now(),
                'status' => 'valid',
            ]
        );

        $project->update(['ssl_enabled' => true]);

        // Reapply virtual host with SSL paths enabled
        $vhostConfig = [
            'slug' => $project->slug,
            'primary_domain' => $project->domain,
            'aliases' => $project->aliases ?? [],
            'project_type' => $project->project_type,
            'root_path' => $project->root_dir,
            'public_dir' => $project->public_dir,
            'php_version' => $project->php_version,
            'upstream_port' => $project->upstream_port,
            'ssl_enabled' => true,
            'cert_path' => $cert->cert_path,
            'key_path' => $cert->key_path,
            'hsts' => true,
        ];
        $this->daemon->applyVhost($vhostConfig);

        return response()->json([
            'status' => 'success',
            'message' => 'SSL certificate installed and HTTPS enabled.',
            'data' => $cert,
        ]);
    }

    public function status(string $id): JsonResponse
    {
        $project = Project::findOrFail($id);
        $status = $this->daemon->getSSLStatus($project->domain);

        return response()->json([
            'status' => 'success',
            'data' => $status,
        ]);
    }
}

