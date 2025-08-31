# Multi-Stage Dockerfile for Production
# Maintainer: Nur Wachid <wachid@outlook.com>

# =============================================================================
# STAGE 1: BUILDER (Build stage)
# =============================================================================
FROM golang:1.24.5-alpine AS builder
LABEL maintainer="Nur Wachid <wachid@outlook.com>"
ENV CGO_ENABLED=1
ENV GO111MODULE=on

# Install build dependencies
RUN apk add --no-cache git gcc g++

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -ldflags="-w -s" -o /app/webapi /app/main.go

# =============================================================================
# STAGE 2: PRODUCTION (Optimized for production)
# =============================================================================
FROM alpine:latest AS production
LABEL maintainer="Nur Wachid <wachid@outlook.com>"
ENV TZ=Asia/Jakarta
ENV ENV=production

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/webapi /app/webapi

# Create config directory (configs will be mounted as volumes)
RUN mkdir -p /app/config /app/uploads

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8000/healthz || exit 1

# Start the application
ENTRYPOINT ["/app/webapi", "server"]

