<?php

namespace Database\Seeders;

use App\Models\AiAgent;
use App\Models\User;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class DatabaseSeeder extends Seeder
{
    /**
     * Seed the application's database.
     */
    public function run(): void
    {
        // 1. Create Default Admin User
        $admin = User::firstOrCreate(
            ['email' => 'admin@vpspanel.local'],
            [
                'name' => 'Server Administrator',
                'password' => Hash::make('AdminPass123!'),
            ]
        );

        // 2. Create Default Test AI Agent Key
        // Key: vps_live_test_agent_secret_key_8f0a2c3d
        $testKey = 'vps_live_test_agent_secret_key_8f0a2c3d';
        $keyHash = hash('sha256', $testKey);

        AiAgent::firstOrCreate(
            ['api_key_hash' => $keyHash],
            [
                'created_by' => $admin->id,
                'name' => 'Hermes Autonomous DevOps',
                'agent_type' => 'hermes',
                'api_key_prefix' => 'vps_live_test...',
                'permissions' => [
                    'server.read',
                    'project.read',
                    'project.create',
                    'project.deploy',
                    'ssl.manage',
                    'service.manage',
                ],
                'is_active' => true,
            ]
        );
    }
}
