#!/bin/bash

# Build script for WUI

set -e

VERSION=${VERSION:-"0.1.0"}
BUILD_DIR="build"
BINARY_NAME="wui"

echo "Building WUI v${VERSION}..."

# Clean build directory
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Build frontend
echo "Building frontend..."
cd apps/web
npm install
npm run build
cd ../..

# Copy frontend build to server
cp -r apps/web/dist $BUILD_DIR/web

# Build backend (Go)
echo "Building backend..."
cd apps/server

# Download dependencies
go mod download

# Build for multiple platforms
PLATFORMS=("linux/amd64" "linux/arm64" "linux/arm")

for PLATFORM in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
  
  OUTPUT_NAME="../${BUILD_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}"
  
  if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
  fi
  
  echo "Building for $GOOS/$GOARCH..."
  GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o $OUTPUT_NAME cmd/main.go
done

cd ..

# Create release packages
echo "Creating release packages..."
cd $BUILD_DIR

for PLATFORM in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
  
  PACKAGE_NAME="${BINARY_NAME}-${GOOS}-${GOARCH}-${VERSION}"
  
  mkdir -p $PACKAGE_NAME
  cp ${BINARY_NAME}-${GOOS}-${GOARCH} $PACKAGE_NAME/${BINARY_NAME}
  cp -r web $PACKAGE_NAME/
  
  # Copy default config
  cat > $PACKAGE_NAME/config.json << EOF
{
  "panel": {
    "port": 32451,
    "username": "admin",
    "password": "admin"
  },
  "xray": {
    "binPath": "/opt/wui/bin/xray",
    "configPath": "/opt/wui/configs"
  },
  "database": {
    "path": "/opt/wui/data/wui.db"
  },
  "logs": {
    "path": "/opt/wui/logs",
    "level": "info"
  }
}
EOF
  
  tar -czf ${PACKAGE_NAME}.tar.gz $PACKAGE_NAME
  rm -rf $PACKAGE_NAME
done

cd ..

echo "Build complete! Check the $BUILD_DIR directory for outputs."
