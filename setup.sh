# I would like a salada blog app and a fresh postgres database in a single VM,
# using podman quadlets and kubectl secrets, systemctl --user mode
# Did you get that?

# Install Podman
sudo dnf install podman -y
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Create database secret:
POSTGRES_ROOT_PASSWORD=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 13)
kubectl create secret generic \
    --from-literal=password="${POSTGRES_ROOT_PASSWORD}" \
    postgres-password-kube \
    --dry-run=client \
    -o yaml | \
    podman kube play -

# Create a registry if you dont have one
podman container run -dt -p 5000:5000 --name registry docker.io/library/registry:2
# Build and push salada image
podman build -t salada .
podman image tag localhost/salada localhost:5000/salada:latest
podman image push localhost:5000/salada:latest --tls-verify=false

echo 'net.ipv4.ip_unprivileged_port_start=443' >> /etc/sysctl.conf
# Bring salada up!
cp -r containers ~/.config/containers/
systemctl --user daemon-reload
systemctl --user start salada.service

# Both services should be available from systemd
systemctl --user status salada.service
systemctl --user status salada-db.service


# Don't install psql
alias psql="podman run --network systemd-salada -ti --rm alpine/psql"

#Database migration
psql -d "host=localhost port=5432 dbname=postgres user=postgres" < internal/db/databases.sql
psql -d "host=localhost port=5432 dbname=salada user=postgres" internal/db/tables.sql

