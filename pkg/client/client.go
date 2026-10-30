package client

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

// Client interacts with the vpsd execution daemon over a Unix Domain Socket
type Client struct {
	socketPath string
	timeout    time.Duration
	reqID      uint64
	mu         sync.Mutex
}

// New creates a new client targeting the given Unix Domain Socket
func New(socketPath string) *Client {
	if socketPath == "" {
		socketPath = "/run/vpsd/vpsd.sock"
	}
	return &Client{
		socketPath: socketPath,
		timeout:    30 * time.Second,
	}
}

// SetTimeout configures the request timeout
func (c *Client) SetTimeout(d time.Duration) {
	c.timeout = d
}

// Call performs a JSON-RPC 2.0 request over the Unix domain socket
func (c *Client) Call(method string, params interface{}, result interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to vpsd at %s: %w", c.socketPath, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(c.timeout))

	id := fmt.Sprintf("%d", atomic.AddUint64(&c.reqID, 1))
	req := protocol.Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	decoder := json.NewDecoder(conn)
	var resp protocol.Response
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if result != nil && resp.Result != nil {
		resultBytes, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to re-marshal result: %w", err)
		}
		if err := json.Unmarshal(resultBytes, result); err != nil {
			return fmt.Errorf("failed to unmarshal into target result struct: %w", err)
		}
	}

	return nil
}

// High-level helper methods

func (c *Client) GetSystemOverview() (*protocol.SystemOverview, error) {
	var overview protocol.SystemOverview
	err := c.Call("System.GetOverview", nil, &overview)
	return &overview, err
}

func (c *Client) GetServices(services []string) ([]protocol.ServiceStatus, error) {
	var statuses []protocol.ServiceStatus
	err := c.Call("System.GetServices", services, &statuses)
	return statuses, err
}

func (c *Client) ServiceAction(service, action string) (bool, error) {
	var success bool
	err := c.Call("System.ServiceAction", protocol.ServiceActionRequest{Service: service, Action: action}, &success)
	return success, err
}

func (c *Client) ApplyVhost(cfg protocol.VhostConfig) (*protocol.VhostResult, error) {
	var res protocol.VhostResult
	err := c.Call("Nginx.ApplyVhost", cfg, &res)
	return &res, err
}

func (c *Client) RemoveVhost(slug string) (bool, error) {
	var success bool
	err := c.Call("Nginx.RemoveVhost", slug, &success)
	return success, err
}

func (c *Client) PrepareRelease(projectSlug, releaseTag string) (*protocol.PrepareReleaseResult, error) {
	var res protocol.PrepareReleaseResult
	err := c.Call("Deploy.PrepareRelease", protocol.PrepareReleaseRequest{ProjectSlug: projectSlug, ReleaseTag: releaseTag}, &res)
	return &res, err
}

func (c *Client) ActivateRelease(projectSlug, releaseTag string) (*protocol.ActivateReleaseResult, error) {
	var res protocol.ActivateReleaseResult
	err := c.Call("Deploy.ActivateRelease", protocol.ActivateReleaseRequest{ProjectSlug: projectSlug, ReleaseTag: releaseTag}, &res)
	return &res, err
}

func (c *Client) RollbackRelease(projectSlug, targetReleaseTag string) (bool, error) {
	var success bool
	err := c.Call("Deploy.RollbackRelease", protocol.RollbackRequest{ProjectSlug: projectSlug, TargetReleaseTag: targetReleaseTag}, &success)
	return success, err
}

func (c *Client) CleanupOldReleases(projectSlug string, keepCount int) ([]string, error) {
	var removed []string
	err := c.Call("Deploy.CleanupOldReleases", protocol.CleanupReleasesRequest{ProjectSlug: projectSlug, KeepCount: keepCount}, &removed)
	return removed, err
}

func (c *Client) IssueSSL(req protocol.IssueSSLRequest) (*protocol.IssueSSLResult, error) {
	var res protocol.IssueSSLResult
	err := c.Call("SSL.IssueCertificate", req, &res)
	return &res, err
}

func (c *Client) GetSSLStatus(domain string) (*protocol.SSLStatusResult, error) {
	var res protocol.SSLStatusResult
	err := c.Call("SSL.GetStatus", protocol.SSLStatusRequest{Domain: domain}, &res)
	return &res, err
}

func (c *Client) CreateDatabase(req protocol.CreateDatabaseRequest) (*protocol.DatabaseResult, error) {
	var res protocol.DatabaseResult
	err := c.Call("Database.Create", req, &res)
	return &res, err
}

func (c *Client) GetFirewallStatus() (*protocol.FirewallStatusResult, error) {
	var res protocol.FirewallStatusResult
	err := c.Call("Firewall.GetStatus", nil, &res)
	return &res, err
}

func (c *Client) AddFirewallRule(req protocol.AddFirewallRuleRequest) (bool, error) {
	var success bool
	err := c.Call("Firewall.AddRule", req, &success)
	return success, err
}

func (c *Client) DeleteFirewallRule(ruleNumber int) (bool, error) {
	var success bool
	err := c.Call("Firewall.DeleteRule", protocol.DeleteFirewallRuleRequest{RuleNumber: ruleNumber}, &success)
	return success, err
}
