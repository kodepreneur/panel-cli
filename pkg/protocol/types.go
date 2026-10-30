package protocol

import "time"

// JSON-RPC 2.0 envelope structures
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  interface{}     `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// System types
type OSInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Kernel   string `json:"kernel"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
}

type CPUInfo struct {
	Model        string  `json:"model"`
	Cores        int     `json:"cores"`
	UsagePercent float64 `json:"usage_percent"`
}

type MemoryInfo struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	FreeBytes      uint64  `json:"free_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	SwapTotalBytes uint64  `json:"swap_total_bytes"`
	SwapUsedBytes  uint64  `json:"swap_used_bytes"`
}

type DiskInfo struct {
	Mount        string  `json:"mount"`
	Filesystem   string  `json:"filesystem"`
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	FreeBytes    uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type NetworkInfo struct {
	Interface string `json:"interface"`
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
	RxRateSec uint64 `json:"rx_rate_sec"`
	TxRateSec uint64 `json:"tx_rate_sec"`
}

type SystemOverview struct {
	UptimeSeconds uint64        `json:"uptime_seconds"`
	OS            OSInfo        `json:"os"`
	CPU           CPUInfo       `json:"cpu"`
	Memory        MemoryInfo    `json:"memory"`
	Disks         []DiskInfo    `json:"disks"`
	Network       []NetworkInfo `json:"network"`
	LoadAvg       [3]float64    `json:"load_avg"`
	Timestamp     time.Time     `json:"timestamp"`
}

type ServiceStatus struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Active      bool   `json:"active"`
	Running     bool   `json:"running"`
	Status      string `json:"status"` // "active", "inactive", "failed", etc.
	MemoryBytes uint64 `json:"memory_bytes,omitempty"`
}

type ServiceActionRequest struct {
	Service string `json:"service"` // e.g. "nginx", "mysql", "redis", "php8.4-fpm"
	Action  string `json:"action"`  // "start", "stop", "restart", "reload", "status"
}

// Nginx Virtual Host Configuration
type VhostConfig struct {
	Slug          string   `json:"slug"`
	PrimaryDomain string   `json:"primary_domain"`
	Aliases       []string `json:"aliases"`
	ProjectType   string   `json:"project_type"` // "laravel", "php_native", "nodejs", "nextjs", "react_vite", "python", "docker", "static"
	RootPath      string   `json:"root_path"`     // e.g. "/var/www/vps-panel/projects/my-app/current"
	PublicDir     string   `json:"public_dir"`    // e.g. "public" or "dist" or ""
	UpstreamPort  int      `json:"upstream_port,omitempty"` // For Node/Python/Docker reverse proxy
	PhpVersion    string   `json:"php_version,omitempty"`   // e.g. "8.4", "8.3"
	PhpFpmSock    string   `json:"php_fpm_sock,omitempty"`  // custom sock path if any
	SSLEnabled    bool     `json:"ssl_enabled"`
	CertPath      string   `json:"cert_path,omitempty"`
	KeyPath       string   `json:"key_path,omitempty"`
	HSTS          bool     `json:"hsts"`
	CustomRules   string   `json:"custom_rules,omitempty"`
}

type VhostResult struct {
	Success      bool   `json:"success"`
	ConfigPath   string `json:"config_path"`
	SyntaxValid  bool   `json:"syntax_valid"`
	ReloadOutput string `json:"reload_output"`
	Error        string `json:"error,omitempty"`
}

// Deployment & Release Types
type PrepareReleaseRequest struct {
	ProjectSlug string `json:"project_slug"`
	ReleaseTag  string `json:"release_tag"`
}

type PrepareReleaseResult struct {
	ReleasePath string `json:"release_path"`
	SharedPath  string `json:"shared_path"`
}

type ActivateReleaseRequest struct {
	ProjectSlug string `json:"project_slug"`
	ReleaseTag  string `json:"release_tag"`
}

type ActivateReleaseResult struct {
	Success        bool   `json:"success"`
	CurrentSymlink string `json:"current_symlink"`
	TargetRelease  string `json:"target_release"`
}

type RollbackRequest struct {
	ProjectSlug      string `json:"project_slug"`
	TargetReleaseTag string `json:"target_release_tag"`
}

type CleanupReleasesRequest struct {
	ProjectSlug string `json:"project_slug"`
	KeepCount   int    `json:"keep_count"`
}

// SSL / Let's Encrypt Types
type IssueSSLRequest struct {
	Domain        string   `json:"domain"`
	SanDomains    []string `json:"san_domains,omitempty"`
	Email         string   `json:"email"`
	ChallengeType string   `json:"challenge_type"` // "http-01" or "dns-01"
	WebrootDir    string   `json:"webroot_dir,omitempty"`
}

type IssueSSLResult struct {
	Success   bool      `json:"success"`
	CertPath  string    `json:"cert_path"`
	KeyPath   string    `json:"key_path"`
	Issuer    string    `json:"issuer"`
	ExpiresAt time.Time `json:"expires_at"`
	Logs      string    `json:"logs"`
	Error     string    `json:"error,omitempty"`
}

type SSLStatusRequest struct {
	Domain string `json:"domain"`
}

type SSLStatusResult struct {
	Exists    bool      `json:"exists"`
	Valid     bool      `json:"valid"`
	Issuer    string    `json:"issuer"`
	Domains   []string  `json:"domains"`
	ExpiresAt time.Time `json:"expires_at"`
	DaysLeft  int       `json:"days_left"`
	CertPath  string    `json:"cert_path"`
}

// Database Provisioning Types
type CreateDatabaseRequest struct {
	Engine   string `json:"engine"` // "mysql", "mariadb", "postgresql"
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host,omitempty"` // default "localhost" or "%"
}

type DatabaseResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Firewall Types
type FirewallRule struct {
	Number int    `json:"number"`
	Port   string `json:"port"`
	Proto  string `json:"proto"`  // "tcp", "udp", "any"
	Action string `json:"action"` // "ALLOW", "DENY"
	From   string `json:"from"`   // "Anywhere", "192.168.1.1"
}

type FirewallStatusResult struct {
	Enabled         bool           `json:"enabled"`
	DefaultIncoming string         `json:"default_incoming"`
	Rules           []FirewallRule `json:"rules"`
}

type AddFirewallRuleRequest struct {
	Port   string `json:"port"`
	Proto  string `json:"proto"`  // "tcp", "udp"
	Action string `json:"action"` // "allow", "deny"
	FromIP string `json:"from_ip,omitempty"`
}

type DeleteFirewallRuleRequest struct {
	RuleNumber int `json:"rule_number"`
}

// UpdateFromRepo Types
type UpdateFromRepoRequest struct {
	ProjectSlug string `json:"project_slug"`
	RepoURL     string `json:"repo_url,omitempty"`
	Branch      string `json:"branch,omitempty"`
	ReleaseTag  string `json:"release_tag,omitempty"`
	RunBuild    bool   `json:"run_build,omitempty"`
}

type UpdateFromRepoResult struct {
	Success        bool   `json:"success"`
	ProjectSlug    string `json:"project_slug"`
	ReleaseTag     string `json:"release_tag"`
	CommitHash     string `json:"commit_hash"`
	CommitMessage  string `json:"commit_message"`
	CommitAuthor   string `json:"commit_author"`
	Branch         string `json:"branch"`
	Logs           string `json:"logs"`
	CurrentSymlink string `json:"current_symlink"`
	TargetRelease  string `json:"target_release"`
}

// System / Panel Update Types
type SystemUpdateStatus struct {
	CurrentVersion  string `json:"current_version"`
	CurrentCommit   string `json:"current_commit"`
	CurrentBranch   string `json:"current_branch"`
	RemoteURL       string `json:"remote_url"`
	UpdateAvailable bool   `json:"update_available"`
	LatestCommit    string `json:"latest_commit,omitempty"`
	LatestMessage   string `json:"latest_message,omitempty"`
	BehindCount     int    `json:"behind_count"`
}

type SystemUpdateRequest struct {
	Branch string `json:"branch,omitempty"`
}

type SystemUpdateResult struct {
	Success       bool   `json:"success"`
	OldCommit     string `json:"old_commit"`
	NewCommit     string `json:"new_commit"`
	UpdatedFiles  int    `json:"updated_files"`
	Logs          string `json:"logs"`
	BinariesBuilt bool   `json:"binaries_built"`
}


