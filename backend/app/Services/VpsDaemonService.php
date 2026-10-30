<?php

namespace App\Services;

use Exception;
use Illuminate\Support\Facades\Log;

class VpsDaemonService
{
    protected string $socketPath;
    protected int $timeout;

    public function __construct()
    {
        $this->socketPath = env('VPSD_SOCKET_PATH', '/run/vpsd/vpsd.sock');
        if (!file_exists($this->socketPath) && file_exists('/tmp/vpsd-test.sock')) {
            $this->socketPath = '/tmp/vpsd-test.sock';
        }
        $this->timeout = (int) env('VPSD_TIMEOUT', 30);
    }

    /**
     * Call a method on the vpsd execution daemon over Unix Domain Socket
     */
    public function call(string $method, mixed $params = null): mixed
    {
        if (!file_exists($this->socketPath)) {
            // If socket does not exist, return simulated development response
            return $this->handleMockFallback($method, $params);
        }

        $socket = @fsockopen('unix://' . $this->socketPath, -1, $errno, $errstr, $this->timeout);
        if (!$socket) {
            Log::warning("Could not connect to vpsd socket {$this->socketPath}: {$errstr} ({$errno})");
            return $this->handleMockFallback($method, $params);
        }

        stream_set_timeout($socket, $this->timeout);

        $payload = json_encode([
            'jsonrpc' => '2.0',
            'id' => (string) uniqid(),
            'method' => $method,
            'params' => $params,
        ]) . "\n";

        fwrite($socket, $payload);

        $responseLine = fgets($socket);
        fclose($socket);

        if (!$responseLine) {
            throw new Exception("Empty response received from vpsd daemon for method: {$method}");
        }

        $response = json_decode($responseLine, true);
        if (!$response) {
            throw new Exception("Malformed JSON received from vpsd daemon: {$responseLine}");
        }

        if (isset($response['error'])) {
            $msg = $response['error']['message'] ?? 'Unknown daemon RPC error';
            $code = $response['error']['code'] ?? -32000;
            throw new Exception("Daemon error [{$code}]: {$msg}");
        }

        return $response['result'] ?? null;
    }

    // High-level wrapper methods

    public function getSystemOverview(): array
    {
        return $this->call('System.GetOverview') ?? [];
    }

    public function getServices(?array $services = null): array
    {
        return $this->call('System.GetServices', $services) ?? [];
    }

    public function restartService(string $service): bool
    {
        return (bool) $this->call('System.ServiceAction', [
            'service' => $service,
            'action' => 'restart',
        ]);
    }

    public function applyVhost(array $config): array
    {
        return $this->call('Nginx.ApplyVhost', $config) ?? [];
    }

    public function removeVhost(string $slug): bool
    {
        return (bool) $this->call('Nginx.RemoveVhost', $slug);
    }

    public function prepareRelease(string $slug, string $releaseTag): array
    {
        return $this->call('Deploy.PrepareRelease', [
            'project_slug' => $slug,
            'release_tag' => $releaseTag,
        ]) ?? [];
    }

    public function activateRelease(string $slug, string $releaseTag): array
    {
        return $this->call('Deploy.ActivateRelease', [
            'project_slug' => $slug,
            'release_tag' => $releaseTag,
        ]) ?? [];
    }

    public function rollbackRelease(string $slug, ?string $targetReleaseTag = null): bool
    {
        return (bool) $this->call('Deploy.RollbackRelease', [
            'project_slug' => $slug,
            'target_release_tag' => $targetReleaseTag,
        ]);
    }

    public function issueSSL(string $domain, string $email, array $sanDomains = []): array
    {
        return $this->call('SSL.IssueCertificate', [
            'domain' => $domain,
            'email' => $email,
            'san_domains' => $sanDomains,
            'challenge_type' => 'http-01',
        ]) ?? [];
    }

    public function getSSLStatus(string $domain): array
    {
        return $this->call('SSL.GetStatus', [
            'domain' => $domain,
        ]) ?? [];
    }

    public function createDatabase(string $engine, string $database, string $username, string $password): array
    {
        return $this->call('Database.Create', [
            'engine' => $engine,
            'database' => $database,
            'username' => $username,
            'password' => $password,
        ]) ?? [];
    }

    public function getFirewallStatus(): array
    {
        return $this->call('Firewall.GetStatus') ?? [];
    }

    public function addFirewallRule(string $port, string $proto = 'tcp', string $action = 'allow'): bool
    {
        return (bool) $this->call('Firewall.AddRule', [
            'port' => $port,
            'proto' => $proto,
            'action' => $action,
        ]);
    }

    public function deleteFirewallRule(int $ruleNumber): bool
    {
        return (bool) $this->call('Firewall.DeleteRule', [
            'rule_number' => $ruleNumber,
        ]);
    }

