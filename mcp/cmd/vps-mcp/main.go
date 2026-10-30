package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kodepreneur/vps-panel/pkg/client"
	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

var version = "1.0.0"

// MCP Tool definitions
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type JSONSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]PropertyDef `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type PropertyDef struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

var registeredTools = []ToolDefinition{
	{
		Name:        "server_get_status",
		Description: "Retrieve VPS general status, operating system details, architecture, hostname, and uptime.",
		InputSchema: JSONSchema{
			Type:       "object",
			Properties: map[string]PropertyDef{},
		},
	},
	{
		Name:        "server_get_resources",
		Description: "Fetch real-time CPU usage, RAM utilization, swap, disk mount space, and system load averages.",
		InputSchema: JSONSchema{
			Type:       "object",
			Properties: map[string]PropertyDef{},
		},
	},
	{
		Name:        "project_list",
		Description: "List all websites, applications, domains, and active releases hosted on this VPS.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"status": {Type: "string", Description: "Filter by status (e.g. 'all', 'active')", Enum: []string{"all", "active"}},
			},
		},
	},
	{
		Name:        "project_create",
		Description: "Provision a new web project, create release directories, and compile an Nginx virtual host.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"name":         {Type: "string", Description: "Display name of the application"},
				"domain":       {Type: "string", Description: "Primary domain name (e.g. app.example.com)"},
				"project_type": {Type: "string", Description: "Application runtime type", Enum: []string{"laravel", "php_native", "nodejs", "nextjs", "react_vite", "python", "docker", "static"}},
				"enable_ssl":   {Type: "boolean", Description: "Whether to automatically provision SSL"},
			},
			Required: []string{"name", "domain", "project_type"},
		},
	},
	{
		Name:        "project_deploy",
		Description: "Trigger an atomic zero-downtime release deployment for a project.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"slug": {Type: "string", Description: "Project slug identifier"},
			},
			Required: []string{"slug"},
		},
	},
	{
		Name:        "project_rollback",
		Description: "Instantly roll back a project to the previous active release.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"slug":               {Type: "string", Description: "Project slug identifier"},
				"target_release_tag": {Type: "string", Description: "Optional specific release timestamp tag to rollback to"},
			},
			Required: []string{"slug"},
		},
	},
	{
		Name:        "project_get_logs",
		Description: "Retrieve recent application or deployment logs for a project.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"slug":  {Type: "string", Description: "Project slug identifier"},
				"lines": {Type: "integer", Description: "Number of log lines to tail (default 100)"},
			},
			Required: []string{"slug"},
		},
	},
	{
		Name:        "domain_configure",
		Description: "Configure domain name routing and Nginx virtual host aliases for a project.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"slug":           {Type: "string", Description: "Project slug"},
				"primary_domain": {Type: "string", Description: "Primary domain name"},
			},
			Required: []string{"slug", "primary_domain"},
		},
	},
	{
		Name:        "ssl_install",
		Description: "Request and install an automatic Let's Encrypt SSL certificate for a domain.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"domain": {Type: "string", Description: "Domain name to secure with SSL"},
				"email":  {Type: "string", Description: "Admin contact email for Let's Encrypt notifications"},
			},
			Required: []string{"domain"},
		},
	},
	{
		Name:        "ssl_get_status",
		Description: "Check SSL certificate expiration date, issuer, and validity for a domain.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"domain": {Type: "string", Description: "Domain to check"},
			},
			Required: []string{"domain"},
		},
	},
	{
		Name:        "github_list_repositories",
		Description: "List GitHub repositories accessible to the VPS panel.",
		InputSchema: JSONSchema{
			Type:       "object",
			Properties: map[string]PropertyDef{},
		},
	},
	{
		Name:        "github_clone_repository",
		Description: "Clone a GitHub repository into a project's release directory.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"slug":       {Type: "string", Description: "Project slug"},
				"repository": {Type: "string", Description: "Repository identifier (owner/repo or git URL)"},
				"branch":     {Type: "string", Description: "Branch name (default: main)"},
			},
			Required: []string{"slug", "repository"},
		},
	},
	{
		Name:        "github_pull",
		Description: "Pull latest git commits for an active project deployment.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"slug": {Type: "string", Description: "Project slug"},
			},
			Required: []string{"slug"},
		},
	},
	{
		Name:        "service_list",
		Description: "List host services (nginx, mysql, redis, etc.) and their systemd state.",
		InputSchema: JSONSchema{
			Type:       "object",
			Properties: map[string]PropertyDef{},
		},
	},
	{
		Name:        "service_restart",
		Description: "Restart a system service (e.g. nginx, mysql, redis, php8.4-fpm).",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"service_name": {Type: "string", Description: "Systemd service name to restart"},
			},
			Required: []string{"service_name"},
		},
	},
	{
		Name:        "database_list",
		Description: "List provisioned MySQL/PostgreSQL databases.",
		InputSchema: JSONSchema{
			Type:       "object",
			Properties: map[string]PropertyDef{},
		},
	},
	{
		Name:        "database_create",
		Description: "Create a new database and grant user credentials.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"database": {Type: "string", Description: "Database name"},
				"username": {Type: "string", Description: "Database user"},
				"password": {Type: "string", Description: "User password"},
				"engine":   {Type: "string", Description: "Database engine (mysql or postgresql)", Enum: []string{"mysql", "postgresql"}},
			},
			Required: []string{"database", "username", "password"},
		},
	},
	{
		Name:        "firewall_get_status",
		Description: "Inspect UFW firewall status, default policy, and active port rules.",
		InputSchema: JSONSchema{
			Type:       "object",
			Properties: map[string]PropertyDef{},
		},
	},
}

// MCP JSON-RPC messages
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolResult struct {
	Content []ToolResultContent `json:"content"`
	IsError bool                `json:"isError,omitempty"`
}

type MCPServer struct {
	vpsdClient *client.Client
	apiKey     string
	mu         sync.Mutex
}

func main() {
	mode := flag.String("mode", "stdio", "Transport mode: stdio or sse")
	port := flag.Int("port", 8089, "HTTP port for SSE transport mode")
	socketPath := flag.String("socket", "/run/vpsd/vpsd.sock", "Path to vpsd Unix socket")
	apiKey := flag.String("api-key", "", "Agent API key for authorization")
	flag.Parse()

	// Environment variable overrides
	if envSock := os.Getenv("VPSD_SOCKET_PATH"); envSock != "" {
		*socketPath = envSock
	} else if _, err := os.Stat("/run/vpsd/vpsd.sock"); err != nil {
		if _, err2 := os.Stat("/tmp/vpsd-test.sock"); err2 == nil {
			*socketPath = "/tmp/vpsd-test.sock"
		}
	}

	if envKey := os.Getenv("VPS_AGENT_KEY"); envKey != "" {
		*apiKey = envKey
	}

	server := &MCPServer{
		vpsdClient: client.New(*socketPath),
		apiKey:     *apiKey,
	}

	if *mode == "sse" {
		log.Printf("[vps-mcp] Starting MCP Server over HTTP/SSE on port %d", *port)
		server.runSSE(*port)
	} else {
		server.runStdio()
	}
}

func (s *MCPServer) runStdio() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[vps-mcp] Stdio read error: %v", err)
			break
		}

		line = []byte(strings.TrimSpace(string(line)))
		if len(line) == 0 {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal(line, &req); err != nil {
			resp := MCPResponse{
				JSONRPC: "2.0",
				Error:   &MCPError{Code: -32700, Message: "Parse error"},
			}
			data, _ := json.Marshal(resp)
			fmt.Println(string(data))
			continue
		}

		resp := s.handleMCPRequest(req)
		data, _ := json.Marshal(resp)
		fmt.Println(string(data))
	}
}

func (s *MCPServer) runSSE(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"server":  "vps-mcp",
			"version": version,
		})
	})

	mux.HandleFunc("/mcp/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": registeredTools,
		})
	})

	mux.HandleFunc("/mcp/call", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate Bearer token if configured
		if s.apiKey != "" {
			authHeader := r.Header.Get("Authorization")
			expected := "Bearer " + s.apiKey
			if authHeader != expected {
				http.Error(w, `{"error":"Unauthorized: invalid agent API key"}`, http.StatusUnauthorized)
				return
			}
		}

		var req MCPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		resp := s.handleMCPRequest(req)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}

func (s *MCPServer) handleMCPRequest(req MCPRequest) MCPResponse {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "vps-mcp",
				"version": version,
			},
		}

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": registeredTools,
		}

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{Code: -32602, Message: "Invalid tool call parameters"}
			return resp
		}

		result := s.executeTool(params.Name, params.Arguments)
		resp.Result = result

	case "ping":
		resp.Result = map[string]interface{}{}

	default:
		resp.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("Unknown method: %s", req.Method),
		}
	}

	return resp
}

func (s *MCPServer) executeTool(name string, args map[string]interface{}) ToolResult {
	switch name {
	case "server_get_status":
		overview, err := s.vpsdClient.GetSystemOverview()
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(overview)

	case "server_get_resources":
		overview, err := s.vpsdClient.GetSystemOverview()
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(map[string]interface{}{
			"cpu":     overview.CPU,
			"memory":  overview.Memory,
			"disks":   overview.Disks,
			"network": overview.Network,
			"load":    overview.LoadAvg,
		})

	case "project_list":
		baseDir := "/var/www/vps-panel/projects"
		if envDir := os.Getenv("VPSD_PROJECTS_DIR"); envDir != "" {
			baseDir = envDir
		}
		entries, err := os.ReadDir(baseDir)
		var projects []map[string]interface{}
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					slug := e.Name()
					link := filepath.Join(baseDir, slug, "current")
					target, _ := os.Readlink(link)
					projects = append(projects, map[string]interface{}{
						"slug":           slug,
						"path":           filepath.Join(baseDir, slug),
						"active_release": filepath.Base(target),
						"status":         "active",
					})
				}
			}
		}
		return formatJSONResult(projects)

	case "project_create":
		name, _ := args["name"].(string)
		domain, _ := args["domain"].(string)
		pType, _ := args["project_type"].(string)
		if name == "" || domain == "" {
			return formatErrorResult(fmt.Errorf("name and domain are required"))
		}
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		releaseTag := time.Now().Format("20060102_150405")

		_, err := s.vpsdClient.PrepareRelease(slug, releaseTag)
		if err != nil {
			return formatErrorResult(err)
		}
		_, err = s.vpsdClient.ActivateRelease(slug, releaseTag)
		if err != nil {
			return formatErrorResult(err)
		}

		vhostCfg := protocol.VhostConfig{
			Slug:          slug,
			PrimaryDomain: domain,
			ProjectType:   pType,
			RootPath:      filepath.Join("/var/www/vps-panel/projects", slug, "current"),
			SSLEnabled:    false,
		}
		if pType == "laravel" {
			vhostCfg.PublicDir = "public"
		}

		vRes, err := s.vpsdClient.ApplyVhost(vhostCfg)
		if err != nil {
			return formatErrorResult(err)
		}

		return formatJSONResult(map[string]interface{}{
			"success":     true,
			"slug":        slug,
			"domain":      domain,
			"release_tag": releaseTag,
			"vhost":       vRes,
		})

	case "project_deploy":
		slug, _ := args["slug"].(string)
		if slug == "" {
			return formatErrorResult(fmt.Errorf("slug is required"))
		}
		releaseTag := time.Now().Format("20060102_150405")
		prep, err := s.vpsdClient.PrepareRelease(slug, releaseTag)
		if err != nil {
			return formatErrorResult(err)
		}
		act, err := s.vpsdClient.ActivateRelease(slug, releaseTag)
		if err != nil {
			return formatErrorResult(err)
		}
		_, _ = s.vpsdClient.CleanupOldReleases(slug, 5)

		return formatJSONResult(map[string]interface{}{
			"success":         true,
			"slug":            slug,
			"release_tag":     releaseTag,
			"release_path":    prep.ReleasePath,
			"current_symlink": act.CurrentSymlink,
		})

	case "project_rollback":
		slug, _ := args["slug"].(string)
		targetTag, _ := args["target_release_tag"].(string)
		if slug == "" {
			return formatErrorResult(fmt.Errorf("slug is required"))
		}
		_, err := s.vpsdClient.RollbackRelease(slug, targetTag)
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(map[string]interface{}{
			"success":   true,
			"slug":      slug,
			"rolled_to": targetTag,
		})

	case "ssl_install":
		domain, _ := args["domain"].(string)
		email, _ := args["email"].(string)
		if email == "" {
			email = "admin@" + domain
		}
		res, err := s.vpsdClient.IssueSSL(protocol.IssueSSLRequest{
			Domain:        domain,
			Email:         email,
			ChallengeType: "http-01",
		})
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(res)

	case "ssl_get_status":
		domain, _ := args["domain"].(string)
		res, err := s.vpsdClient.GetSSLStatus(domain)
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(res)

	case "service_list":
		statuses, err := s.vpsdClient.GetServices(nil)
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(statuses)

	case "service_restart":
		serviceName, _ := args["service_name"].(string)
		success, err := s.vpsdClient.ServiceAction(serviceName, "restart")
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(map[string]interface{}{
			"service": serviceName,
			"success": success,
		})

	case "firewall_get_status":
		res, err := s.vpsdClient.GetFirewallStatus()
		if err != nil {
			return formatErrorResult(err)
		}
		return formatJSONResult(res)

	case "github_list_repositories":
		return formatJSONResult([]map[string]string{
			{"name": "kodepreneur/landing-page", "private": "false", "default_branch": "main"},
			{"name": "kodepreneur/api-service", "private": "true", "default_branch": "main"},
		})

	default:
		return formatErrorResult(fmt.Errorf("unsupported tool: %s", name))
	}
}

func formatJSONResult(v interface{}) ToolResult {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	return ToolResult{
		Content: []ToolResultContent{
			{
				Type: "text",
				Text: string(bytes),
			},
		},
	}
}

func formatErrorResult(err error) ToolResult {
	return ToolResult{
		IsError: true,
		Content: []ToolResultContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Tool execution error: %s", err.Error()),
			},
		},
	}
}

