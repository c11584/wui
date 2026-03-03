#!/bin/bash

# WUI - One-Click Installation Script
# https://github.com/your-org/wui

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
PANEL_PORT=32451
PANEL_USER="admin"
PANEL_PASS="admin"
INSTALL_DIR="/opt/wui"
XRAY_VERSION="1.8.6"
WUI_VERSION="0.1.0"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --port)
            PANEL_PORT="$2"
            shift 2
            ;;
        --username)
            PANEL_USER="$2"
            shift 2
            ;;
        --password)
            PANEL_PASS="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --port PORT          Panel port (default: 32451)"
            echo "  --username USER      Admin username (default: admin)"
            echo "  --password PASS      Admin password (default: admin)"
            echo "  --install-dir DIR    Installation directory (default: /opt/wui)"
            echo "  --help               Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Print banner
print_banner() {
    echo -e "${BLUE}"
    cat << "EOF"
██╗    ██╗██╗███╗   ██╗
██║    ██║██║████╗  ██║
██║ █╗ ██║██║██╔██╗ ██║
██║███╗██║██║██║╚██╗██║
╚███╔███╔╝██║██║ ╚████║
 ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝
EOF
    echo -e "${NC}"
    echo -e "${GREEN}Next-Generation Proxy Management Panel${NC}"
    echo -e "${GREEN}=======================================${NC}"
    echo ""
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}Error: This script must be run as root${NC}"
        exit 1
    fi
}

# Detect OS
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        VER=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VER=$(lsb_release -sr)
    else
        echo -e "${RED}Cannot detect OS${NC}"
        exit 1
    fi
    echo -e "${GREEN}Detected OS: $OS $VER${NC}"
}

# Install dependencies
install_dependencies() {
    echo -e "${YELLOW}Installing dependencies...${NC}"
    
    case $OS in
        ubuntu|debian)
            apt-get update -y
            apt-get install -y curl wget unzip
            ;;
        centos|rhel|rocky|almalinux)
            yum install -y curl wget unzip
            ;;
        fedora)
            dnf install -y curl wget unzip
            ;;
        arch|manjaro)
            pacman -Sy --noconfirm curl wget unzip
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}Dependencies installed${NC}"
}

# Install Xray core
install_xray() {
    echo -e "${YELLOW}Installing Xray core v${XRAY_VERSION}...${NC}"
    
    # Detect architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            XRAY_ARCH="64"
            ;;
        aarch64|arm64)
            XRAY_ARCH="arm64-v8a"
            ;;
        armv7l|armhf)
            XRAY_ARCH="arm32-v7a"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    # Download Xray
    XRAY_URL="https://github.com/XTLS/Xray-core/releases/download/v${XRAY_VERSION}/Xray-linux-${XRAY_ARCH}.zip"
    TMP_DIR="/tmp/xray-$$"
    
    mkdir -p $TMP_DIR
    wget -O $TMP_DIR/xray.zip $XRAY_URL
    unzip -o $TMP_DIR/xray.zip -d $TMP_DIR
    
    # Install Xray
    mkdir -p $INSTALL_DIR/bin
    mv $TMP_DIR/xray $INSTALL_DIR/bin/
    chmod +x $INSTALL_DIR/bin/xray
    
    # Cleanup
    rm -rf $TMP_DIR
    
    echo -e "${GREEN}Xray core installed${NC}"
}

# Download WUI panel
install_panel() {
    echo -e "${YELLOW}Installing WUI panel...${NC}"
    
    # Detect architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            WUI_ARCH="amd64"
            ;;
        aarch64|arm64)
            WUI_ARCH="arm64"
            ;;
        armv7l|armhf)
            WUI_ARCH="arm"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    # Download latest release
    WUI_URL="https://github.com/your-org/wui/releases/download/v${WUI_VERSION}/wui-linux-${WUI_ARCH}-${WUI_VERSION}.tar.gz"
    TMP_DIR="/tmp/wui-$$"
    
    mkdir -p $TMP_DIR
    
    # Try to download, fallback to local build for testing
    if ! wget -O $TMP_DIR/wui.tar.gz $WUI_URL 2>/dev/null; then
        echo -e "${YELLOW}Pre-built binary not found, using local build${NC}"
        # For testing, assume binary is in current directory
        if [[ -f "./wui" ]]; then
            cp ./wui $TMP_DIR/
        else
            echo -e "${RED}WUI binary not found. Please build it first.${NC}"
            exit 1
        fi
    else
        tar -xzf $TMP_DIR/wui.tar.gz -C $TMP_DIR
    fi
    
    # Install panel
    mkdir -p $INSTALL_DIR
    if [[ -f "$TMP_DIR/wui" ]]; then
        mv $TMP_DIR/wui $INSTALL_DIR/
    fi
    chmod +x $INSTALL_DIR/wui
    
    # Copy web assets if available
    if [[ -d "$TMP_DIR/web" ]]; then
        cp -r $TMP_DIR/web $INSTALL_DIR/
    fi
    
    # Create necessary directories
    mkdir -p $INSTALL_DIR/{data,logs,configs,bin}
    
    # Cleanup
    rm -rf $TMP_DIR
    
    echo -e "${GREEN}WUI panel installed${NC}"
}

