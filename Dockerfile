FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=build /out/api /app/api
COPY --chown=appuser:appuser configs /app/configs

EXPOSE 8080
USER appuser

ENTRYPOINT ["/app/api"]
CMD ["serve"]

