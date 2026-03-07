# salada

salada is a blog I wrote for learning a very simple webapp stack with golang and Postgres. This is a new me trying to KISS.

salada blog app should run in a single VM, ideally.

# Pre requisites

git
podman
golang
postgres 18
pnpm


```
# Install Podman
sudo dnf install podman -y

If you need a dev db:

podman run -d \
  --name salada-db \
  -e POSTGRES_PASSWORD=$your_secure_password \
  -p 5432:5432 \
  -v systemd-salada-db:/var/lib/postgresql/18/data:Z \
  docker.io/library/postgres:latest

# Build  salada image
podman build -t salada .

#Migrate
podman exec -i salada-db psql -d "host=localhost port=5432 dbname=postgres user=postgres" < internal/db/schema_dump.sql


#DB Ip should be
podman inspect salada-db --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'


# Don't install psql
alias psql="podman run --network systemd-salada -ti --rm alpine/psql"
psql -d "host=localhost port=5432 dbname=salada user=postgres"
podman exec -ti salada-db bash



# Bring salada up
podman run -d -p 80:80 -p 443:443 --network host --name salada --replace salada


# You should also build frontend with pnpm

pnpm i
pnpm run build


# Automated build
```
Usage

# Build locally (creates dist/salada)
make build

# Build and deploy
make deploy SERVER=user@yourserver.com

# Deploy without rebuilding (if you already ran make build)
make quick-deploy SERVER=user@yourserver.com

# Or use scripts directly:
./scripts/deploy.sh user@yourserver.com
```
