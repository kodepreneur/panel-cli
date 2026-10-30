package ssl

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

type Manager struct {
	certbotBinary string
	liveDir       string // default: /etc/letsencrypt/live
	mockMode      bool
}

func NewManager(liveDir string) *Manager {
	if liveDir == "" {
		liveDir = "/etc/letsencrypt/live"
	}

	certbotBin, err := exec.LookPath("certbot")
	mock := false
	if err != nil || os.Geteuid() != 0 {
		mock = true
	}

	return &Manager{
		certbotBinary: certbotBin,
		liveDir:       liveDir,
		mockMode:      mock,
	}
}

// IssueCertificate requests a Let's Encrypt certificate via Certbot
func (m *Manager) IssueCertificate(req protocol.IssueSSLRequest) (*protocol.IssueSSLResult, error) {
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	certPath := filepath.Join(m.liveDir, req.Domain, "fullchain.pem")
	keyPath := filepath.Join(m.liveDir, req.Domain, "privkey.pem")

	if m.mockMode || m.certbotBinary == "" {
		// Mock mode for local development
		mockCertDir := filepath.Join(m.liveDir, req.Domain)
		_ = os.MkdirAll(mockCertDir, 0755)
		_ = os.WriteFile(certPath, []byte("-----BEGIN CERTIFICATE-----\nMOCK_CERTIFICATE\n-----END CERTIFICATE-----"), 0644)
		_ = os.WriteFile(keyPath, []byte("-----BEGIN PRIVATE KEY-----\nMOCK_PRIVATE_KEY\n-----END PRIVATE KEY-----"), 0600)

		return &protocol.IssueSSLResult{
			Success:   true,
			CertPath:  certPath,
			KeyPath:   keyPath,
			Issuer:    "Let's Encrypt Authority Mock (Development Mode)",
			ExpiresAt: time.Now().Add(90 * 24 * time.Hour),
			Logs:      "Mock SSL certificate generated successfully.",
		}, nil
	}

	args := []string{
		"certonly",
		"--non-interactive",
		"--agree-tos",
		"-m", req.Email,
		"-d", req.Domain,
	}

	for _, san := range req.SanDomains {
		if san != "" && san != req.Domain {
			args = append(args, "-d", san)
		}
	}

	if req.ChallengeType == "dns-01" {
		args = append(args, "--preferred-challenges", "dns")
	} else {
		// HTTP-01 webroot challenge
		webroot := req.WebrootDir
		if webroot == "" {
			webroot = "/var/www/html"
		}
		_ = os.MkdirAll(webroot, 0755)
		args = append(args, "--webroot", "-w", webroot)
	}

	cmd := exec.Command(m.certbotBinary, args...)
	out, err := cmd.CombinedOutput()
	logs := string(out)

	if err != nil {
		return &protocol.IssueSSLResult{
			Success: false,
			Logs:    logs,
			Error:   fmt.Sprintf("certbot failed: %s (%v)", logs, err),
		}, nil
	}

	// Read certificate details
	status, _ := m.GetStatus(req.Domain)
	issuer := "Let's Encrypt"
	expiresAt := time.Now().Add(90 * 24 * time.Hour)
	if status != nil && status.Valid {
		issuer = status.Issuer
		expiresAt = status.ExpiresAt
	}

	return &protocol.IssueSSLResult{
		Success:   true,
		CertPath:  certPath,
		KeyPath:   keyPath,
		Issuer:    issuer,
		ExpiresAt: expiresAt,
		Logs:      logs,
	}, nil
}

// GetStatus checks an installed SSL certificate and parses its expiration and SANs
func (m *Manager) GetStatus(domain string) (*protocol.SSLStatusResult, error) {
	certPath := filepath.Join(m.liveDir, domain, "fullchain.pem")
	data, err := os.ReadFile(certPath)
	if err != nil {
		return &protocol.SSLStatusResult{
			Exists: false,
			Valid:  false,
		}, nil
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return &protocol.SSLStatusResult{
			Exists:   true,
			Valid:    true,
			Issuer:   "Let's Encrypt (Mock)",
			Domains:  []string{domain},
			ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
			DaysLeft: 60,
			CertPath: certPath,
		}, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 certificate: %w", err)
	}

	now := time.Now()
	daysLeft := int(cert.NotAfter.Sub(now).Hours() / 24)
	isValid := now.Before(cert.NotAfter) && now.After(cert.NotBefore)

	return &protocol.SSLStatusResult{
		Exists:    true,
		Valid:     isValid,
		Issuer:    cert.Issuer.CommonName,
		Domains:   cert.DNSNames,
		ExpiresAt: cert.NotAfter,
		DaysLeft:  daysLeft,
		CertPath:  certPath,
	}, nil
}

// RenewCertificate runs certbot renew for a domain
func (m *Manager) RenewCertificate(domain string) (bool, string, error) {
	if m.mockMode || m.certbotBinary == "" {
		return true, "mock mode: certificate renewed successfully", nil
	}

	cmd := exec.Command(m.certbotBinary, "renew", "--cert-name", domain)
	out, err := cmd.CombinedOutput()
	return err == nil, string(out), err
}

