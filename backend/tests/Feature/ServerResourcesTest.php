<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class ServerResourcesTest extends TestCase
{
    use RefreshDatabase;
    public function test_server_resources_returns_full_overview_contract(): void
    {
        $response = $this->withHeader('Authorization', 'Bearer admin-test-token')
            ->getJson('/api/v1/server/resources');

        $response->assertStatus(200)
            ->assertJsonStructure([
                'status',
                'data' => [
                    'uptime_seconds',
                    'os' => [
                        'name',
                        'version',
                        'kernel',
                        'arch',
                        'hostname',
                    ],
                    'cpu' => [
                        'model',
                        'cores',
                        'usage_percent',
                    ],
                    'memory' => [
                        'total_bytes',
                        'used_bytes',
                        'free_bytes',
                        'usage_percent',
                    ],
                    'disks',
                    'network',
                    'load_avg',
                    'load',
                    'timestamp',
                ],
            ]);

        $data = $response->json('data');
        $this->assertIsInt($data['uptime_seconds']);
        $this->assertNotEmpty($data['os']['hostname']);
        $this->assertNotEmpty($data['os']['arch']);
    }

    public function test_server_status_returns_success(): void
    {
        $response = $this->withHeader('Authorization', 'Bearer admin-test-token')
            ->getJson('/api/v1/server/status');

        $response->assertStatus(200)
            ->assertJson([
                'status' => 'success',
            ])
            ->assertJsonStructure([
                'status',
                'data' => [
                    'panel_version',
                    'hostname',
                    'os',
                    'uptime_seconds',
                    'load_avg',
                ],
            ]);
    }
}