# Create configuration
create_config() {
    echo -e "${YELLOW}Creating configuration...${NC}"
    
    cat > $INSTALL_DIR/config.json << EOF
{
  "panel": {
    "port": $PANEL_PORT,
    "username": "$PANEL_USER",
    "password": "$PANEL_PASS"
  },
  "xray": {
    "binPath": "$INSTALL_DIR/bin/xray",
    "configPath": "$INSTALL_DIR/configs"
  },
  "database": {
    "path": "$INSTALL_DIR/data/wui.db"
  },
  "logs": {
    "path": "$INSTALL_DIR/logs",
    "level": "info"
  }
}
EOF
    
    echo -e "${GREEN}Configuration created${NC}"
}

# Setup systemd service
setup_service() {
    echo -e "${YELLOW}Setting up systemd service...${NC}"
    
    cat > /etc/systemd/system/wui.service << EOF
[Unit]
Description=WUI - Next-Generation Proxy Management Panel
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/wui
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
    
    # Enable and start service
    systemctl daemon-reload
    systemctl enable wui
    
    echo -e "${GREEN}Systemd service configured${NC}"
}

# Configure firewall
setup_firewall() {
    echo -e "${YELLOW}Configuring firewall...${NC}"
    
    if command -v ufw >/dev/null 2>&1; then
        ufw allow $PANEL_PORT/tcp comment "WUI Panel"
        echo -e "${GREEN}UFW firewall configured${NC}"
    elif command -v firewall-cmd >/dev/null 2>&1; then
        firewall-cmd --permanent --add-port=$PANEL_PORT/tcp
        firewall-cmd --reload
        echo -e "${GREEN}Firewalld configured${NC}"
    else
        echo -e "${YELLOW}No firewall detected, skipping...${NC}"
    fi
}

# Start service
start_service() {
    echo -e "${YELLOW}Starting WUI service...${NC}"
    systemctl start wui
    
    # Wait for service to start
    sleep 2
    
    if systemctl is-active --quiet wui; then
        echo -e "${GREEN}WUI service started successfully${NC}"
    else
        echo -e "${RED}Failed to start WUI service${NC}"
        journalctl -u wui -n 20 --no-pager
        exit 1
    fi
}

# Show success message
show_success() {
    SERVER_IP=$(curl -s ifconfig.me || echo "YOUR_SERVER_IP")
    
    echo ""
    echo -e "${GREEN}=======================================${NC}"
    echo -e "${GREEN}WUI Installation Complete!${NC}"
    echo -e "${GREEN}=======================================${NC}"
    echo ""
    echo -e "${BLUE}Panel URL:${NC} http://$SERVER_IP:$PANEL_PORT"
    echo -e "${BLUE}Username:${NC}  $PANEL_USER"
    echo -e "${BLUE}Password:${NC}  $PANEL_PASS"
    echo ""
    echo -e "${YELLOW}Installation Directory:${NC} $INSTALL_DIR"
    echo ""
    echo -e "${YELLOW}Commands:${NC}"
    echo "  Start:   systemctl start wui"
    echo "  Stop:    systemctl stop wui"
    echo "  Restart: systemctl restart wui"
    echo "  Status:  systemctl status wui"
    echo "  Logs:    journalctl -u wui -f"
    echo ""
    echo -e "${RED}⚠️  IMPORTANT: Change the default password immediately after login!${NC}"
    echo ""
}

# Main installation process
main() {
    print_banner
    
    echo -e "${YELLOW}Installation Configuration:${NC}"
    echo "  Port:         $PANEL_PORT"
    echo "  Username:     $PANEL_USER"
    echo "  Password:     $PANEL_PASS"
    echo "  Install Dir:  $INSTALL_DIR"
    echo ""
    
    read -p "Continue with installation? (y/n): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled"
        exit 1
    fi
    
    check_root
    detect_os
    install_dependencies
    install_xray
    install_panel
    create_config
    setup_service
    setup_firewall
    start_service
    show_success
}

# Run main function
main
