FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o /out/api ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/api /app/api
COPY --chown=nonroot:nonroot configs /app/configs

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/api"]
CMD ["serve"]

