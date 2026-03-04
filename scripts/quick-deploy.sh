#!/bin/bash
# Quick deploy - for when you just want to push a new binary without full rebuild
# Assumes you've already run build.sh
# Usage: ./scripts/quick-deploy.sh user@server.example.com

set -e

SERVER=${1:-}
if [ -z "$SERVER" ]; then
    echo "Usage: $0 user@server"
    exit 1
fi

REMOTE_DIR="/opt/salada"

echo "Quick deploying to $SERVER..."

# Upload and atomically replace
scp dist/salada "$SERVER:/tmp/salada.new"
ssh "$SERVER" "sudo mv /tmp/salada.new $REMOTE_DIR/salada && \
    sudo chmod +x $REMOTE_DIR/salada && \
    sudo systemctl restart salada"

echo "Done! Service restarted."
