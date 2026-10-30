# Ubuntu Server Setup Guide (22.04 LTS & 24.04 LTS)

This guide walks through configuring and running the **AI-Native VPS Management Panel** on fresh Ubuntu 22.04 LTS or 24.04 LTS instances.

---

## 1. Quick Automated Installation

Run the automated one-line installer as `root`:

```bash
curl -fsSL https://raw.githubusercontent.com/kodepreneur/vps-panel/main/scripts/install.sh | bash
```

The installer will:
1. Update system packages and install prerequisites (`nginx`, `php8.4`, `ufw`, `certbot`, `redis`).
2. Create isolated system users (`panel`, `panel-rpc`).
3. Compile the Go binaries: `vpsd`, `vpsctl`, and `vps-mcp`.
4. Register and start the `vpsd.service` systemd daemon.
5. Initialize the database schema and seed default credentials.
6. Build and deploy the React Web Dashboard to `/var/www/vps-panel/public`.
7. Configure the UFW firewall (allowing ports `22`, `80`, `443`).

---

## 2. Manual Step-by-Step Installation

### Step A: System Prerequisites

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git unzip software-properties-common ufw certbot python3-certbot-nginx
```

### Step B: PHP 8.4 and Nginx

```bash
sudo add-apt-repository -y ppa:ondrej/php
sudo apt update -y
sudo apt install -y nginx php8.4-fpm php8.4-cli php8.4-common php8.4-sqlite3 \
    php8.4-curl php8.4-mbstring php8.4-xml php8.4-zip php8.4-mysql php8.4-pgsql redis-server
```

### Step C: Service Users and Unix Socket Permissions

```bash
sudo groupadd -f panel-rpc
sudo useradd -r -s /bin/false -g panel-rpc -d /var/www/vps-panel panel || true
sudo usermod -a -G panel-rpc www-data

sudo mkdir -p /run/vpsd
sudo chown -R root:panel-rpc /run/vpsd
sudo chmod 0770 /run/vpsd

sudo mkdir -p /var/www/vps-panel/projects
sudo chown -R panel:panel-rpc /var/www/vps-panel/projects
sudo chmod -R 0775 /var/www/vps-panel/projects
```

### Step D: Compile Go Binaries

```bash
# Install Go 1.24+ if not present
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Build the binaries
go build -o /usr/local/bin/vpsd ./daemon/cmd/vpsd
go build -o /usr/local/bin/vpsctl ./cli/cmd/vpsctl
go build -o /usr/local/bin/vps-mcp ./mcp/cmd/vps-mcp

sudo chmod 0755 /usr/local/bin/vpsd /usr/local/bin/vpsctl /usr/local/bin/vps-mcp
```

### Step E: Systemd Service Registration

Create `/etc/systemd/system/vpsd.service`:

```ini
[Unit]
Description=VPS Panel Privileged Execution Host Daemon
After=network.target nginx.service

[Service]
Type=simple
User=root
Group=panel-rpc
ExecStart=/usr/local/bin/vpsd --socket=/run/vpsd/vpsd.sock --projects-dir=/var/www/vps-panel/projects
Restart=always
RestartSec=3s
UMask=0007

[Install]
WantedBy=multi-user.target
```

Enable and start the daemon:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now vpsd.service
```

### Step F: Verify Installation

Check daemon status:

```bash
vpsctl server status
```

Output:

```text
=== SERVER OVERVIEW ===
Hostname:    vps-server
OS:          Ubuntu 24.04 LTS (6.8.0-generic, x86_64)
Uptime:      1 hours 20 mins
Load Avg:    0.15, 0.22, 0.18
CPU:         AMD EPYC (4 cores) - 12.5% used
Memory:      2.1 GB / 8.0 GB (26.2% used)
Disk (/):    18.4 GB / 80.0 GB (23.0% used)
```

