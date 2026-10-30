<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        // 1. Projects Table
        Schema::create('projects', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignId('user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->string('name');
            $table->string('slug')->unique();
            $table->string('domain')->unique();
            $table->json('aliases')->nullable();
            $table->string('project_type'); // laravel, php_native, nodejs, nextjs, react_vite, python, docker, static
            $table->string('install_method')->default('git'); // git, upload, empty, docker
            $table->string('repository_url')->nullable();
            $table->string('branch')->default('main');
            $table->string('root_dir');
            $table->string('public_dir')->nullable();
            $table->string('php_version')->nullable();
            $table->string('node_version')->nullable();
            $table->unsignedInteger('upstream_port')->nullable();
            $table->string('status')->default('active'); // active, maintenance, suspended
            $table->boolean('ssl_enabled')->default(false);
            $table->boolean('ssl_auto_renew')->default(true);
            $table->text('env_vars_encrypted')->nullable();
            $table->json('deploy_script')->nullable();
            $table->timestamps();
        });

        // 2. Releases Table
        Schema::create('releases', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignUuid('project_id')->constrained('projects')->cascadeOnDelete();
            $table->string('release_tag'); // e.g. 20261031_010000
            $table->string('commit_hash')->nullable();
            $table->string('commit_message')->nullable();
            $table->string('commit_author')->nullable();
            $table->string('branch')->nullable();
            $table->string('path');
            $table->string('status')->default('active'); // active, archived, failed
            $table->timestamps();
        });

        // 3. Deployments Table
        Schema::create('deployments', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignUuid('project_id')->constrained('projects')->cascadeOnDelete();
            $table->foreignUuid('release_id')->nullable()->constrained('releases')->nullOnDelete();
            $table->string('trigger_type')->default('manual'); // manual, webhook, agent, cli
            $table->string('triggered_by_type')->nullable(); // user, ai_agent, system
            $table->string('triggered_by_id')->nullable();
            $table->string('status')->default('queued'); // queued, building, activating, success, failed, cancelled
            $table->longText('build_logs')->nullable();
            $table->integer('exit_code')->nullable();
            $table->integer('duration_seconds')->nullable();
            $table->timestamp('started_at')->nullable();
            $table->timestamp('completed_at')->nullable();
            $table->timestamps();
        });

        // 4. SSL Certificates Table
        Schema::create('ssl_certificates', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignUuid('project_id')->constrained('projects')->cascadeOnDelete();
            $table->string('primary_domain');
            $table->json('san_domains')->nullable();
            $table->string('issuer')->default("Let's Encrypt");
            $table->string('cert_path');
            $table->string('key_path');
            $table->timestamp('expires_at')->nullable();
            $table->boolean('auto_renew')->default(true);
            $table->string('challenge_type')->default('http-01');
            $table->timestamp('last_renewed_at')->nullable();
            $table->string('status')->default('valid'); // valid, expired, pending
            $table->timestamps();
        });

        // 5. AI Agents Table
        Schema::create('ai_agents', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignId('created_by')->nullable()->constrained('users')->nullOnDelete();
            $table->string('name');
            $table->string('agent_type'); // hermes, openclaw, openhands, custom
            $table->string('api_key_hash')->unique();
            $table->string('api_key_prefix'); // e.g. vps_live_8f0a...
            $table->json('permissions'); // array of permission strings
            $table->json('allowed_project_ids')->nullable(); // null means all accessible
            $table->json('ip_allowlist')->nullable();
            $table->boolean('is_active')->default(true);
            $table->timestamp('expires_at')->nullable();
            $table->timestamp('last_used_at')->nullable();
            $table->timestamps();
        });

        // 6. Approval Requests Table (Human-in-the-Loop)
        Schema::create('approval_requests', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignUuid('agent_id')->constrained('ai_agents')->cascadeOnDelete();
            $table->string('action_type'); // project.delete, server.reboot, firewall.modify, terminal.execute
            $table->string('resource_type')->nullable(); // project, firewall, server
            $table->string('resource_id')->nullable();
            $table->json('payload')->nullable();
            $table->string('status')->default('pending'); // pending, approved, rejected, expired
            $table->foreignId('approved_by')->nullable()->constrained('users')->nullOnDelete();
            $table->text('rejection_reason')->nullable();
            $table->timestamp('expires_at')->nullable();
            $table->timestamps();
        });

        // 7. Agent Audit Logs Table
        Schema::create('agent_audit_logs', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignUuid('agent_id')->constrained('ai_agents')->cascadeOnDelete();
            $table->string('endpoint_or_tool');
            $table->string('method')->default('POST');
            $table->string('ip_address')->nullable();
            $table->json('request_params')->nullable();
            $table->integer('response_code')->default(200);
            $table->integer('execution_time_ms')->default(0);
            $table->timestamps();
        });

        // 8. Databases Table
        Schema::create('databases', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->string('engine')->default('mysql'); // mysql, mariadb, postgresql
            $table->string('name')->unique();
            $table->string('username');
            $table->string('charset')->default('utf8mb4');
            $table->string('collation')->default('utf8mb4_unicode_ci');
            $table->string('status')->default('active');
            $table->timestamps();
        });

        // 9. Services Table
        Schema::create('services', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->string('name')->unique(); // nginx, php8.4-fpm, mysql, redis, ufw
            $table->string('display_name');
            $table->string('systemd_unit');
            $table->boolean('is_active')->default(true);
            $table->timestamps();
        });

        // 10. Backups Table
        Schema::create('backups', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->foreignUuid('project_id')->nullable()->constrained('projects')->nullOnDelete();
            $table->foreignUuid('database_id')->nullable()->constrained('databases')->nullOnDelete();
            $table->string('backup_type'); // full, project_files, database
            $table->string('file_path');
            $table->unsignedBigInteger('file_size_bytes')->default(0);
            $table->string('storage_driver')->default('local'); // local, s3
            $table->string('status')->default('completed'); // pending, in_progress, completed, failed
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('backups');
        Schema::dropIfExists('services');
        Schema::dropIfExists('databases');
        Schema::dropIfExists('agent_audit_logs');
        Schema::dropIfExists('approval_requests');
        Schema::dropIfExists('ai_agents');
        Schema::dropIfExists('ssl_certificates');
        Schema::dropIfExists('deployments');
        Schema::dropIfExists('releases');
        Schema::dropIfExists('projects');
    }
};

