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

4. **Provide Database Credentials**:
   Copy `containers/env` to your systemd config directory:
   ```bash
   cp containers/env ~/.config/containers/
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
