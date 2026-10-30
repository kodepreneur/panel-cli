<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;

class Project extends Model
{
    use HasFactory, HasUuids;

    protected $guarded = [];

    protected $casts = [
        'aliases' => 'array',
        'ssl_enabled' => 'boolean',
        'ssl_auto_renew' => 'boolean',
        'deploy_script' => 'array',
    ];

    public function releases(): HasMany
    {
        return $this->hasMany(Release::class)->orderBy('created_at', 'desc');
    }

    public function deployments(): HasMany
    {
        return $this->hasMany(Deployment::class)->orderBy('created_at', 'desc');
    }

    public function sslCertificate(): HasOne
    {
        return $this->hasOne(SslCertificate::class);
    }
}

