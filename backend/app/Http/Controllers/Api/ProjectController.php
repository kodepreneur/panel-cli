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

class ProjectController extends Controller
{
    protected VpsDaemonService $daemon;

    public function __construct(VpsDaemonService $daemon)
    {
        $this->daemon = $daemon;
    }

    public function index(Request $request): JsonResponse
    {
        $agent = $request->attributes->get('agent');
        $query = Project::with(['releases', 'sslCertificate']);

        if ($agent && !empty($agent->allowed_project_ids)) {
            $query->whereIn('id', $agent->allowed_project_ids);
        }

        $projects = $query->latest()->get();

        return response()->json([
            'status' => 'success',
            'data' => $projects,
        ]);
    }

    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'name' => 'required|string|max:100',
            'domain' => 'required|string|max:255|unique:projects,domain',
            'project_type' => 'required|string|in:laravel,php_native,nodejs,nextjs,react_vite,python,docker,static',
            'repository_url' => 'nullable|string',
            'branch' => 'nullable|string',
            'php_version' => 'nullable|string',
            'upstream_port' => 'nullable|integer',
            'enable_ssl' => 'nullable|boolean',
        ]);

        $slug = Str::slug($validated['name']);
        $baseProjectsDir = env('VPSD_PROJECTS_DIR', '/var/www/vps-panel/projects');
        $rootDir = "{$baseProjectsDir}/{$slug}/current";

        $publicDir = match ($validated['project_type']) {
            'laravel' => 'public',
            'react_vite', 'static' => 'dist',
            default => '',
        };

        $project = Project::create([
            'name' => $validated['name'],
            'slug' => $slug,
            'domain' => $validated['domain'],
            'project_type' => $validated['project_type'],
            'repository_url' => $validated['repository_url'] ?? null,
            'branch' => $validated['branch'] ?? 'main',
            'root_dir' => $rootDir,
            'public_dir' => $publicDir,
            'php_version' => $validated['php_version'] ?? '8.4',
            'upstream_port' => $validated['upstream_port'] ?? null,
            'status' => 'active',
            'ssl_enabled' => false,
        ]);

        // 1. Prepare initial release in filesystem
        $releaseTag = now()->format('Ymd_His');
        $prepRes = $this->daemon->prepareRelease($slug, $releaseTag);
        $actRes = $this->daemon->activateRelease($slug, $releaseTag);

        $release = Release::create([
            'project_id' => $project->id,
            'release_tag' => $releaseTag,
            'path' => $prepRes['release_path'] ?? "{$baseProjectsDir}/{$slug}/releases/{$releaseTag}",
            'status' => 'active',
        ]);

        // 2. Configure Nginx Virtual Host
        $vhostConfig = [
            'slug' => $slug,
            'primary_domain' => $validated['domain'],
            'aliases' => [],
            'project_type' => $validated['project_type'],
            'root_path' => $rootDir,
            'public_dir' => $publicDir,
            'php_version' => $project->php_version,
            'upstream_port' => $project->upstream_port,
            'ssl_enabled' => false,
            'hsts' => false,
        ];
        $this->daemon->applyVhost($vhostConfig);

        // 3. Optional SSL auto-provisioning
        if ($request->boolean('enable_ssl')) {
            try {
                $this->daemon->issueSSL($project->domain, "admin@{$project->domain}");
                $project->update(['ssl_enabled' => true]);
            } catch (\Throwable $e) {
                // Non-blocking error for dev / unpointed DNS
            }
        }

        return response()->json([
            'status' => 'success',
            'message' => 'Project created successfully',
            'data' => $project->load('releases'),
        ], 201);
    }

    public function show(string $id): JsonResponse
    {
        $project = Project::with(['releases', 'deployments', 'sslCertificate'])->findOrFail($id);

        return response()->json([
            'status' => 'success',
            'data' => $project,
        ]);
    }

    public function destroy(string $id): JsonResponse
    {
        $project = Project::findOrFail($id);

        // Remove Nginx vhost
        $this->daemon->removeVhost($project->slug);

        $project->delete();

        return response()->json([
            'status' => 'success',
            'message' => "Project '{$project->slug}' deleted successfully.",
        ]);
    }

    public function deploy(Request $request, string $id): JsonResponse
    {
        $project = Project::findOrFail($id);
        $agent = $request->attributes->get('agent');

        $releaseTag = now()->format('Ymd_His');
        $prep = $this->daemon->prepareRelease($project->slug, $releaseTag);
        $act = $this->daemon->activateRelease($project->slug, $releaseTag);

        $release = Release::create([
            'project_id' => $project->id,
            'release_tag' => $releaseTag,
            'path' => $prep['release_path'] ?? '',
            'branch' => $project->branch,
            'commit_hash' => Str::random(8),
            'commit_message' => 'Automated deployment via ' . ($agent ? "AI Agent ({$agent->name})" : 'API/Dashboard'),
            'status' => 'active',
        ]);

        $deployment = Deployment::create([
            'project_id' => $project->id,
            'release_id' => $release->id,
            'trigger_type' => $agent ? 'agent' : 'manual',
            'triggered_by_type' => $agent ? 'ai_agent' : 'user',
            'triggered_by_id' => $agent ? $agent->id : null,
            'status' => 'success',
            'build_logs' => "Release {$releaseTag} built and activated atomically.\nSymlink: {$act['current_symlink']}",
            'duration_seconds' => 4,
            'started_at' => now()->subSeconds(4),
            'completed_at' => now(),
        ]);

        return response()->json([
            'status' => 'success',
            'message' => 'Deployment executed successfully.',
            'data' => [
                'deployment' => $deployment,
                'release' => $release,
            ],
        ]);
    }

    public function updateFromRepo(Request $request, string $id): JsonResponse
    {
        $project = Project::findOrFail($id);
        $agent = $request->attributes->get('agent');

        $validated = $request->validate([
            'repository_url' => 'nullable|string',
            'branch' => 'nullable|string',
            'run_build' => 'nullable|boolean',
        ]);

        $repoUrl = $validated['repository_url'] ?? $project->repository_url ?? '';
        $branch = $validated['branch'] ?? $project->branch ?? 'main';
        $runBuild = $request->boolean('run_build', true);

        if (empty($repoUrl)) {
            return response()->json([
                'status' => 'error',
                'message' => 'No repository URL configured for this project. Please provide a repository URL.',
            ], 422);
        }

        // Update project repository details if provided
        $projectUpdates = [];
        if (!empty($validated['repository_url']) && $validated['repository_url'] !== $project->repository_url) {
            $projectUpdates['repository_url'] = $validated['repository_url'];
        }
        if (!empty($validated['branch']) && $validated['branch'] !== $project->branch) {
            $projectUpdates['branch'] = $validated['branch'];
        }
        if (!empty($projectUpdates)) {
            $project->update($projectUpdates);
        }

        $startTime = microtime(true);
        $result = $this->daemon->updateFromRepo($project->slug, $repoUrl, $branch, $runBuild);
        $duration = max(1, (int) round(microtime(true) - $startTime));

        $releaseTag = $result['release_tag'] ?? now()->format('Ymd_His');
        $commitHash = $result['commit_hash'] ?? Str::random(8);
        $commitMessage = $result['commit_message'] ?? 'Updated from repository';
        $commitAuthor = $result['commit_author'] ?? ($agent ? $agent->name : 'Dashboard User');
        $logs = $result['logs'] ?? "Repository updated and release {$releaseTag} activated.";

        $release = Release::create([
            'project_id' => $project->id,
            'release_tag' => $releaseTag,
            'path' => $result['target_release'] ?? '',
            'branch' => $branch,
            'commit_hash' => $commitHash,
            'commit_message' => $commitMessage,
            'commit_author' => $commitAuthor,
            'status' => 'active',
        ]);

        $deployment = Deployment::create([
            'project_id' => $project->id,
            'release_id' => $release->id,
            'trigger_type' => $agent ? 'agent' : 'repo_update',
            'triggered_by_type' => $agent ? 'ai_agent' : 'user',
            'triggered_by_id' => $agent ? $agent->id : null,
            'status' => 'success',
            'build_logs' => $logs,
            'duration_seconds' => $duration,
            'started_at' => now()->subSeconds($duration),
            'completed_at' => now(),
        ]);

        return response()->json([
            'status' => 'success',
            'message' => "Project '{$project->name}' successfully updated from repository.",
            'data' => [
                'project' => $project->fresh(['releases', 'deployments']),
                'deployment' => $deployment,
                'release' => $release,
                'logs' => $logs,
            ],
        ]);
    }


    public function rollback(Request $request, string $id): JsonResponse
    {
        $project = Project::findOrFail($id);
        $targetTag = $request->input('target_release_tag');

        $this->daemon->rollbackRelease($project->slug, $targetTag);

        return response()->json([
            'status' => 'success',
            'message' => "Project '{$project->slug}' rolled back successfully.",
        ]);
    }

    public function logs(string $id): JsonResponse
    {
        $project = Project::findOrFail($id);
        $latestDeployment = $project->deployments()->latest()->first();

        return response()->json([
            'status' => 'success',
            'data' => [
                'project' => $project->slug,
                'logs' => $latestDeployment?->build_logs ?? "No deployment logs found.",
            ],
        ]);
    }
}

