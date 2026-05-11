FROM debian:bookworm-slim

WORKDIR /app

# Install CA certificates for TLS (required for Let's Encrypt / autotls)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the binary from the local dist/ folder
# We assume 'make build' was run on the host first
COPY dist/salada /app/salada

# Copy static assets and templates
COPY web/assets /app/web/assets
COPY web/images /app/web/images
COPY web/templates /app/web/templates

# Ensure uploads directory exists (will be used as a mount point)
RUN mkdir -p /app/uploads

# Expose ports
EXPOSE 80
EXPOSE 443

# Set environment defaults to dev
ENV MODE=dev
ENV BIND_IP=0.0.0.0

CMD [ "/app/salada" ]
