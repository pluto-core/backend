# ---- Build stage ----
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates for go mod and TLS
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Кэшируем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Генерируем код по OpenAPI (types, сервер, spec)
#RUN go generate ./api/openapi/auth.yaml

# Сборка auth-service бинаря
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -a -o /out/auth-service ./cmd/auth-service

# Build the migrate binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -a -o /out/migrate \
      ./cmd/migrate

# ---- Final stage ----
FROM scratch

# Копируем auth-service бинарь
COPY --from=builder /out/auth-service /usr/local/bin/auth-service

# Копируем конфиг и миграции
COPY --from=builder /src/configs/auth.yaml /configs/auth.yaml
COPY --from=builder /src/migrations/auth /migrations/auth

# Некорневая учётка
USER 1001

ENTRYPOINT ["/usr/local/bin/auth-service"]
CMD ["--config", "/configs/auth.yaml"]