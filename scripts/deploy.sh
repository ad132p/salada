#!/bin/bash
# Deploy script for salada - builds locally and deploys to server
# Usage: ./scripts/deploy.sh user@server.example.com

set -e

SERVER=${1:-}
if [ -z "$SERVER" ]; then
    echo "Usage: $0 user@server"
    echo "Example: $0 deploy@myserver.com"
    exit 1
fi

REMOTE_DIR="/opt/salada"

echo "=== Salada Deploy Script ==="
echo "Target server: $SERVER"

# Build locally
echo "Building locally..."
./scripts/build.sh

# Deploy
echo "Deploying to server..."
rsync -avz --delete web/assets/ "$SERVER:$REMOTE_DIR/web/assets/"
scp dist/salada "$SERVER:/tmp/salada.new"

ssh "$SERVER" "sudo mv /tmp/salada.new $REMOTE_DIR/salada && \
    sudo chmod +x $REMOTE_DIR/salada && \
    sudo systemctl restart salada"

echo "Deployment complete!"
