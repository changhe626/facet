.PHONY: dev backend frontend test lint build clean swagger e2e

# Start full stack with Docker Compose
dev:
	docker compose up --build

# Start backend locally
backend:
	cd facet-backend && go run ./cmd/server/main.go

# Start frontend dev server
frontend:
	cd facet-front && npm run dev

# Run all tests
test: test-backend test-frontend

test-backend:
	cd facet-backend && go test ./... -count=1

test-frontend:
	cd facet-front && npm test

# E2E tests (requires Playwright)
e2e:
	cd facet-front && npx playwright test

# Lint all code
lint: lint-backend lint-frontend

lint-backend:
	cd facet-backend && go vet ./...
	cd facet-backend && $(shell go env GOPATH)/bin/golangci-lint run ./... 2>/dev/null || true

lint-frontend:
	cd facet-front && npm run lint

# Format all code
format: format-go format-frontend

format-go:
	cd facet-backend && gofmt -w . 2>/dev/null || true

format-frontend:
	cd facet-front && npm run format

format-check:
	cd facet-front && npm run format:check

# Build all
build: build-backend build-frontend

build-backend:
	cd facet-backend && go build ./...

build-frontend:
	cd facet-front && npm run build

# Docker build
docker-build:
	docker compose build

# Generate Swagger API documentation
swagger:
	cd facet-backend && $(shell go env GOPATH)/bin/swag init --parseDependency --dir ./cmd/server,./internal --output docs/swagger

# Clean
clean:
	cd facet-front && rm -rf dist node_modules
	cd facet-backend && rm -rf docs/swagger
