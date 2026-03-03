#!/bin/bash

# Quick test script for development

set -e

echo "Starting WUI development environment..."

# Create test config
mkdir -p /tmp/wui-test/{data,logs,configs}

cat > /tmp/wui-test/config.json << EOF
{
  "panel": {
    "port": 32451,
    "username": "admin",
    "password": "admin"
  },
  "xray": {
    "binPath": "/usr/local/bin/xray",
    "configPath": "/tmp/wui-test/configs"
  },
  "database": {
    "path": "/tmp/wui-test/data/wui.db"
  },
  "logs": {
    "path": "/tmp/wui-test/logs",
    "level": "debug"
  }
}
EOF

# Set environment variable
export WUI_CONFIG=/tmp/wui-test/config.json

# Start backend
echo "Starting backend on port 32451..."
cd apps/server
go run cmd/main.go
