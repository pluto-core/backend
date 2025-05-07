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

# Build the migrate binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -a -o /out/migrate \
      ./cmd/migrate

# ---- Final stage ----
FROM scratch

# Скопировали бинарь
COPY --from=builder /out/manifest-service /usr/local/bin/manifest-service
COPY --from=builder /out/migrate          /usr/local/bin/migrate

# Добавляем монтирование конфига на ту же относительную локацию
COPY --from=builder /src/configs/manifest.yaml /configs/manifest.yaml
COPY --from=builder /src/migrations /migrations

USER 1001

ENTRYPOINT ["/usr/local/bin/manifest-service"]
# CMD можно даже убрать, т.к. приложение само открывает configs/manifest.yaml

CMD ["--config", "/etc/manifest.yaml"]
