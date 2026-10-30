package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"

	"github.com/kodepreneur/vps-panel/daemon/internal/handler"
	"github.com/kodepreneur/vps-panel/pkg/protocol"
)

var version = "1.0.0"

func main() {
	socketPath := flag.String("socket", "/run/vpsd/vpsd.sock", "Path to Unix Domain Socket")
	availableDir := flag.String("nginx-available", "/etc/nginx/sites-available", "Nginx sites-available directory")
	enabledDir := flag.String("nginx-enabled", "/etc/nginx/sites-enabled", "Nginx sites-enabled directory")
	projectsDir := flag.String("projects-dir", "/var/www/vps-panel/projects", "Base directory for hosted projects")
	sslLiveDir := flag.String("ssl-dir", "/etc/letsencrypt/live", "Let's Encrypt live directory")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("vpsd version %s\n", version)
		os.Exit(0)
	}

	// Environment variable overrides
	if envSock := os.Getenv("VPSD_SOCKET_PATH"); envSock != "" {
		*socketPath = envSock
	}
	if envProjects := os.Getenv("VPSD_PROJECTS_DIR"); envProjects != "" {
		*projectsDir = envProjects
	}

	log.Printf("[vpsd] Starting VPS Management Host Daemon v%s", version)
	log.Printf("[vpsd] Socket path: %s", *socketPath)

	// Ensure socket directory exists
	sockDir := filepath.Dir(*socketPath)
	if err := os.MkdirAll(sockDir, 0755); err != nil {
		log.Fatalf("[vpsd] Failed to create socket directory: %v", err)
	}

	// Clean up stale socket file if it exists
	if _, err := os.Stat(*socketPath); err == nil {
		_ = os.Remove(*socketPath)
	}

	listener, err := net.Listen("unix", *socketPath)
	if err != nil {
		log.Fatalf("[vpsd] Failed to listen on unix socket %s: %v", *socketPath, err)
	}
	defer listener.Close()

	// Set permissions on the socket (0660 so root and panel-rpc group can read/write)
	_ = os.Chmod(*socketPath, 0660)

	dispatcher := handler.NewDispatcher(*availableDir, *enabledDir, *projectsDir, *sslLiveDir)

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var shuttingDown uint32

	go func() {
		<-sigChan
		atomic.StoreUint32(&shuttingDown, 1)
		log.Println("[vpsd] Shutting down daemon gracefully...")
		_ = listener.Close()
		_ = os.Remove(*socketPath)
		os.Exit(0)
	}()

	log.Printf("[vpsd] Ready and listening for privileged RPC commands")

	for {
		conn, err := listener.Accept()
		if err != nil {
			if atomic.LoadUint32(&shuttingDown) == 1 || errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("[vpsd] Connection accept error: %v", err)
			continue
		}

		go handleConnection(conn, dispatcher)
	}

}

func handleConnection(conn net.Conn, d *handler.Dispatcher) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req protocol.Request
		if err := decoder.Decode(&req); err != nil {
			if err != io.EOF {
				log.Printf("[vpsd] JSON decode error: %v", err)
			}
			break
		}

		resp := d.Handle(req)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("[vpsd] JSON encode error: %v", err)
			break
		}
	}
}

