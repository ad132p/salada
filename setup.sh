podman build -t salada .
podman run --name salada -d -p 8080:8080 localhost/salada
podman run -p 5432:5432 --name salada-db -e POSTGRES_PASSWORD=password -d postgres
psql -d "host=localhost port=5432 dbname=postgres user=postgres"
