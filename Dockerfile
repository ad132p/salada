# stage de build
FROM golang:1.24-bookworm AS build

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go build -o salada main.go

EXPOSE 8080

CMD [ "./salada" ]
