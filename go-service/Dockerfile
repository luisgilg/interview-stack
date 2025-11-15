# syntax=docker/dockerfile:1
FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/http

FROM alpine:3.20
RUN apk add --no-cache ca-certificates netcat-openbsd
WORKDIR /srv

COPY --from=builder /app/server ./server
COPY config.yaml ./config.yaml
COPY wait-for.sh ./wait-for.sh

ENV GO_SERVICE_PORT=8081
EXPOSE 8081

ENTRYPOINT ["/bin/sh", "-c", "./wait-for.sh ${POSTGRES_HOST:-postgres} ${POSTGRES_PORT:-5432} ./wait-for.sh ${MONGO_HOST:-mongo} ${MONGO_PORT:-27017} ./wait-for.sh ${REDIS_HOST:-redis} ${REDIS_PORT:-6379} ./server"]
