## Build
FROM golang:1.19-alpine AS build

WORKDIR /sedb

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY app/ app/
COPY main.go .

RUN go build -o /medb

## Deploy
FROM alpine:latest

WORKDIR /app

COPY --from=build /medb ./medb

# TODO: improve this to load user defined configs.
COPY config.yaml .
COPY machines.yaml .

EXPOSE 8080

RUN chmod +x ./medb

ENTRYPOINT ["./medb"]