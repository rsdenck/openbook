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
    log_info "Fetching latest version..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

if [ -z "$VERSION" ]; then
    log_error "Could not determine version to install."
    exit 1
fi

log_info "Installing OpenBook $VERSION..."

# --- Download ---
FILENAME="openbook_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"
TMP_DIR=$(mktemp -d)

log_info "Downloading from $URL..."
if ! curl -L -o "$TMP_DIR/$FILENAME" "$URL"; then
    log_error "Download failed."
    exit 1
fi

# --- Verify Checksum ---
log_info "Verifying checksum..."
CHECKSUMS_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"
curl -sL -o "$TMP_DIR/checksums.txt" "$CHECKSUMS_URL"

cd "$TMP_DIR"
if ! sha256sum -c checksums.txt --ignore-missing --status; then
    log_error "Checksum verification failed!"
    exit 1
fi
log_info "Checksum verified."

# --- Extract ---
tar -xzf "$FILENAME"

# --- Create User ---
if ! id "$USER" &>/dev/null; then
    log_info "Creating user $USER..."
    useradd -r -s /bin/false -d "$INSTALL_DIR" "$USER"
fi

# --- Install Files ---
log_info "Installing binaries..."
mkdir -p "$BIN_DIR" "$LOG_DIR" "$STORAGE_DIR"

# Stop services if running
systemctl stop openbook-api openbook-worker || true

cp openbook-api "$BIN_DIR/"
cp openbook-worker "$BIN_DIR/"
chmod +x "$BIN_DIR/openbook-api" "$BIN_DIR/openbook-worker"

# Set Permissions
chown -R "$USER:$GROUP" "$INSTALL_DIR"

# --- Systemd Setup ---
log_info "Configuring Systemd..."

# API Service
cat > /etc/systemd/system/openbook-api.service <<EOF
[Unit]
Description=OpenBook API
After=network.target

[Service]
User=$USER
Group=$GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$BIN_DIR/openbook-api
Restart=always
LimitNOFILE=65535
EnvironmentFile=-/etc/default/openbook

[Install]
WantedBy=multi-user.target
EOF

# Worker Service
cat > /etc/systemd/system/openbook-worker.service <<EOF
[Unit]
Description=OpenBook Worker
After=network.target

[Service]
User=$USER
Group=$GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$BIN_DIR/openbook-worker
Restart=always
LimitNOFILE=65535
EnvironmentFile=-/etc/default/openbook

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable openbook-api openbook-worker
systemctl start openbook-api openbook-worker

log_info "OpenBook $VERSION installed successfully!"
log_info "Status:"
systemctl status openbook-api --no-pager
systemctl status openbook-worker --no-pager
