<?php

namespace Tests\Feature;

use App\Models\Project;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class UpdateFromRepoTest extends TestCase
{
    use RefreshDatabase;

    public function test_can_get_system_update_status(): void
    {
        $response = $this->withHeader('Authorization', 'Bearer admin-test-token')
            ->getJson('/api/v1/server/update/status');

        $response->assertStatus(200)
            ->assertJsonStructure([
                'status',
                'data' => [
                    'current_version',
                    'current_commit',
                    'current_branch',
                    'remote_url',
                    'update_available',
                ],
            ]);
    }

    public function test_can_update_system_from_repo(): void
    {
        $response = $this->withHeader('Authorization', 'Bearer admin-test-token')
            ->postJson('/api/v1/server/update');

        $response->assertStatus(200)
            ->assertJsonStructure([
                'status',
                'message',
                'data' => [
                    'success',
                    'logs',
                ],
            ]);
    }

    public function test_can_update_project_from_repo(): void
    {
        $project = Project::firstOrCreate(
            ['slug' => 'test-repo-app'],
            [
                'name' => 'Test Repo App',
                'domain' => 'test-repo.local',
                'project_type' => 'static',
                'repository_url' => 'https://github.com/example/sample.git',
                'branch' => 'main',
                'root_dir' => '/var/www/vps-panel/projects/test-repo-app/current',
            ]
        );

        $response = $this->withHeader('Authorization', 'Bearer admin-test-token')
            ->postJson("/api/v1/projects/{$project->id}/update-from-repo", [
                'branch' => 'main',
            ]);

        $response->assertStatus(200)
            ->assertJson([
                'status' => 'success',
            ])
            ->assertJsonStructure([
                'status',
                'message',
                'data' => [
                    'project',
                    'deployment',
                    'release',
                    'logs',
                ],
            ]);

        $this->assertDatabaseHas('releases', [
            'project_id' => $project->id,
            'branch' => 'main',
        ]);

        $this->assertDatabaseHas('deployments', [
            'project_id' => $project->id,
            'status' => 'success',
        ]);
    }
}
