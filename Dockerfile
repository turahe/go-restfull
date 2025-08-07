# Multi-Stage Dockerfile for Development, Staging, and Production
# Maintainer: Nur Wachid <wachid@outlook.com>

# =============================================================================
# STAGE 1: BUILDER (Common build stage for all environments)
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
# STAGE 2: DEVELOPMENT (Hot reload with Air)
# =============================================================================
FROM golang:1.24.5-alpine AS development
LABEL maintainer="Nur Wachid <wachid@outlook.com>"
ENV CGO_ENABLED=1
ENV GO111MODULE=on
ENV TZ=Asia/Jakarta

# Install runtime dependencies
RUN apk add --no-cache git gcc g++ ca-certificates

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Install Air for hot reloading
RUN go install github.com/air-verse/air@latest

# Copy source code
COPY . .

# Create configs directory for external configuration mounting
RUN mkdir -p /configs

# Copy default configuration
COPY config/config.example.yaml config/config.yaml
COPY config/config.example.yaml /configs/default.yaml
COPY config/rbac_model.conf config/rbac_model.conf
COPY config/rbac_policy.csv config/rbac_policy.csv

# Expose port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8000/healthz || exit 1

# Start with Air for development
CMD ["air", "server", ".air.toml"]

# =============================================================================
# STAGE 3: STAGING (Testing environment)
# =============================================================================
FROM alpine:latest AS staging
LABEL maintainer="Nur Wachid <wachid@outlook.com>"
ENV TZ=Asia/Jakarta
ENV ENV=staging

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/webapi /app/webapi

# Create configs directory for external configuration mounting
RUN mkdir -p /configs

# Copy default configuration for staging
COPY config/config.example.yaml config/config.yaml
COPY config/config.example.yaml /configs/default.yaml
COPY config/rbac_model.conf config/rbac_model.conf
COPY config/rbac_policy.csv config/rbac_policy.csv

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

# =============================================================================
# STAGE 4: PRODUCTION (Optimized for production)
# =============================================================================
FROM alpine:latest AS production
LABEL maintainer="Nur Wachid <wachid@outlook.com>"
ENV TZ=Asia/Jakarta
ENV ENV=production

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/webapi /app/webapi

# Create configs directory for external configuration mounting
RUN mkdir -p /configs

# Copy default configuration for production
COPY config/config.example.yaml config/config.yaml
COPY config/config.example.yaml /configs/default.yaml
COPY config/rbac_model.conf config/rbac_model.conf
COPY config/rbac_policy.csv config/rbac_policy.csv

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

