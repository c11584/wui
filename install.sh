#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

WUI_VERSION="1.0.0"
INSTALL_DIR="/opt/wui"
DATA_DIR="/opt/wui/data"
PORT=32452
USERNAME="admin"
PASSWORD=""

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

check_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        VER=$VERSION_ID
    else
        log_error "Cannot detect OS"
        exit 1
    fi
    log_info "Detected OS: $OS $VER"
}

install_deps() {
    log_info "Installing dependencies..."
    
    case $OS in
        ubuntu|debian)
            apt-get update -y
            apt-get install -y curl wget unzip tar
            ;;
        centos|rhel|rocky|almalinux)
            yum install -y curl wget unzip tar
            ;;
        alpine)
            apk add --no-cache curl wget unzip tar
            ;;
        *)
            log_warn "Unknown OS, assuming dependencies are installed"
            ;;
    esac
}

install_docker() {
    if command -v docker &> /dev/null; then
        log_success "Docker already installed"
        return
    fi
    
    log_info "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
    log_success "Docker installed"
}

install_xray() {
    if [[ -f /usr/local/bin/xray ]]; then
        log_success "Xray already installed"
        return
    fi
    
    log_info "Installing Xray..."
    bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install
    log_success "Xray installed"
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --port)
                PORT="$2"
                shift 2
                ;;
            --username)
                USERNAME="$2"
                shift 2
                ;;
            --password)
                PASSWORD="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --port PORT       Panel port (default: 32452)"
                echo "  --username USER   Admin username (default: admin)"
                echo "  --password PASS   Admin password (default: auto-generated)"
                echo "  --help            Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    if [[ -z "$PASSWORD" ]]; then
        PASSWORD=$(openssl rand -base64 12 | tr -d '/+=' | head -c 16)
    fi
}

create_config() {
    log_info "Creating configuration..."
    
    mkdir -p $DATA_DIR
    mkdir -p $DATA_DIR/xray
    mkdir -p $DATA_DIR/logs
    
    cat > $DATA_DIR/config.json << EOF
{
  "panel": {
    "port": $PORT,
    "username": "$USERNAME",
    "password": "$PASSWORD"
  },
  "xray": {
    "binPath": "/usr/local/bin/xray",
    "configPath": "$DATA_DIR/xray"
  },
  "database": {
    "path": "$DATA_DIR/wui.db"
  },
  "logs": {
    "path": "$DATA_DIR/logs",
    "level": "info"
  },
  "license": {
    "serverUrl": "http://localhost:32453",
    "gracePeriodDays": 7
  }
}
EOF
    
    log_success "Configuration created at $DATA_DIR/config.json"
}

download_binary() {
    log_info "Downloading WUI binary..."
    
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    DOWNLOAD_URL="https://github.com/your-org/wui/releases/download/v${WUI_VERSION}/wui-linux-${ARCH}.tar.gz"
    
    cd /tmp
    wget -q --show-progress "$DOWNLOAD_URL" -O wui.tar.gz || {
        log_error "Failed to download binary"
        exit 1
    }
    
    tar -xzf wui.tar.gz
    mv wui-server /usr/local/bin/
    chmod +x /usr/local/bin/wui-server
    
    rm -f wui.tar.gz
    
    log_success "Binary installed to /usr/local/bin/wui-server"
}

create_service() {
    log_info "Creating systemd service..."
    
    cat > /etc/systemd/system/wui.service << EOF
[Unit]
Description=WUI Panel
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/wui-server
Environment=WUI_CONFIG=$DATA_DIR/config.json
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable wui
    systemctl start wui
    
    log_success "Service created and started"
}

show_result() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}         WUI Installation Complete!     ${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "Panel URL:    ${BLUE}http://$(hostname -I | awk '{print $1}'):$PORT${NC}"
    echo -e "Username:     ${BLUE}$USERNAME${NC}"
    echo -e "Password:     ${BLUE}$PASSWORD${NC}"
    echo ""
    echo -e "Config file:  ${BLUE}$DATA_DIR/config.json${NC}"
    echo -e "Database:     ${BLUE}$DATA_DIR/wui.db${NC}"
    echo -e "Logs:         ${BLUE}$DATA_DIR/logs/${NC}"
    echo ""
    echo -e "Commands:"
    echo -e "  Start:   ${BLUE}systemctl start wui${NC}"
    echo -e "  Stop:    ${BLUE}systemctl stop wui${NC}"
    echo -e "  Restart: ${BLUE}systemctl restart wui${NC}"
    echo -e "  Status:  ${BLUE}systemctl status wui${NC}"
    echo -e "  Logs:    ${BLUE}journalctl -u wui -f${NC}"
    echo ""
}

main() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}     WUI Panel Installation Script      ${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    check_root
    check_os
    parse_args "$@"
    install_deps
    install_xray
    create_config
    download_binary
    create_service
    show_result
}

main "$@"
