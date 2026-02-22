FROM golang:1.26-alpine AS builder
WORKDIR /app

# Define build arguments
ARG VERSION=v0.0.0
ARG COMMIT=unknown
ARG BUILDTIME=unknown

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "\
    -s -w \
    -X main.Version=${VERSION} \
    -X main.Commit=${COMMIT} \
    -X main.BuildTime=${BUILDTIME}" \
    -o iot-ephemeral-value-store-server main.go

RUN mkdir /db

FROM gcr.io/distroless/static-debian13

LABEL org.opencontainers.image.source="https://github.com/dhcgn/iot-ephemeral-value-store"
LABEL org.opencontainers.image.description="IoT Ephemeral Value Store - a simple HTTP interface for storing and retrieving temporary IoT data"
LABEL org.opencontainers.image.licenses="MIT"

USER nonroot:nonroot
WORKDIR /app
COPY --from=builder /app/iot-ephemeral-value-store-server /app/
COPY --from=builder --chown=nonroot:nonroot /db /db
VOLUME ["/db"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/iot-ephemeral-value-store-server", "-healthcheck"]
ENTRYPOINT ["/app/iot-ephemeral-value-store-server", "-port", "8080", "-store","/db","-persist-values-for", "24h" ]
