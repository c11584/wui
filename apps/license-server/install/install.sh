#!/bin/bash

set -e

REPO_URL="https://github.com/your-org/wui"
LICENSE_SERVER_VERSION="0.1.0"
INSTALL_DIR="/opt/wui-license"
DATA_DIR="/opt/wui-license/data"
LOG_DIR="/opt/wui-license/logs"
SERVICE_NAME="wui-license"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
    elif [[ -f /etc/centos-release ]]; then
        OS="centos"
    elif [[ -f /etc/debian_version ]]; then
        OS="debian"
    else
        OS="unknown"
    fi
}

install_dependencies() {
    log_info "Installing dependencies..."
    
    case $OS in
        ubuntu|debian)
            apt-get update -y
            apt-get install -y curl wget
            ;;
        centos|rhel)
            yum install -y curl wget
            ;;
        arch)
            pacman -Sy --noconfirm curl wget
            ;;
        *)
            log_warn "Unknown OS, skipping dependency installation"
            ;;
    esac
}

detect_arch() {
    case $(uname -m) in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armhf)
            ARCH="arm"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

download_binary() {
    log_info "Downloading WUI License Server for $ARCH..."
    
    BINARY_URL="${REPO_URL}/releases/download/v${LICENSE_SERVER_VERSION}/wui-license-linux-${ARCH}-${LICENSE_SERVER_VERSION}.tar.gz"
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$LOG_DIR"
    
    cd /tmp
    
    if curl -L -o wui-license.tar.gz "$BINARY_URL"; then
        log_info "Download successful"
    else
        log_error "Failed to download WUI License Server"
        exit 1
    fi
    
    tar -xzf wui-license.tar.gz
    
    mv wui-license "$INSTALL_DIR/wui-license"
    chmod +x "$INSTALL_DIR/wui-license"
    
    rm -f wui-license.tar.gz
}

create_config() {
    log_info "Creating configuration..."
    
    CONFIG_FILE="$DATA_DIR/config.json"
    
    if [[ ! -f "$CONFIG_FILE" ]]; then
        cat > "$CONFIG_FILE" <<EOF
{
  "server": {
    "port": 8080,
    "host": "0.0.0.0"
  },
  "database": {
    "path": "$DATA_DIR/license.db"
  },
  "logs": {
    "path": "$LOG_DIR",
    "level": "info"
  },
  "jwt": {
    "secret": "$(openssl rand -hex 32)"
  }
}
EOF
        log_info "Configuration created at $CONFIG_FILE"
    else
        log_warn "Configuration file already exists, skipping"
    fi
}

create_systemd_service() {
    log_info "Creating systemd service..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=WUI License Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/wui-license
Restart=on-failure
RestartSec=5s

Environment=LICENSE_SERVER_PORT=8080
Environment=LICENSE_DB_PATH=${DATA_DIR}/license.db

StandardOutput=append:${LOG_DIR}/wui-license.log
StandardError=append:${LOG_DIR}/wui-license.log

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable ${SERVICE_NAME}
    systemctl start ${SERVICE_NAME}
    
    log_info "Service started successfully"
}

configure_firewall() {
    log_info "Configuring firewall..."
    
    if command -v ufw >/dev/null 2>&1; then
        ufw allow 8080/tcp
        log_info "UFW: Allowed port 8080"
    elif command -v firewall-cmd >/dev/null 2>&1; then
        firewall-cmd --permanent --add-port=8080/tcp
        firewall-cmd --reload
        log_info "Firewalld: Allowed port 8080"
    else
        log_warn "No firewall detected, please manually allow port 8080"
    fi
}

print_success() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}WUI License Server Installation Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "License Server Port: 8080"
    echo "Config File: $DATA_DIR/config.json"
    echo "Database: $DATA_DIR/license.db"
    echo "Logs: $LOG_DIR/"
    echo ""
    echo "Commands:"
    echo "  Start:   systemctl start ${SERVICE_NAME}"
    echo "  Stop:    systemctl stop ${SERVICE_NAME}"
    echo "  Restart: systemctl restart ${SERVICE_NAME}"
    echo "  Status:  systemctl status ${SERVICE_NAME}"
    echo "  Logs:    journalctl -u ${SERVICE_NAME} -f"
    echo ""
}

main() {
    check_root
    detect_os
    install_dependencies
    detect_arch
    download_binary
    create_config
    create_systemd_service
    configure_firewall
    print_success
}

main "$@"
