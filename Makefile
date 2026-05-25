# =============================================================================
# Root Makefile — CAPSTONE-BE Monorepo
# =============================================================================
# Targets:
#   make test            Run unit tests across all services
#   make coverage        Run tests with HTML coverage report per service
#   make load-test       10,000 req load test against LOCAL api-gateway (:8888)
#   make load-test-login Load test POST /api/v1/auth/login with real payload
#   make check-hey       Verify 'hey' load testing tool is installed
#   make install-hey     Install 'hey' load testing tool
# =============================================================================

GREEN  := \033[0;32m
YELLOW := \033[1;33m
RED    := \033[0;31m
CYAN   := \033[0;36m
NC     := \033[0m

# Local api-gateway port (docker-compose.yaml in api-gateway/)
GATEWAY_URL := http://localhost:8888

# Load test parameters (override with: make load-test N=5000 C=200)
N := 10000
C := 200

# Auto-detect hey binary (handles GOPATH installs not in shell PATH)
HEY := $(shell command -v hey 2>/dev/null || echo $(HOME)/go/bin/hey)

# All service directories (each is a separate Go module)
SERVICES := \
	api-gateway \
	bansos-service \
	bapenda-service \
	jdih-service \
	klinik-service \
	rssa-service \
	sidita-service \
	sinaker-service \
	siskaperbapo-service \
	transjatim-service \
	user-service

.DEFAULT_GOAL := help

# E2E test credentials (override: make e2e EMAIL=... PASSWORD=...)
EMAIL    := admin@example.com
PASSWORD := admin123456

# =============================================================================
# HELP
# =============================================================================

.PHONY: help
help:
	@echo "$(CYAN)CAPSTONE-BE Monorepo$(NC)"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  $(GREEN)make test$(NC)              Run unit tests across all services"
	@echo "  $(GREEN)make coverage$(NC)          Run tests + HTML coverage per service"
	@echo "  $(GREEN)make test-svc SVC=klinik-service$(NC)  Test a single service"
	@echo "  $(GREEN)make test-gateway$(NC)      Run api-gateway middleware + routing tests"
	@echo "  $(GREEN)make e2e$(NC)               Run E2E flow test against GATEWAY_URL"
	@echo ""
	@echo "$(YELLOW)Load Testing (LOCAL only — requires api-gateway running on :8888):$(NC)"
	@echo "  $(GREEN)make load-test$(NC)         GET $(GATEWAY_URL)/health  N=$(N) C=$(C)"
	@echo "  $(GREEN)make load-test-login$(NC)   POST /api/v1/auth/login   N=$(N) C=$(C)"
	@echo "  $(GREEN)make load-test-custom URL=/api/v1/... METHOD=GET$(NC)"
	@echo ""
	@echo "  Override defaults: $(YELLOW)make load-test N=5000 C=100$(NC)"
	@echo ""
	@echo "$(YELLOW)Tools:$(NC)"
	@echo "  $(GREEN)make install-hey$(NC)       Install 'hey' HTTP load tester"
	@echo "  $(GREEN)make check-hey$(NC)         Check if 'hey' is installed"

# =============================================================================
# UNIT TESTS
# =============================================================================

.PHONY: test
test:
	@echo "$(CYAN)Running unit tests across all services...$(NC)"
	@PASS=0; FAIL=0; SKIP=0; \
	for svc in $(SERVICES); do \
		if [ -d "$$svc" ]; then \
			printf "  %-26s " "$$svc"; \
			OUTPUT=$$(cd $$svc && go test -race -timeout 60s ./... 2>&1); \
			if echo "$$OUTPUT" | grep -q "^FAIL"; then \
				echo "$(RED)FAIL$(NC)"; \
				echo "$$OUTPUT" | grep -E "^(FAIL|---)" | sed 's/^/    /'; \
				FAIL=$$((FAIL+1)); \
			elif echo "$$OUTPUT" | grep -q "^ok"; then \
				echo "$(GREEN)ok$(NC)"; \
				PASS=$$((PASS+1)); \
			else \
				echo "$(YELLOW)no tests$(NC)"; \
				SKIP=$$((SKIP+1)); \
			fi; \
		fi; \
	done; \
	echo ""; \
	echo "$(CYAN)Results:$(NC) $(GREEN)$$PASS passed$(NC)  $(YELLOW)$$SKIP skipped$(NC)  $(RED)$$FAIL failed$(NC)"; \
	[ $$FAIL -eq 0 ]

.PHONY: test-svc
test-svc:
	@if [ -z "$(SVC)" ]; then echo "$(RED)Usage: make test-svc SVC=klinik-service$(NC)"; exit 1; fi
	@echo "$(CYAN)Testing $(SVC)...$(NC)"
	@cd $(SVC) && go test -v -race -timeout 60s ./...

.PHONY: test-gateway
test-gateway:
	@echo "$(CYAN)Testing api-gateway (middleware + routing)...$(NC)"
	@cd api-gateway && go test -v -race ./internal/...

.PHONY: e2e
e2e:
	@echo "$(CYAN)E2E test against $(GATEWAY_URL)...$(NC)"
	@go run scripts/e2e/main.go -url $(GATEWAY_URL) -email $(EMAIL) -password $(PASSWORD)

# =============================================================================
# COVERAGE
# =============================================================================

.PHONY: coverage
coverage:
	@echo "$(CYAN)Running tests with coverage across all services...$(NC)"
	@mkdir -p .coverage
	@for svc in $(SERVICES); do \
		if [ -d "$$svc" ]; then \
			COVFILE=".coverage/$$svc.out"; \
			HTMLFILE=".coverage/$$svc.html"; \
			OUTPUT=$$(cd $$svc && go test -race -coverprofile=../$$COVFILE -covermode=atomic ./... 2>&1); \
			if echo "$$OUTPUT" | grep -q "^FAIL"; then \
				printf "  %-26s $(RED)FAIL$(NC)\n" "$$svc"; \
			elif echo "$$OUTPUT" | grep -q "^ok"; then \
				(cd $$svc && go tool cover -html=../$$COVFILE -o ../$$HTMLFILE 2>/dev/null) || true; \
				TOTAL=$$(cd $$svc && go tool cover -func=../$$COVFILE 2>/dev/null | tail -1 | awk '{print $$3}'); \
				printf "  %-26s $(GREEN)%-8s$(NC) → %s\n" "$$svc" "$${TOTAL:-n/a}" "$$HTMLFILE"; \
			else \
				printf "  %-26s $(YELLOW)no tests$(NC)\n" "$$svc"; \
			fi; \
		fi; \
	done; \
	echo ""; \
	echo "$(CYAN)Coverage reports saved to .coverage/$(NC)"

.PHONY: coverage-svc
coverage-svc:
	@if [ -z "$(SVC)" ]; then echo "$(RED)Usage: make coverage-svc SVC=klinik-service$(NC)"; exit 1; fi
	@echo "$(CYAN)Coverage for $(SVC)...$(NC)"
	@mkdir -p .coverage
	@cd $(SVC) && go test -race -coverprofile=../.coverage/$(SVC).out -covermode=atomic ./...
	@go tool cover -html=.coverage/$(SVC).out -o .coverage/$(SVC).html
	@go tool cover -func=.coverage/$(SVC).out | tail -1
	@echo "$(GREEN)Report: .coverage/$(SVC).html$(NC)"

# =============================================================================
# LOAD TESTING  (LOCAL only — targets localhost:8888)
# =============================================================================

.PHONY: check-hey
check-hey:
	@if [ ! -x "$(HEY)" ]; then \
		echo "$(RED)'hey' not found. Run: make install-hey$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)hey found at: $(HEY)$(NC)"

.PHONY: install-hey
install-hey:
	@echo "$(CYAN)Installing hey...$(NC)"
	@go install github.com/rakyll/hey@latest
	@echo "$(GREEN)✓ hey installed to $$(go env GOPATH)/bin/hey$(NC)"

.PHONY: load-test
load-test: check-hey
	@echo "$(CYAN)Load test: GET $(GATEWAY_URL)/health$(NC)"
	@echo "$(YELLOW)Requests=$(N)  Concurrency=$(C)$(NC)"
	@echo ""
	@$(HEY) -n $(N) -c $(C) -q 0 $(GATEWAY_URL)/health

.PHONY: load-test-login
load-test-login: check-hey
	@echo "$(CYAN)Load test: POST $(GATEWAY_URL)/api/v1/auth/login$(NC)"
	@echo "$(YELLOW)Requests=$(N)  Concurrency=$(C)$(NC)"
	@echo ""
	@$(HEY) -n $(N) -c $(C) -q 0 \
		-m POST \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","password":"password123"}' \
		$(GATEWAY_URL)/api/v1/auth/login

.PHONY: load-test-users
load-test-users:
	@echo "$(CYAN)Load test: simulasi $(N) user unik → $(GATEWAY_URL)$(NC)"
	@echo "$(YELLOW)Concurrency=$(C) | tiap request pakai IP berbeda (X-Forwarded-For)$(NC)"
	@echo ""
	@go run scripts/loadtest/main.go -n $(N) -c $(C) -url $(GATEWAY_URL)/health

.PHONY: load-test-custom
load-test-custom: check-hey
	@if [ -z "$(URL)" ]; then echo "$(RED)Usage: make load-test-custom URL=/api/v1/... [METHOD=GET]$(NC)"; exit 1; fi
	@METHOD_FLAG=""; if [ -n "$(METHOD)" ] && [ "$(METHOD)" != "GET" ]; then METHOD_FLAG="-m $(METHOD)"; fi; \
	echo "$(CYAN)Load test: $${METHOD:-GET} $(GATEWAY_URL)$(URL)$(NC)"; \
	echo "$(YELLOW)Requests=$(N)  Concurrency=$(C)$(NC)"; \
	echo ""; \
	$(HEY) -n $(N) -c $(C) -q 0 $$METHOD_FLAG $(GATEWAY_URL)$(URL)

# =============================================================================
# UTILITIES
# =============================================================================

.PHONY: tidy-all
tidy-all:
	@echo "$(CYAN)Running go mod tidy across all services...$(NC)"
	@for svc in $(SERVICES); do \
		if [ -d "$$svc" ]; then \
			printf "  %-26s " "$$svc"; \
			cd $$svc && go mod tidy && cd ..; \
			echo "$(GREEN)ok$(NC)"; \
		fi; \
	done

.PHONY: vet-all
vet-all:
	@echo "$(CYAN)Running go vet across all services...$(NC)"
	@FAIL=0; \
	for svc in $(SERVICES); do \
		if [ -d "$$svc" ]; then \
			printf "  %-26s " "$$svc"; \
			if cd $$svc && go vet ./... 2>&1 && cd ..; then \
				echo "$(GREEN)ok$(NC)"; \
			else \
				echo "$(RED)FAIL$(NC)"; \
				FAIL=$$((FAIL+1)); \
			fi; \
		fi; \
	done; \
	[ $$FAIL -eq 0 ]

.PHONY: clean
clean:
	@rm -rf .coverage/
	@echo "$(GREEN)✓ Cleaned .coverage/$(NC)"
