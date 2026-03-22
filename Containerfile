FROM golang:1.26-bookworm AS build

WORKDIR /app

COPY . .

EXPOSE 8080
EXPOSE 443

CMD [ "./dist/salada" ]
