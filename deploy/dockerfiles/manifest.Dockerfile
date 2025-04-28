# ---- Build stage ----
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates for go mod and TLS
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the manifest-service binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -a -o /out/manifest-service \
      ./cmd/manifest-service

# ---- Final stage ----
FROM scratch

# Copy the built binary
COPY --from=builder /out/manifest-service /usr/local/bin/manifest-service

# Runtime user (non-root)
USER 1001

ENTRYPOINT ["/usr/local/bin/manifest-service"]
CMD ["--config", "/etc/manifest.yaml"]
