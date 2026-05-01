.PHONY: help test test-verbose test-coverage test-race bench clean run dev

## help: Display this help message
help:
	@echo "Bapenda Service - Development Commands"
	@echo "======================================="
	@echo ""
	@echo "Testing:"
	@echo "  make test            - Run all tests"
	@echo "  make test-verbose    - Run tests with verbose output"
	@echo "  make test-race       - Run tests with race condition detection"
	@echo "  make test-coverage   - Run tests and generate coverage report"
	@echo "  make coverage-html   - Generate coverage HTML report"
	@echo "  make bench           - Run benchmarks"
	@echo ""
	@echo "Development:"
	@echo "  make run             - Run the application"
	@echo "  make dev             - Run with hot-reload (requires air)"
	@echo "  make build           - Build the application"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint            - Run linter (requires golangci-lint)"
	@echo "  make fmt             - Format code"
	@echo "  make clean           - Clean build artifacts and coverage files"
	@echo ""

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test -v -cover ./...

## test-race: Run tests with race condition detection
test-race:
	@echo "Running tests with race detection..."
	@go test -v -race ./...

## test-coverage: Run tests and show coverage percentage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...
	@go test -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -1

## coverage-html: Generate HTML coverage report
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./internal/utils
	@go test -bench=. -benchmem ./internal/cache

## run: Run the application
run:
	@echo "Building and running Bapenda Service..."
	@cd cmd/api && go run main.go

## dev: Run with hot-reload using air
dev:
	@echo "Running Bapenda Service with hot-reload..."
	@if command -v air &> /dev/null; then \
		air; \
	else \
		echo "air not installed. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

## build: Build the application
build:
	@echo "Building Bapenda Service..."
	@mkdir -p bin
	@cd cmd/api && go build -o ../../bin/bapenda-service main.go
	@echo "Build successful: bin/bapenda-service"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Please install it: https://golangci-lint.run/usage/install/"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@echo "Code formatted"

## clean: Clean build artifacts and coverage files
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## test-unit: Run only unit tests (excluding integration)
test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./...

## test-specific: Run tests matching a pattern (usage: make test-specific TEST=TestValidate)
test-specific:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific TEST=TestPattern"; \
		exit 1; \
	fi
	@echo "Running tests matching pattern: $(TEST)"
	@go test -v -run "$(TEST)" ./...

## test-package: Run tests for a specific package (usage: make test-package PKG=utils)
test-package:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-package PKG=utils"; \
		exit 1; \
	fi
	@echo "Running tests for package: $(PKG)"
	@go test -v ./internal/$(PKG)

## watch-tests: Run tests continuously (requires entr)
watch-tests:
	@echo "Watching for changes and running tests..."
	@if command -v entr &> /dev/null; then \
		find . -name "*.go" | entr -r make test; \
	else \
		echo "entr not installed. Please install it: http://eradman.com/entrproject/"; \
	fi

## stats: Show code statistics
stats:
	@echo "Code Statistics:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo ""
	@echo "Test Statistics:"
	@find . -name "*_test.go" -not -path "./vendor/*" | xargs wc -l | tail -1 || echo "No test files found"

.DEFAULT_GOAL := help
