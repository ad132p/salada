# salada

salada is a blog I wrote for learning a very simple webapp stack with golang and Postgres. This is a new me trying to KISS.

salada blog app should run in a single VM, ideally using Podman Quadlets to manage the containers via Systemd.

**Note: salada is TLS-only by default.** All traffic is served over HTTPS on port 443. Port 80 is not used or redirected.

# Pre requisites

git
podman
golang
postgres 18
npm


```bash
# Install Podman
sudo dnf install podman -y
```

## Running the Complete App with Podman Quadlets

We use native Podman `systemd` integration (Quadlets) to run both the Golang app and the PostgreSQL database in rootless containers.

1. **Configure Unprivileged Ports** (for rootless container binding to 443):
   ```bash
   sudo sysctl net.ipv4.ip_unprivileged_port_start=443
   # To make permanent:
   echo "net.ipv4.ip_unprivileged_port_start=443" | sudo tee /etc/sysctl.d/99-unprivileged-ports.conf
   ```

2. **Build the frontend and binary, then build the container image**:
   The container image is built from the local source tree, so the frontend assets and the compiled binary must exist before running `podman build`. Run `make build` first — it compiles the JS bundle with webpack, the CSS with Tailwind, and the Go binary into `dist/salada`:
   ```bash
   # Compile frontend assets (web/assets/js/, web/assets/css/) and Go binary (dist/salada)
   make build

   # Now build the container image — COPY . . will include the built assets
   podman build -t localhost/salada .
   ```

   **Alternative — build the image directly on the remote server via SSH**:
   Instead of building locally and shipping the image, you can point your local Podman client at the remote server. Podman sends the build context (including the pre-built assets from `make build`) over SSH, and the image is built on the server — no separate push/pull needed.

   On the server, enable the Podman socket:
   ```bash
   systemctl --user enable --now podman.socket
   ```

   On your laptop, register the server as a Podman connection (once):
   ```bash
   podman system connection add salada ssh://user@YOUR_IP/run/user/1000/podman/podman.sock
   ```

   Then build and tag the image on the remote server:
   ```bash
   # Run make build locally first so web/assets/ and dist/salada are present
   make build

   # Build the image on the remote server using the local source tree as context
   podman -c salada build -t localhost/salada .
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
   mkdir -p ~/.config/containers/
   cp containers/env ~/.config/containers/
   cp -r containers/certs ~/.config/containers/
   ```
   *Note: Edit `~/.config/containers/env` to set strong, production passwords if deploying publicly!*

5. **Enable Linger for your User** (allows user services to run at boot and persist after logout):
   ```bash
   sudo loginctl enable-linger $USER
   ```

6. **Initialize the Database Schema**:
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

7. **Start the Application via Systemd**:
   App will start automatically on boot.
   ```bash
   systemctl --user start salada.service
   ```

8. **Accessing the Database interactively**:
   If you ever need to inspect your tables or manually run queries, you can drop straight into a `psql` session inside the database container:
   ```bash
   podman exec -it salada-db psql -U salada -d salada
   ```

   You may access by configuring a psql alias:
   ```bash
   alias psql='podman run --rm -it --network host -v "$PWD":/psql -w /psql postgres:alpine psql'
   ```

## Environment Modes: `MODE=dev` vs `MODE=prod`

The application behaves differently depending on the `MODE` environment variable defined in your `containers/env` file. **Note:** `MODE` must be set to exactly `"dev"` or `"prod"`; any other value will cause the server to exit with an error.

- **`MODE=dev`**: 
  - Starts the server on port `443` binding to the IP address specified in `BIND_IP`.
  - Expects locally generated self-signed certificates (`cert.pem` and `key.pem`) for TLS.
  - Enables Gin `DebugMode` for verbose logging.

- **`MODE=prod`**: 
  - Starts the server on standard HTTPS port `443`.
  - Automatically provisions and manages real TLS certificates via Let's Encrypt for `salada.dev` and `www.salada.dev` using `autocert`.
  - Caches TLS certificates in `/var/www/.cache`.
  - Disables Gin's verbose debug logging (runs in release mode).
## Development & Makefile Usage

For quick development iteration, a Makefile is provided:

```bash
# Install frontend deps
npm i

# Build frontend and compile go binary (creates dist/salada)
make build

# Build for ARM64 (e.g. targeting an ARM server)
make build-arm64

# Clean build artifacts
make clean
```

