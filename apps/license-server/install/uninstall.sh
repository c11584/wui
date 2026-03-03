#!/bin/bash

set -e

SERVICE_NAME="wui-license"
INSTALL_DIR="/opt/wui-license"
DATA_DIR="/opt/wui-license/data"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root"
    exit 1
fi

read -p "Are you sure you want to uninstall WUI License Server? (y/N): " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo "Uninstall cancelled"
    exit 0
fi

log_info "Stopping service..."
systemctl stop ${SERVICE_NAME} 2>/dev/null || true
systemctl disable ${SERVICE_NAME} 2>/dev/null || true

log_info "Removing systemd service..."
rm -f /etc/systemd/system/${SERVICE_NAME}.service
systemctl daemon-reload

log_info "Removing files..."
rm -rf ${INSTALL_DIR}

log_info "Removing firewall rules..."
if command -v ufw >/dev/null 2>&1; then
    ufw delete allow 8080/tcp 2>/dev/null || true
elif command -v firewall-cmd >/dev/null 2>&1; then
    firewall-cmd --permanent --remove-port=8080/tcp 2>/dev/null || true
    firewall-cmd --reload 2>/dev/null || true
fi

echo ""
echo -e "${GREEN}WUI License Server has been uninstalled${NC}"
echo ""
