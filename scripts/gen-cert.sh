#!/bin/bash
set -e

CERTS_DIR="containers/certs"
mkdir -p "$CERTS_DIR"

if [ ! -f "$CERTS_DIR/server.key" ] || [ ! -f "$CERTS_DIR/server.crt" ]; then
    echo "Generating self-signed SSL certificate for the database..."
    openssl req -new -x509 -days 365 -nodes -text -out "$CERTS_DIR/server.crt" \
      -keyout "$CERTS_DIR/server.key" -subj "/CN=salada-db"
    
    # Postgres necessitates restrictive permissions on key and cert files
    chmod 600 "$CERTS_DIR/server.key"
    chmod 600 "$CERTS_DIR/server.crt"
    echo "Database certificates generated successfully in $CERTS_DIR"
else
    echo "Database certificates already exist in $CERTS_DIR. Skipping generation."
fi

# Application certificates for dev mode
if [ ! -f "cert.pem" ] || [ ! -f "key.pem" ]; then
    echo "Generating self-signed SSL certificate for the application (dev mode)..."
    openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/CN=localhost"
    echo "Application certificates generated successfully in root directory"
else
    echo "Application certificates already exist in root directory. Skipping generation."
fi
