<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class AiAgent extends Model
{
    use HasFactory, HasUuids;

    protected $guarded = [];

    protected $hidden = [
        'api_key_hash',
    ];

    protected $casts = [
        'permissions' => 'array',
        'allowed_project_ids' => 'array',
        'ip_allowlist' => 'array',
        'is_active' => 'boolean',
        'expires_at' => 'datetime',
        'last_used_at' => 'datetime',
    ];

    public function approvals(): HasMany
    {
        return $this->hasMany(ApprovalRequest::class, 'agent_id');
    }

    public function auditLogs(): HasMany
    {
        return $this->hasMany(AgentAuditLog::class, 'agent_id');
    }

    public function hasPermission(string $permission): bool
    {
        if (in_array('*', $this->permissions ?? [])) {
            return true;
        }
        return in_array($permission, $this->permissions ?? []);
    }

    public function canAccessProject(string $projectId): bool
    {
        if (empty($this->allowed_project_ids)) {
            return true; // No restriction = all projects accessible
        }
        return in_array($projectId, $this->allowed_project_ids);
    }
}

