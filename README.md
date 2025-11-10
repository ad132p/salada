# salada

salada is a blog I wrote for learning a very simple webapp stack with golang and Postgres. This is a new me trying to KISS.

salada blog app should run in a single VM, ideally, but you can always decouple a database pod if you really have to.
I'm using podman quadlets and kubectl secrets that pods have access to
and under Rocky Linux you should be able to run commands such as:


# Pre requisites

git
ansible
Rocky Linux 9


```
# Install Podman
sudo dnf install podman -y

If you need a dev db:

podman run -d \
  --name my-postgres \
  -e POSTGRES_PASSWORD=$your_secure_password \
  -p 5432:5432 \
  -v systemd-salada-db:/var/lib/postgresql/data:Z \
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

```

# Bring salada up
podman run -d -p 80:80 -p 443:443 --network host --name salada --replace salada
