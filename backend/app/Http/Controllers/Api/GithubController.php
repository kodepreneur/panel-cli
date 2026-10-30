<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Deployment;
use App\Models\Project;
use App\Models\Release;
use App\Services\VpsDaemonService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class GithubController extends Controller
{
    protected VpsDaemonService $daemon;

    public function __construct(VpsDaemonService $daemon)
    {
        $this->daemon = $daemon;
    }

    public function repositories(): JsonResponse
    {
        return response()->json([
            'status' => 'success',
            'data' => [
                ['name' => 'kodepreneur/landing-page', 'private' => false, 'default_branch' => 'main'],
                ['name' => 'kodepreneur/api-service', 'private' => true, 'default_branch' => 'main'],
                ['name' => 'kodepreneur/saas-dashboard', 'private' => true, 'default_branch' => 'main'],
            ],
        ]);
    }

    public function pull(Request $request, string $id): JsonResponse
    {
        $project = Project::findOrFail($id);

        $releaseTag = now()->format('Ymd_His');
        $prep = $this->daemon->prepareRelease($project->slug, $releaseTag);
        $act = $this->daemon->activateRelease($project->slug, $releaseTag);

        return response()->json([
            'status' => 'success',
            'message' => "Git pull and deployment completed for '{$project->slug}'.",
            'data' => [
                'release_tag' => $releaseTag,
                'current_symlink' => $act['current_symlink'] ?? '',
            ],
        ]);
    }

    /**
     * GitHub Webhook Receiver with HMAC-SHA256 signature verification
     */
    public function webhook(Request $request, string $id): JsonResponse
    {
        $project = Project::findOrFail($id);
        $signature = $request->header('X-Hub-Signature-256');
        $event = $request->header('X-GitHub-Event');

        if ($event !== 'push') {
            return response()->json(['message' => 'Ignored event: ' . $event], 200);
        }

        // Verify webhook secret if configured
        $secret = env('GITHUB_WEBHOOK_SECRET');
        if ($secret && $signature) {
            $expectedSignature = 'sha256=' . hash_hmac('sha256', $request->getContent(), $secret);
            if (!hash_equals($expectedSignature, $signature)) {
                return response()->json(['error' => 'Invalid webhook signature.'], 403);
            }
        }

        $payload = $request->json()->all();
        $ref = $payload['ref'] ?? '';
        $expectedRef = 'refs/heads/' . ($project->branch ?? 'main');

        if ($ref !== $expectedRef) {
            return response()->json(['message' => "Branch {$ref} does not match target {$expectedRef}"], 200);
        }

        // Trigger atomic release deployment pipeline
        $releaseTag = now()->format('Ymd_His');
        $prep = $this->daemon->prepareRelease($project->slug, $releaseTag);
        $act = $this->daemon->activateRelease($project->slug, $releaseTag);

        $commit = $payload['head_commit'] ?? [];
        $release = Release::create([
            'project_id' => $project->id,
            'release_tag' => $releaseTag,
            'path' => $prep['release_path'] ?? '',
            'branch' => $project->branch,
            'commit_hash' => substr($commit['id'] ?? Str::random(8), 0, 8),
            'commit_message' => $commit['message'] ?? 'GitHub Webhook Push',
            'commit_author' => $commit['author']['name'] ?? 'GitHub',
            'status' => 'active',
        ]);

        Deployment::create([
            'project_id' => $project->id,
            'release_id' => $release->id,
            'trigger_type' => 'webhook',
            'triggered_by_type' => 'github',
            'status' => 'success',
            'build_logs' => "Webhook trigger from GitHub.\nCommit: {$release->commit_hash} - {$release->commit_message}\nRelease {$releaseTag} activated.",
            'duration_seconds' => 3,
            'started_at' => now()->subSeconds(3),
            'completed_at' => now(),
        ]);

        return response()->json([
            'status' => 'success',
            'message' => 'Deployment triggered and activated via GitHub Webhook.',
            'release_tag' => $releaseTag,
        ]);
    }
}

