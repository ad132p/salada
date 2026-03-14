#!/bin/bash

# Get the IP from .env or default to 127.0.0.1
IP=$(grep SALADA_HOST .env | cut -d '=' -f2)
if [ -z "$IP" ]; then
    IP="127.0.0.1"
fi

echo "Generating certificate for IP: $IP"

openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 -nodes \
  -subj "/CN=$IP" \
  -addext "subjectAltName = IP:$IP, DNS:localhost, IP:127.0.0.1"

echo "Certificates generated: cert.pem, key.pem"
echo "Note: You may still need to 'Proceed' in the browser if the certificate is not manually trusted."
