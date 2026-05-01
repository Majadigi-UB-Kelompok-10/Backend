# =============================================================================
# SIDITA SERVICE — Development Makefile
# =============================================================================

SERVICE_NAME := sidita-service
DOCKER_COMPOSE := docker compose
SQLC := sqlc
GO := go

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# BUILD & RUN
# =============================================================================

.PHONY: build
build: ## Build binary
	$(GO) build -o ./bin/api ./cmd/api

.PHONY: run
run: ## Run service locally (gak via docker)
	$(GO) run ./cmd/api

.PHONY: tidy
tidy: ## Sync go.mod & go.sum
	$(GO) mod tidy

# =============================================================================
# DATABASE
# =============================================================================

.PHONY: sqlc
sqlc: ## Generate Go code from SQL queries
	$(SQLC) generate

.PHONY: db-shell
db-shell: ## Connect to local sidita-db Postgres
	$(DOCKER_COMPOSE) exec sidita-db psql -U $$POSTGRES_USER -d $$POSTGRES_DB

.PHONY: db-reset
db-reset: ## Wipe DB volume + reload schema (DANGER: data loss)
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d sidita-db
	@sleep 5
	@echo "DB reset done. Bring up the rest with: make up"

# =============================================================================
# DOCKER
# =============================================================================

.PHONY: up
up: ## Start all services
	$(DOCKER_COMPOSE) up --build

.PHONY: up-d
up-d: ## Start all services (detached)
	$(DOCKER_COMPOSE) up --build -d

.PHONY: down
down: ## Stop services (preserve volume)
	$(DOCKER_COMPOSE) down

.PHONY: down-v
down-v: ## Stop services + remove volumes (DANGER)
	$(DOCKER_COMPOSE) down -v

.PHONY: logs
logs: ## Tail sidita-api logs
	$(DOCKER_COMPOSE) logs -f sidita-api

.PHONY: ps
ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

# =============================================================================
# TESTING
# =============================================================================

.PHONY: test
test: ## Run all tests
	$(GO) test ./... -count=1

.PHONY: test-utils
test-utils: ## Run only utils tests (validator, slug)
	$(GO) test ./internal/utils/... -count=1 -v

.PHONY: test-cover
test-cover: ## Run tests with coverage
	$(GO) test ./... -count=1 -cover

# =============================================================================
# QUALITY
# =============================================================================

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint (install: brew install golangci-lint)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	$(GO) fmt ./...

# =============================================================================
# ENV (dotenvx)
# =============================================================================

.PHONY: env-decrypt
env-decrypt: ## Decrypt .env.encrypted ke .env (DANGER: jangan commit .env)
	dotenvx decrypt

.PHONY: env-encrypt
env-encrypt: ## Re-encrypt .env ke .env.encrypted, lalu hapus .env
	dotenvx encrypt
	rm -f .env

# =============================================================================
# SMOKE TEST
# =============================================================================

PORT_HOST ?= 3000

.PHONY: smoke
smoke: ## Smoke test endpoint dasar (asumsi port host = 3000, override dengan PORT_HOST=xxxx)
	@echo "=== /health ==="
	@curl -s -o - -w "\nHTTP: %{http_code}\n" http://localhost:$(PORT_HOST)/health
	@echo "\n=== /ready ==="
	@curl -s -o - -w "\nHTTP: %{http_code}\n" http://localhost:$(PORT_HOST)/ready
	@echo "\n=== /api/v1/public/areas ==="
	@curl -s -o - -w "\nHTTP: %{http_code}\n" http://localhost:$(PORT_HOST)/api/v1/public/areas

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf ./bin