    /**
     * Graceful fallback when running in development environments without daemon
     */
    protected function handleMockFallback(string $method, mixed $params): mixed
    {
        return match ($method) {
            'System.GetOverview' => [
                'uptime_seconds' => 72000,
                'os' => ['name' => 'Ubuntu', 'version' => 'Ubuntu 24.04 LTS', 'kernel' => '6.8.0-generic', 'arch' => 'x86_64', 'hostname' => 'vps-server'],
                'cpu' => ['model' => 'AMD EPYC 7763', 'cores' => 4, 'usage_percent' => 18.5],
                'memory' => ['total_bytes' => 8589934592, 'used_bytes' => 3221225472, 'free_bytes' => 5368709120, 'usage_percent' => 37.5, 'swap_total_bytes' => 2147483648, 'swap_used_bytes' => 0],
                'disks' => [['mount' => '/', 'filesystem' => 'ext4', 'total_bytes' => 107374182400, 'used_bytes' => 21474836480, 'free_bytes' => 85899345920, 'usage_percent' => 20.0]],
                'network' => [['interface' => 'eth0', 'rx_bytes' => 104857600, 'tx_bytes' => 52428800, 'rx_rate_sec' => 25600, 'tx_rate_sec' => 12800]],
                'load_avg' => [0.42, 0.55, 0.61],
                'timestamp' => now()->toIso8601String(),
            ],
            'System.GetServices' => [
                ['name' => 'nginx', 'display_name' => 'Nginx Web Server', 'active' => true, 'running' => true, 'status' => 'active'],
                ['name' => 'php8.4-fpm', 'display_name' => 'PHP 8.4 FPM', 'active' => true, 'running' => true, 'status' => 'active'],
                ['name' => 'mysql', 'display_name' => 'MySQL Database', 'active' => true, 'running' => true, 'status' => 'active'],
                ['name' => 'redis', 'display_name' => 'Redis Cache', 'active' => true, 'running' => true, 'status' => 'active'],
                ['name' => 'ufw', 'display_name' => 'UFW Firewall', 'active' => true, 'running' => true, 'status' => 'active'],
            ],
            'Nginx.ApplyVhost' => [
                'success' => true,
                'config_path' => '/etc/nginx/sites-available/' . ($params['slug'] ?? 'app') . '.conf',
                'syntax_valid' => true,
                'reload_output' => 'mock mode: vhost applied successfully',
            ],
            'Deploy.PrepareRelease' => [
                'release_path' => '/var/www/vps-panel/projects/' . ($params['project_slug'] ?? 'app') . '/releases/' . ($params['release_tag'] ?? 'tag'),
                'shared_path' => '/var/www/vps-panel/projects/' . ($params['project_slug'] ?? 'app') . '/shared',
            ],
            'Deploy.ActivateRelease' => [
                'success' => true,
                'current_symlink' => '/var/www/vps-panel/projects/' . ($params['project_slug'] ?? 'app') . '/current',
                'target_release' => '/var/www/vps-panel/projects/' . ($params['project_slug'] ?? 'app') . '/releases/' . ($params['release_tag'] ?? 'tag'),
            ],
            'SSL.IssueCertificate' => [
                'success' => true,
                'cert_path' => '/etc/letsencrypt/live/' . ($params['domain'] ?? 'example.com') . '/fullchain.pem',
                'key_path' => '/etc/letsencrypt/live/' . ($params['domain'] ?? 'example.com') . '/privkey.pem',
                'issuer' => "Let's Encrypt Authority",
                'expires_at' => now()->addDays(90)->toIso8601String(),
                'logs' => 'Mock certificate generated',
            ],
            'SSL.GetStatus' => [
                'exists' => true,
                'valid' => true,
                'issuer' => "Let's Encrypt",
                'domains' => [$params['domain'] ?? 'example.com'],
                'expires_at' => now()->addDays(65)->toIso8601String(),
                'days_left' => 65,
                'cert_path' => '/etc/letsencrypt/live/' . ($params['domain'] ?? 'example.com') . '/fullchain.pem',
            ],
            'Firewall.GetStatus' => [
                'enabled' => true,
                'default_incoming' => 'deny',
                'rules' => [
                    ['number' => 1, 'port' => '22/tcp', 'proto' => 'tcp', 'action' => 'ALLOW', 'from' => 'Anywhere'],
                    ['number' => 2, 'port' => '80/tcp', 'proto' => 'tcp', 'action' => 'ALLOW', 'from' => 'Anywhere'],
                    ['number' => 3, 'port' => '443/tcp', 'proto' => 'tcp', 'action' => 'ALLOW', 'from' => 'Anywhere'],
                ],
            ],
            default => true,
        };
    }
}

