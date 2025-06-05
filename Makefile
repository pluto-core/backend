.PHONY: fmt test build run

fmt:
	go fmt ./...

test:
	go test ./...

build:
	go build -o bin/auth-service ./cmd/auth-service
	go build -o bin/manifest-service ./cmd/manifest-service

run-auth:
	go run ./cmd/auth-service

run-manifest:
	go run ./cmd/manifest-service

manifest-db-generate:
	sqlc generate --file internal/manifest/repository/sqlc.yaml

auth-db-generate:
	sqlc generate --file internal/auth/repository/sqlc.yaml

manifest-migrate:
	go run cmd/migrate/main.go manifest

auth-migrate:
	go run cmd/migrate/main.go auth

.PHONY: generate
generate:
	go generate ./internal/manifest/api
