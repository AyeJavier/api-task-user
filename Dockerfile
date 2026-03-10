# ==============================================================================
# Dockerfile — multi-stage build for api-task-user
# ==============================================================================

# ---- Stage 1: build ----
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies layer separately
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

# ---- Stage 2: runtime ----
FROM scratch

# Copy timezone data and CA certs from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary only — no shell, no OS
COPY --from=builder /app/bin/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
