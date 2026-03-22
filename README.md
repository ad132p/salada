# salada

salada is a blog I wrote for learning a very simple webapp stack with golang and Postgres. This is a new me trying to KISS.

salada blog app should run in a single VM, ideally using Podman Quadlets to manage the containers via Systemd.

# Pre requisites

git
podman
golang
postgres 18
pnpm


```bash
# Install Podman
sudo dnf install podman -y
```

## Running the Complete App with Podman Quadlets

We use native Podman `systemd` integration (Quadlets) to run both the Golang app and the PostgreSQL database in rootless containers.

1. **Configure Unprivileged Ports** (for rootless container binding to 80/443):
   ```bash
   sudo sysctl net.ipv4.ip_unprivileged_port_start=80
   # To make permanent:
   echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee /etc/sysctl.d/99-unprivileged-ports.conf
   ```

2. **Build the `salada` image locally**:
   ```bash
   podman build -t localhost/salada .
   ```

3. **Set up the Quadlet files**:
   Copy the provided configuration files from the `containers/systemd` folder into your user's systemd config directory:
   ```bash
   mkdir -p ~/.config/containers/systemd/
   cp containers/systemd/*.container containers/systemd/*.volume containers/systemd/*.network ~/.config/containers/systemd/
   ```

4. **Provide Database Credentials and Certificates**:
   First, ensure the SSL certificates have been generated via `./scripts/gen-cert.sh`.
   Then copy `containers/env` and the `containers/certs` directory to your systemd config path:
   ```bash
   cp containers/env ~/.config/containers/
   cp -r containers/certs ~/.config/containers/
   ```
   *Note: Edit `~/.config/containers/env` to set strong, production passwords if deploying publicly!*

5. **Initialize the Database Schema**:
   The webapp depends on the database having the correct tables.
   You must either edit `salada-db.container` to bind mount `schema_dump.sql` to `/docker-entrypoint-initdb.d/`, or manually execute it against the fresh database:
   ```bash
   # Generate systemd files
   systemctl --user daemon-reload

   # Make sure DB is running first
   systemctl --user start salada-db.service

   # Then initialize the schema (might need to wait a few seconds for DB to be ready)
   podman exec -i salada-db psql -U salada -d salada < internal/db/schema_dump.sql
   ```

6. **Start the application via Systemd**:
   App will start automatically on boot.
   ```bash
   systemctl --user start salada.service
   ```

7. **Accessing the Database interactively**:
   If you ever need to inspect your tables or manually run queries, you can drop straight into a `psql` session inside the database container:
   ```bash
   podman exec -it salada-db psql -U salada -d salada
   ```

## Environment Modes: `MODE=dev` vs `MODE=prod`

The application behaves differently depending on the `MODE` environment variable defined in your `containers/env` file:

- **`MODE=dev`**: 
  - Starts the server on port `8080` binding to the IP address specified in `BIND_IP`.
  - Expects locally generated self-signed certificates (`cert.pem` and `key.pem`) for TLS.
  - Enables Gin `DebugMode` for verbose logging.

- **`MODE=prod`** (or any value other than `dev`):
  - Starts the server on standard web ports `80` and `443`.
  - Automatically provisions and manages real TLS certificates via Let's Encrypt for `salada.dev` and `www.salada.dev` using `autocert`.
  - Caches TLS certificates in `/var/www/.cache`.
  - Disables Gin's verbose debug logging (runs in release mode).

## Development & Makefile Usage

For quick development iteration, a Makefile is provided:

```bash
# Install frontend deps
pnpm i

# Build frontend and compile go binary (creates dist/salada)
make build

# Build and deploy
make deploy SERVER=user@yourserver.com

# Deploy without rebuilding (if you already ran make build)
make quick-deploy SERVER=user@yourserver.com

# Or use scripts directly:
./scripts/deploy.sh user@yourserver.com
```
