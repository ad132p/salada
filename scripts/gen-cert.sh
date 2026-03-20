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
    echo "Certificates generated successfully in $CERTS_DIR"
else
    echo "Certificates already exist in $CERTS_DIR. Skipping generation."
fi
