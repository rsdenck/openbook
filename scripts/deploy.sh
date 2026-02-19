#!/bin/bash
set -e

APP_DIR="/opt/openbook"
API_SERVICE="openbook-api"
WORKER_SERVICE="openbook-worker"

echo "Starting Deployment for OpenBook..."

# Ensure application directory exists
if [ ! -d "$APP_DIR" ]; then
    echo "Creating application directory: $APP_DIR"
    sudo mkdir -p $APP_DIR
    sudo chown $USER:$USER $APP_DIR
fi

# 1. Stop Services (if running) to release file locks
echo "Stopping services..."
sudo systemctl stop $API_SERVICE || true
sudo systemctl stop $WORKER_SERVICE || true

# 2. Backup existing binaries
echo "Backing up old binaries..."
timestamp=$(date +%Y%m%d%H%M%S)
if [ -f "$APP_DIR/api" ]; then
    mv $APP_DIR/api $APP_DIR/api.bak.$timestamp
fi
if [ -f "$APP_DIR/worker" ]; then
    mv $APP_DIR/worker $APP_DIR/worker.bak.$timestamp
fi

# 3. Install new binaries
echo "Installing new binaries..."
if [ -f "api_new" ]; then
    # If script runs where scp dropped files (usually home dir)
    mv api_new $APP_DIR/api
elif [ -f "$APP_DIR/api_new" ]; then
    mv $APP_DIR/api_new $APP_DIR/api
fi

if [ -f "worker_new" ]; then
    mv worker_new $APP_DIR/worker
elif [ -f "$APP_DIR/worker_new" ]; then
    mv $APP_DIR/worker_new $APP_DIR/worker
fi

# Verify binaries exist
if [ ! -f "$APP_DIR/api" ] || [ ! -f "$APP_DIR/worker" ]; then
    echo "Error: Binaries missing in $APP_DIR"
    exit 1
fi

chmod +x $APP_DIR/api
chmod +x $APP_DIR/worker

# 4. Run Migrations
# Assuming migrate tool is available or we use a go command if built into binary.
# For this setup, we assume migrations are applied via CI or manually. 
# Ideally, we should run them here.

# 5. Start Services
echo "Starting services..."
sudo systemctl start $API_SERVICE
sudo systemctl start $WORKER_SERVICE

# 6. Verify Health
echo "Verifying services health..."
sleep 5

if systemctl is-active --quiet $API_SERVICE; then
    echo "✅ $API_SERVICE is active"
else
    echo "❌ $API_SERVICE failed to start"
    sudo journalctl -u $API_SERVICE -n 20 --no-pager
    exit 1
fi

if systemctl is-active --quiet $WORKER_SERVICE; then
    echo "✅ $WORKER_SERVICE is active"
else
    echo "❌ $WORKER_SERVICE failed to start"
    sudo journalctl -u $WORKER_SERVICE -n 20 --no-pager
    exit 1
fi

echo "🚀 Deployment Completed Successfully!"
