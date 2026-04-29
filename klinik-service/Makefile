# =============================================================================
# Makefile — Capstone-BE Service
# =============================================================================
# Usage:
#   make help              List all available targets
#   make build             Compile the service
#   make run               Run with docker-compose
#   make test              Run all tests
#   make test-coverage     Run tests with HTML coverage report
#   make sqlc              Regenerate sqlc code
#   make lint              Run linters
#   make fmt               Format code
#   make clean             Clean build artifacts
#
# Adjust SERVICE_NAME below if needed.
# =============================================================================

SERVICE_NAME := klinik-service
GO_VERSION := 1.26
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

# =============================================================================
# DEFAULT
# =============================================================================

.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)Capstone-BE: $(SERVICE_NAME)$(NC)"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# =============================================================================
# BUILD
# =============================================================================

.PHONY: build
build: ## Compile binary to ./bin/api
	@echo "$(GREEN)Building $(SERVICE_NAME)...$(NC)"
	@mkdir -p bin
	@go build -o bin/api ./cmd/api
	@echo "$(GREEN)✓ Built bin/api$(NC)"

.PHONY: build-prod
build-prod: ## Compile production binary with version info
	@echo "$(GREEN)Building production binary...$(NC)"
	@mkdir -p bin
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
	BUILD_TIME=$$(date -u +%Y-%m-%dT%H:%M:%SZ); \
	go build -ldflags="-X main.version=$$VERSION -X main.commit=$$COMMIT -X main.buildTime=$$BUILD_TIME" \
		-o bin/api ./cmd/api
	@echo "$(GREEN)✓ Built bin/api with build info$(NC)"

.PHONY: build-all
build-all: ## Verify all packages compile
	@echo "$(GREEN)Building all packages...$(NC)"
	@go build ./...
	@echo "$(GREEN)✓ All packages compile$(NC)"

# =============================================================================
# RUN
# =============================================================================

.PHONY: run
run: ## Run the service via docker-compose (from monorepo root)
	@echo "$(GREEN)Starting service via docker-compose...$(NC)"
	@cd .. && docker compose up -d $(SERVICE_NAME)-api $(SERVICE_NAME)-db $(SERVICE_NAME)-redis

.PHONY: run-local
run-local: build ## Run binary directly (no docker)
	@echo "$(GREEN)Running service locally...$(NC)"
	@./bin/api

.PHONY: stop
stop: ## Stop service containers
	@cd .. && docker compose stop $(SERVICE_NAME)-api $(SERVICE_NAME)-db $(SERVICE_NAME)-redis

.PHONY: logs
logs: ## Tail service logs
	@cd .. && docker compose logs -f $(SERVICE_NAME)-api

.PHONY: db-shell
db-shell: ## Open psql shell to service DB
	@cd .. && docker compose exec $(SERVICE_NAME)-db psql -U $$POSTGRES_USER -d $$POSTGRES_DB

# =============================================================================
# TEST
# =============================================================================

.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -race -timeout 60s ./...

.PHONY: test-short
test-short: ## Run only short tests (skip integration)
	@go test -short -race -timeout 30s ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)✓ Coverage report: $(COVERAGE_HTML)$(NC)"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

.PHONY: test-pkg
test-pkg: ## Test a specific package (usage: make test-pkg PKG=./internal/utils)
	@go test -v -race ./$(PKG)

.PHONY: bench
bench: ## Run benchmarks
	@go test -bench=. -benchmem -run=^$$ ./...

# =============================================================================
# SQLC
# =============================================================================

.PHONY: sqlc
sqlc: ## Regenerate sqlc code
	@echo "$(GREEN)Generating sqlc code...$(NC)"
	@sqlc generate
	@echo "$(GREEN)✓ Generated internal/db/$(NC)"

.PHONY: sqlc-vet
sqlc-vet: ## Validate sqlc queries against schema
	@sqlc vet

# =============================================================================
# CODE QUALITY
# =============================================================================

.PHONY: fmt
fmt: ## Format code with gofmt
	@echo "$(GREEN)Formatting code...$(NC)"
	@gofmt -w -s .
	@echo "$(GREEN)✓ Formatted$(NC)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "$(RED)golangci-lint not installed. Install: https://golangci-lint.run/usage/install/$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	@go mod tidy
	@echo "$(GREEN)✓ go.mod tidied$(NC)"

.PHONY: check
check: fmt vet lint test ## Run all quality checks

# =============================================================================
# DATABASE
# =============================================================================

.PHONY: db-reset
db-reset: ## DESTRUCTIVE: drop volume + restart with fresh schema
	@echo "$(RED)WARNING: This will DELETE all data in $(SERVICE_NAME) DB$(NC)"
	@read -p "Continue? (y/N) " confirm && [ "$$confirm" = "y" ]
	@cd .. && docker compose down
	@docker volume rm capstone-project_$(SERVICE_NAME)-postgres-data 2>/dev/null || true
	@docker volume rm capstone-project_$(SERVICE_NAME)-redis-data 2>/dev/null || true
	@cd .. && docker compose up -d
	@echo "$(GREEN)✓ DB reset complete$(NC)"

# =============================================================================
# UTILITIES
# =============================================================================

.PHONY: clean
clean: ## Remove build artifacts
	@rm -rf bin/ tmp/ $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "$(GREEN)✓ Cleaned$(NC)"

.PHONY: deps
deps: ## Install/update dependencies
	@go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

.PHONY: install-tools
install-tools: ## Install dev tools (sqlc, golangci-lint, air)
	@echo "$(GREEN)Installing dev tools...$(NC)"
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/cosmtrek/air@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "$(GREEN)✓ Tools installed$(NC)"

.PHONY: env-decrypt
env-decrypt: ## Decrypt .env.encrypted to .env (do not commit)
	@dotenvx decrypt
	@echo "$(YELLOW)Warning: .env is plaintext, do not commit$(NC)"

.PHONY: env-encrypt
env-encrypt: ## Encrypt .env back to .env.encrypted
	@dotenvx encrypt
	@rm -f .env
	@echo "$(GREEN)✓ Encrypted; .env removed$(NC)"

.PHONY: smoke-test
smoke-test: ## Quick smoke test (assumes service running)
	@echo "$(GREEN)Smoke testing endpoints...$(NC)"
	@PORT=$$(grep -oE '[0-9]+:8080' ../$(SERVICE_NAME)/docker-compose.yml | head -1 | cut -d: -f1); \
	echo "Testing port $$PORT"; \
	curl -fsS http://localhost:$$PORT/health > /dev/null && echo "  ✓ /health" || echo "  ✗ /health"; \
	curl -fsS http://localhost:$$PORT/ready > /dev/null && echo "  ✓ /ready" || echo "  ✗ /ready"; \
	curl -fsS http://localhost:$$PORT/api/v1/ > /dev/null && echo "  ✓ /api/v1/" || echo "  ✗ /api/v1/"

# Default target
.DEFAULT_GOAL := help
