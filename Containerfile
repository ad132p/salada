FROM golang:1.26-bookworm AS build

WORKDIR /app

COPY . .

EXPOSE 80
EXPOSE 443

CMD [ "./dist/salada" ]
