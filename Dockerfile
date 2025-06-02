# Build Stage

# --- Stage 1: Builder (Development & Build Environment) ---
# This stage contains all the tools needed to build your Go application and run development tools like 'air'.
FROM golang:1.24.2-alpine AS builder
LABEL maintainer="Nur Wachid <wachid@outlook.com>"
ENV CGO_ENABLED=1
ENV GO111MODULE=on

RUN apk add --no-cache git gcc g++
# Set the working directory inside the container
WORKDIR /app


# Copy Go Modules files
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the rest of the application source code
COPY . .

RUN go build -o /app/webapi /app/main.go
RUN ls /app -lah
# Run stage

FROM alpine AS production
RUN apk add --no-cache ca-certificates

ENV TZ=Asia/Jakarta
WORKDIR /app

# Copy the built binary
COPY --from=builder /app/webapi /app/webapi
COPY --from=builder /app/config/config.example.yaml /app/config/config.yaml
# Ensure executable permissions
RUN chmod +x /app/webapi

EXPOSE 8000

ENTRYPOINT ["/app/webapi", "server"]

# --- Stage 3: Development (Optional, for specific development image builds) ---
# This stage is primarily for local development with 'air'.
# We reference the 'builder' stage but override the CMD.
FROM builder AS development

# Install 'air' for development (only needed in the builder stage)
RUN go install github.com/air-verse/air@latest
COPY --from=builder /app/config/config.example.yaml /app/config/config.yaml
# Expose the port for the Go application
EXPOSE 8000

CMD ["air", "server", ".air.toml"]

