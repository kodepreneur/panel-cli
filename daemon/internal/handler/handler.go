package handler

import (
	"encoding/json"
	"fmt"

	"github.com/kodepreneur/vps-panel/daemon/internal/database"
	"github.com/kodepreneur/vps-panel/daemon/internal/deploy"
	"github.com/kodepreneur/vps-panel/daemon/internal/firewall"
	"github.com/kodepreneur/vps-panel/daemon/internal/nginx"
	"github.com/kodepreneur/vps-panel/daemon/internal/service"
	"github.com/kodepreneur/vps-panel/daemon/internal/ssl"
	"github.com/kodepreneur/vps-panel/pkg/metrics"
	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

type Dispatcher struct {
	metricsCollector metrics.Collector
	nginxManager     *nginx.Manager
	deployManager    *deploy.Manager
	sslManager       *ssl.Manager
	serviceManager   *service.Manager
	firewallManager  *firewall.Manager
	databaseManager  *database.Manager
}

func NewDispatcher(
	availableDir, enabledDir string,
	projectsBaseDir string,
	sslLiveDir string,
) *Dispatcher {
	return &Dispatcher{
		metricsCollector: metrics.NewCollector(),
		nginxManager:     nginx.NewManager(availableDir, enabledDir),
		deployManager:    deploy.NewManager(projectsBaseDir),
		sslManager:       ssl.NewManager(sslLiveDir),
		serviceManager:   service.NewManager(),
		firewallManager:  firewall.NewManager(),
		databaseManager:  database.NewManager(),
	}
}

// Handle executes a JSON-RPC 2.0 request and returns the response
func (d *Dispatcher) Handle(req protocol.Request) protocol.Response {
	resp := protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	// --- System Methods ---
	case "System.GetOverview":
		overview, err := d.metricsCollector.GetOverview()
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = overview

	case "System.GetServices":
		var services []string
		_ = unmarshalParams(req.Params, &services)
		statuses, err := d.serviceManager.GetStatus(services)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = statuses

	case "System.ServiceAction":
		var actReq protocol.ServiceActionRequest
		if err := unmarshalParams(req.Params, &actReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		success, err := d.serviceManager.Action(actReq)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = success

	case "System.GetUpdateStatus":
		status, err := d.serviceManager.GetUpdateStatus()
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = status

	case "System.UpdateSystem":
		var uReq protocol.SystemUpdateRequest
		_ = unmarshalParams(req.Params, &uReq)
		result, err := d.serviceManager.UpdateSystem(uReq)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = result

	// --- Nginx Methods ---
	case "Nginx.ApplyVhost":
		var cfg protocol.VhostConfig
		if err := unmarshalParams(req.Params, &cfg); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		result, err := d.nginxManager.ApplyVhost(cfg)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = result

	case "Nginx.RemoveVhost":
		var slug string
		if err := unmarshalParams(req.Params, &slug); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		if err := d.nginxManager.RemoveVhost(slug); err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = true

	// --- Deployment Methods ---
	case "Deploy.PrepareRelease":
		var pReq protocol.PrepareReleaseRequest
		if err := unmarshalParams(req.Params, &pReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		result, err := d.deployManager.PrepareRelease(pReq.ProjectSlug, pReq.ReleaseTag)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = result

	case "Deploy.ActivateRelease":
		var aReq protocol.ActivateReleaseRequest
		if err := unmarshalParams(req.Params, &aReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		result, err := d.deployManager.ActivateRelease(aReq.ProjectSlug, aReq.ReleaseTag)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = result

	case "Deploy.RollbackRelease":
		var rReq protocol.RollbackRequest
		if err := unmarshalParams(req.Params, &rReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		if err := d.deployManager.RollbackRelease(rReq.ProjectSlug, rReq.TargetReleaseTag); err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = true

	case "Deploy.CleanupOldReleases":
		var cReq protocol.CleanupReleasesRequest
		if err := unmarshalParams(req.Params, &cReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		removed, err := d.deployManager.CleanupOldReleases(cReq.ProjectSlug, cReq.KeepCount)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = removed

	case "Deploy.UpdateFromRepo":
		var uReq protocol.UpdateFromRepoRequest
		if err := unmarshalParams(req.Params, &uReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		result, err := d.deployManager.UpdateFromRepo(uReq)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = result

	// --- SSL Methods ---
	case "SSL.IssueCertificate":
		var sReq protocol.IssueSSLRequest
		if err := unmarshalParams(req.Params, &sReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		result, err := d.sslManager.IssueCertificate(sReq)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = result

	case "SSL.GetStatus":
		var stReq protocol.SSLStatusRequest
		if err := unmarshalParams(req.Params, &stReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		status, err := d.sslManager.GetStatus(stReq.Domain)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = status

	// --- Database Methods ---
	case "Database.Create":
		var dbReq protocol.CreateDatabaseRequest
		if err := unmarshalParams(req.Params, &dbReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		res, err := d.databaseManager.Create(dbReq)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = res

	// --- Firewall Methods ---
	case "Firewall.GetStatus":
		res, err := d.firewallManager.GetStatus()
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = res

	case "Firewall.AddRule":
		var fReq protocol.AddFirewallRuleRequest
		if err := unmarshalParams(req.Params, &fReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		success, err := d.firewallManager.AddRule(fReq)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = success

	case "Firewall.DeleteRule":
		var delReq protocol.DeleteFirewallRuleRequest
		if err := unmarshalParams(req.Params, &delReq); err != nil {
			resp.Error = &protocol.RPCError{Code: -32602, Message: "invalid params: " + err.Error()}
			return resp
		}
		success, err := d.firewallManager.DeleteRule(delReq.RuleNumber)
		if err != nil {
			resp.Error = &protocol.RPCError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = success

	default:
		resp.Error = &protocol.RPCError{
			Code:    -32601,
			Message: fmt.Sprintf("method not found: %s", req.Method),
		}
	}

	return resp
}

func unmarshalParams(params interface{}, dest interface{}) error {
	if params == nil {
		return nil
	}
	bytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

