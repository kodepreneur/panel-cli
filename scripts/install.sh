#!/usr/bin/env bash
# ==============================================================================
# AI-Native VPS Management Panel - Automated Installer
# Supported OS: Ubuntu 22.04 LTS / Ubuntu 24.04 LTS
# ==============================================================================

set -euo pipefail

# Visual formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}======================================================${NC}"
echo -e "${BLUE}    AI-Native VPS Management Panel Installer         ${NC}"
echo -e "${BLUE}======================================================${NC}"

# Check for root privileges
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}[ERROR] This script must be run as root.${NC}" >&2
  exit 1
fi

# Detect Ubuntu Version
if [ -f /etc/os-release ]; then
    . /etc/os-release
    if [ "$ID" != "ubuntu" ]; then
        echo -e "${YELLOW}[WARNING] This script is optimized for Ubuntu. Detected: $NAME${NC}"
    fi
else
    echo -e "${RED}[ERROR] Cannot detect operating system release.${NC}"
    exit 1
fi

echo -e "${GREEN}[1/8] Updating package index...${NC}"
apt-get update -y
apt-get install -y curl wget git unzip software-properties-common ufw certbot python3-certbot-nginx

echo -e "${GREEN}[2/8] Installing Nginx & PHP 8.4...${NC}"
add-apt-repository -y ppa:ondrej/php
apt-get update -y
apt-get install -y nginx php8.4-fpm php8.4-cli php8.4-common php8.4-sqlite3 php8.4-curl php8.4-mbstring php8.4-xml php8.4-zip php8.4-mysql php8.4-pgsql redis-server

echo -e "${GREEN}[3/8] Creating service users and group permissions...${NC}"
groupadd -f panel-rpc
useradd -r -s /bin/false -g panel-rpc -d /var/www/vps-panel panel || true
usermod -a -G panel-rpc www-data

mkdir -p /run/vpsd
mkdir -p /var/www/vps-panel/projects
mkdir -p /var/www/vps-panel/panel
mkdir -p /var/log/vps-panel
chown -R root:panel-rpc /run/vpsd
chmod 0770 /run/vpsd
chown -R panel:panel-rpc /var/www/vps-panel/projects
chmod -R 0775 /var/www/vps-panel/projects

echo -e "${GREEN}[4/8] Installing Go runtime and compiling vpsd, vpsctl, vps-mcp...${NC}"
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.24.4.linux-amd64.tar.gz -O /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin
fi

# Build binaries
cd "$(dirname "$0")/.."
go build -o /usr/local/bin/vpsd ./daemon/cmd/vpsd
go build -o /usr/local/bin/vpsctl ./cli/cmd/vpsctl
go build -o /usr/local/bin/vps-mcp ./mcp/cmd/vps-mcp

chmod 0755 /usr/local/bin/vpsd
chmod 0755 /usr/local/bin/vpsctl
chmod 0755 /usr/local/bin/vps-mcp

echo -e "${GREEN}[5/8] Setting up systemd services...${NC}"
cat <<EOF > /etc/systemd/system/vpsd.service
[Unit]
Description=VPS Panel Privileged Execution Host Daemon
After=network.target

[Service]
Type=simple
User=root
Group=panel-rpc
ExecStart=/usr/local/bin/vpsd --socket=/run/vpsd/vpsd.sock --projects-dir=/var/www/vps-panel/projects
Restart=always
RestartSec=3
UMask=0007

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now vpsd.service

echo -e "${GREEN}[6/8] Configuring Firewall (UFW)...${NC}"
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo -e "${GREEN}[7/8] Deploying Control Plane Backend...${NC}"
cp -r backend/* /var/www/vps-panel/panel/
cd /var/www/vps-panel/panel

# Install composer if missing
if ! command -v composer &> /dev/null; then
    curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
fi

composer install --no-dev --optimize-autoloader --no-interaction
php artisan migrate --force
php artisan db:seed --force
chown -R www-data:www-data /var/www/vps-panel/panel
chmod -R 0775 /var/www/vps-panel/panel/storage /var/www/vps-panel/panel/bootstrap/cache

echo -e "${GREEN}[8/8] Building & linking Web Dashboard Frontend...${NC}"
cd ../../../frontend
if ! command -v node &> /dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
    apt-get install -y nodejs
fi
npm install
npm run build
mkdir -p /var/www/vps-panel/public
cp -r dist/* /var/www/vps-panel/public/

echo -e "${BLUE}======================================================${NC}"
echo -e "${GREEN}    Installation Complete!                            ${NC}"
echo -e "${BLUE}======================================================${NC}"
echo -e "Web Dashboard:  http://$(hostname -I | awk '{print $1}')"
echo -e "CLI Command:    vpsctl server status"
echo -e "Default Admin:  admin@vpspanel.local / AdminPass123!"
echo -e "MCP Server:     /usr/local/bin/vps-mcp"
echo -e "${BLUE}======================================================${NC}"

