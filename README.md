# salada

salada is a blog I wrote for learning a very simple webapp stack with golang and Postgres. This is a new me trying to KISS.

salada blog app should run in a single VM, ideally, but you can always decouple a database pod if you really have to.
I'm using podman quadlets and kubectl secrets that pods have access to
and under Rocky Linux you should be able to run commands such as:



systemctl --user status salada
systemctl --user status salada-db



# Pre requisites

git
ansible
Rocky Linux 9


```
# Install Podman
sudo dnf install podman -y
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl

# Create database secret:
POSTGRES_ROOT_PASSWORD=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 13)
./kubectl create secret generic \
    --from-literal=password="${POSTGRES_ROOT_PASSWORD}" \
    postgres-password-kube \
    --dry-run=client \
    -o yaml | \
    podman kube play -


If you need a dev db:

podman run -d \
  --name my-postgres \
  -e POSTGRES_PASSWORD=$your_secure_password \
  -p 5432:5432 \
  -v systemd-salada-db:/var/lib/postgresql/data:Z \
  docker.io/library/postgres:latest

# Create a registry if you dont have one
podman container run -dt -p 5000:5000 --name registry docker.io/library/registry:2
# If you have one, make sure it is up
podman start registry

# Build and push salada image
podman build -t salada .
podman image tag localhost/salada localhost:5000/salada:latest
podman image push localhost:5000/salada:latest --tls-verify=false
podman network create systemd-salada


# Create datadir for your postgres
sudo mkdir -p /data/pg/
sudo chown -R $USER: /data/

echo 'net.ipv4.ip_unprivileged_port_start=443' >> /etc/sysctl.conf
# Bring salada up!
cp -r containers/* ~/.config/containers/
systemctl --user daemon-reload
systemctl --user start salada.service


#Migrate
podman exec -i salada-db psql -d "host=localhost port=5432 dbname=postgres user=postgres" < internal/db/databases.sql


#DB Ip should be
podman inspect salada-db --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'


# Both services should be available from systemd
systemctl --user status salada.service
systemctl --user status salada-db.service

# Don't install psql
alias psql="podman run --network systemd-salada -ti --rm alpine/psql"
psql -d "host=localhost port=5432 dbname=salada user=postgres"
podman exec -ti salada-db bash

```
