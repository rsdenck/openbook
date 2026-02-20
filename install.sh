#!/bin/bash
set -e

# --- Configuration ---
INSTALL_DIR="/opt/openbook"
BIN_DIR="$INSTALL_DIR/bin"
LOG_DIR="$INSTALL_DIR/logs"
STORAGE_DIR="$INSTALL_DIR/storage"
USER="openbook"
GROUP="openbook"
REPO="rsdenck/openbook"

# --- Colors ---
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# --- Check Root ---
if [ "$EUID" -ne 0 ]; then
  log_error "Please run as root"
  exit 1
fi

# --- Detect Architecture ---
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$OS" != "linux" ]; then
    log_error "This script supports Linux only."
    exit 1
fi

log_info "Detected System: $OS/$ARCH"

# --- Resolve Version ---
VERSION=$1
if [ -z "$VERSION" ]; then
    log_info "Fetching latest version tag..."
    # Try to fetch latest release tag from GitHub API
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

# Fallback if API fails or no release found
if [ -z "$VERSION" ] || [ "$VERSION" == "null" ]; then
    log_info "Could not detect latest version from API. Defaulting to v1.0.0"
    VERSION="v1.0.0"
fi

log_info "Target Version: $VERSION"

# --- Download ---
FILENAME="openbook_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"
TMP_DIR=$(mktemp -d)

log_info "Downloading from $URL..."
HTTP_STATUS=$(curl -L -o "$TMP_DIR/$FILENAME" -w "%{http_code}" "$URL")

if [ "$HTTP_STATUS" -ne 200 ]; then
    log_error "Download failed with HTTP status $HTTP_STATUS"
    log_error "URL: $URL"
    log_error "Please check if the release version exists."
    exit 1
fi

# --- Verify Checksum ---
log_info "Verifying checksum..."
CHECKSUMS_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"
curl -sL -o "$TMP_DIR/checksums.txt" "$CHECKSUMS_URL"

cd "$TMP_DIR"
if [ -f "checksums.txt" ]; then
    # Filter checksums for the downloaded file only
    grep "$FILENAME" checksums.txt > checksums_target.txt
    if ! sha256sum -c checksums_target.txt --status; then
         log_error "Checksum verification failed!"
         log_error "Expected: $(cat checksums_target.txt)"
         log_error "Calculated: $(sha256sum $FILENAME)"
         exit 1
    fi
    log_info "Checksum verified."
else
    log_info "No checksums.txt found, skipping verification (NOT RECOMMENDED)."
fi

# --- Extract ---
log_info "Extracting package..."
tar -xzf "$FILENAME"

# Find extracted directory (it might be openbook_vX.Y.Z_linux_amd64 or just flat)
# The tar structure in release.yml is: openbook_${VERSION}_${os}_${arch}.tar.gz containing openbook_${VERSION}_${os}_${arch}/...
PKG_DIR_NAME="openbook_${VERSION}_${OS}_${ARCH}"

if [ -d "$PKG_DIR_NAME" ]; then
    cd "$PKG_DIR_NAME"
fi

# --- Create User ---
if ! id "$USER" &>/dev/null; then
    log_info "Creating user $USER..."
    useradd -r -s /bin/false -d "$INSTALL_DIR" "$USER"
fi

# --- Install Files ---
log_info "Installing binaries to $BIN_DIR..."
mkdir -p "$BIN_DIR" "$LOG_DIR" "$STORAGE_DIR"

# Stop services if running
systemctl stop openbook-api openbook-worker || true

cp openbook-api "$BIN_DIR/"
cp openbook-worker "$BIN_DIR/"
chmod +x "$BIN_DIR/openbook-api" "$BIN_DIR/openbook-worker"

# Set Permissions
chown -R "$USER:$GROUP" "$INSTALL_DIR"

# --- Configuration ---
if [ ! -f /etc/default/openbook ]; then
    log_info "Creating default configuration at /etc/default/openbook..."
    cat > /etc/default/openbook <<EOF
# OpenBook Configuration
APP_PORT=8080
STORAGE_PATH=$STORAGE_DIR

# Database (PostgreSQL)
# Format: postgres://user:password@host:port/dbname?sslmode=disable
DB_DSN=postgres://postgres:postgres@localhost:5432/openbook?sslmode=disable

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
EOF
    log_info "Configuration created. Please edit /etc/default/openbook with your database credentials."
fi

# Ensure config is readable by the service user
chown root:$GROUP /etc/default/openbook
chmod 640 /etc/default/openbook

# --- Systemd Setup ---
log_info "Configuring Systemd..."

# API Service
cat > /etc/systemd/system/openbook-api.service <<EOF
[Unit]
Description=OpenBook API
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
User=$USER
Group=$GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$BIN_DIR/openbook-api
Restart=always
RestartSec=5
LimitNOFILE=65535
EnvironmentFile=/etc/default/openbook

[Install]
WantedBy=multi-user.target
EOF

# Worker Service
cat > /etc/systemd/system/openbook-worker.service <<EOF
[Unit]
Description=OpenBook Worker
After=network.target postgresql.service redis.service openbook-api.service
Wants=postgresql.service redis.service

[Service]
User=$USER
Group=$GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$BIN_DIR/openbook-worker
Restart=always
RestartSec=5
LimitNOFILE=65535
EnvironmentFile=/etc/default/openbook

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable openbook-api openbook-worker
systemctl restart openbook-api openbook-worker

log_info "OpenBook $VERSION installed successfully!"
log_info "Configuration file: /etc/default/openbook"
log_info "Status:"
systemctl status openbook-api --no-pager
systemctl status openbook-worker --no-pager
