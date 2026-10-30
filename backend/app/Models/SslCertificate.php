<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Concerns\HasUuids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class SslCertificate extends Model
{
    use HasFactory, HasUuids;

    protected $guarded = [];

    protected $casts = [
        'san_domains' => 'array',
        'expires_at' => 'datetime',
        'last_renewed_at' => 'datetime',
        'auto_renew' => 'boolean',
    ];

    public function project(): BelongsTo
    {
        return $this->belongsTo(Project::class);
    }
}

