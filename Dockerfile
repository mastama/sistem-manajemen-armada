# --- builder: compile semua binary ---
FROM golang:1.24-alpine AS builder
LABEL authors="singgihpratama"

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build 4 binary terpisah
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/api             ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/mqtt-listener   ./cmd/mqtt-listener
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/geofence-worker ./cmd/geofence-worker
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/mqtt-publisher  ./cmd/mqtt-publisher

# --- runtime image ---
FROM alpine:3.20

WORKDIR /app

RUN adduser -D appuser
USER appuser

COPY --from=builder /app/bin ./bin

# default utk service api (akan di-override docker compose)
CMD ["./bin/api"]

