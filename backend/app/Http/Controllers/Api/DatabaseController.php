<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Database;
use App\Services\VpsDaemonService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class DatabaseController extends Controller
{
    protected VpsDaemonService $daemon;

    public function __construct(VpsDaemonService $daemon)
    {
        $this->daemon = $daemon;
    }

    public function index(): JsonResponse
    {
        $databases = Database::latest()->get();

        return response()->json([
            'status' => 'success',
            'data' => $databases,
        ]);
    }

    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'name' => 'required|string|max:64|unique:databases,name',
            'username' => 'required|string|max:64',
            'password' => 'required|string|min:8',
            'engine' => 'nullable|string|in:mysql,mariadb,postgresql',
        ]);

        $engine = $validated['engine'] ?? 'mysql';

        $res = $this->daemon->createDatabase(
            $engine,
            $validated['name'],
            $validated['username'],
            $validated['password']
        );

        if (!($res['success'] ?? false)) {
            return response()->json([
                'status' => 'error',
                'message' => 'Failed to create database on host.',
                'error' => $res['error'] ?? 'Unknown database error',
            ], 422);
        }

        $db = Database::create([
            'name' => $validated['name'],
            'username' => $validated['username'],
            'engine' => $engine,
            'status' => 'active',
        ]);

        return response()->json([
            'status' => 'success',
            'message' => "Database '{$db->name}' created successfully.",
            'data' => $db,
        ], 201);
    }
}

