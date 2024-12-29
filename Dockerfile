FROM golang:1.23-alpine AS builder
WORKDIR /app

# Define build arguments
ARG VERSION=v0.0.0
ARG COMMIT=unknown
ARG BUILDTIME=unknown

COPY . .

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "\
    -s -w \
    -X main.Version=${VERSION} \
    -X main.Commit=${COMMIT} \
    -X main.BuildTime=${BUILDTIME}" \
    -o iot-ephemeral-value-store-server main.go

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/iot-ephemeral-value-store-server /app/
EXPOSE 8080
ENTRYPOINT ["/app/iot-ephemeral-value-store-server", "-port", "8080", "-store","/db","-persist-values-for", "24h" ]
