package database

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

type Manager struct {
	mysqlBinary string
	psqlBinary  string
	mockMode    bool
}

func NewManager() *Manager {
	mysqlBin, _ := exec.LookPath("mysql")
	psqlBin, _ := exec.LookPath("psql")
	mock := false
	if (mysqlBin == "" && psqlBin == "") || os.Geteuid() != 0 {
		mock = true
	}

	return &Manager{
		mysqlBinary: mysqlBin,
		psqlBinary:  psqlBin,
		mockMode:    mock,
	}
}

// Create provisions a database and user with privileges
func (m *Manager) Create(req protocol.CreateDatabaseRequest) (*protocol.DatabaseResult, error) {
	if req.Database == "" || req.Username == "" {
		return nil, fmt.Errorf("database and username are required")
	}

	engine := strings.ToLower(req.Engine)
	if engine == "" {
		engine = "mysql"
	}

	// Validate identifiers to prevent SQL injection
	if !isValidIdentifier(req.Database) || !isValidIdentifier(req.Username) {
		return nil, fmt.Errorf("invalid characters in database or username")
	}

	if m.mockMode {
		return &protocol.DatabaseResult{
			Success: true,
			Message: fmt.Sprintf("mock mode: %s database '%s' and user '%s' created successfully", engine, req.Database, req.Username),
		}, nil
	}

	switch engine {
	case "mysql", "mariadb":
		return m.createMySQL(req)
	case "postgresql", "postgres":
		return m.createPostgres(req)
	default:
		return nil, fmt.Errorf("unsupported database engine: %s", engine)
	}
}

func (m *Manager) createMySQL(req protocol.CreateDatabaseRequest) (*protocol.DatabaseResult, error) {
	host := req.Host
	if host == "" {
		host = "localhost"
	}

	// Escape password single quotes
	escapedPass := strings.ReplaceAll(req.Password, "'", "\\'")

	queries := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;\n"+
			"CREATE USER IF NOT EXISTS '%s'@'%s' IDENTIFIED BY '%s';\n"+
			"GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%s';\n"+
			"FLUSH PRIVILEGES;\n",
		req.Database, req.Username, host, escapedPass, req.Database, req.Username, host,
	)

	cmd := exec.Command(m.mysqlBinary, "-uroot")
	cmd.Stdin = strings.NewReader(queries)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &protocol.DatabaseResult{
			Success: false,
			Error:   fmt.Sprintf("mysql error: %s (%v)", string(out), err),
		}, nil
	}

	return &protocol.DatabaseResult{
		Success: true,
		Message: fmt.Sprintf("MySQL database '%s' created with user '%s'", req.Database, req.Username),
	}, nil
}

func (m *Manager) createPostgres(req protocol.CreateDatabaseRequest) (*protocol.DatabaseResult, error) {
	escapedPass := strings.ReplaceAll(req.Password, "'", "''")

	queries := fmt.Sprintf(
		"CREATE DATABASE \"%s\";\n"+
			"CREATE USER \"%s\" WITH ENCRYPTED PASSWORD '%s';\n"+
			"GRANT ALL PRIVILEGES ON DATABASE \"%s\" TO \"%s\";\n",
		req.Database, req.Username, escapedPass, req.Database, req.Username,
	)

	cmd := exec.Command("su", "-", "postgres", "-c", fmt.Sprintf("psql -c %q", queries))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &protocol.DatabaseResult{
			Success: false,
			Error:   fmt.Sprintf("psql error: %s (%v)", string(out), err),
		}, nil
	}

	return &protocol.DatabaseResult{
		Success: true,
		Message: fmt.Sprintf("PostgreSQL database '%s' created with user '%s'", req.Database, req.Username),
	}, nil
}

func isValidIdentifier(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for i, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		if i > 0 && r == '-' {
			continue
		}
		return false
	}
	return true
}

