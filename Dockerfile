# Stage 1: Build binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/downloader ./cmd/downloader

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install CA certificates for secure connections (Telegram, MinIO SSL if configured)
RUN apk add --no-cache ca-certificates tzdata

# Copy built binary from builder stage
COPY --from=builder /app/downloader /app/downloader

# Default entrypoint
ENTRYPOINT ["/app/downloader"]
