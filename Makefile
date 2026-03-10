# ==============================================================================
# api-task-user — Makefile
# ==============================================================================

APP_NAME    := api-task-user
CMD_PATH    := ./cmd/api
BIN_DIR     := ./bin
BIN         := $(BIN_DIR)/$(APP_NAME)
DOCKER_COMP := docker compose

GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RESET  := \033[0m

.DEFAULT_GOAL := help

# ==============================================================================
# HELP
# ==============================================================================

.PHONY: help
help: ## Show available commands
	@echo ""
	@echo "$(CYAN)$(APP_NAME)$(RESET) — Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-22s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

.PHONY: run
run: ## Run the application
	@echo "$(CYAN)▶ Starting server...$(RESET)"
	go run $(CMD_PATH)/main.go

.PHONY: run-watch
run-watch: ## Run with hot-reload via Air (installs if missing)
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	air

.PHONY: build
build: ## Compile the binary
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD_PATH)
	@echo "$(GREEN)✓ Binary at $(BIN)$(RESET)"

.PHONY: build-linux
build-linux: ## Compile for Linux (Docker)
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN)-linux $(CMD_PATH)

.PHONY: clean
clean: ## Remove binaries and cache
	@rm -rf $(BIN_DIR) coverage.out coverage.html
	go clean -cache
	@echo "$(GREEN)✓ Clean complete$(RESET)"

# ==============================================================================
# DEPENDENCIES & TOOLS
# ==============================================================================

.PHONY: deps
deps: ## Download and tidy dependencies
	go mod download && go mod tidy
	@echo "$(GREEN)✓ Dependencies OK$(RESET)"

.PHONY: tools
tools: ## Install development tools
	@echo "$(CYAN)▶ Installing tools...$(RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	go install go.uber.org/mock/mockgen@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ Tools installed$(RESET)"

# ==============================================================================
# TESTS
# ==============================================================================

.PHONY: test
test: ## Run all tests
	@echo "$(CYAN)▶ Running tests...$(RESET)"
	go test ./... -race -timeout 30s

.PHONY: test-unit
test-unit: ## Run domain unit tests only
	go test ./internal/domain/... -race -v -timeout 30s

.PHONY: test-integration
test-integration: ## Run integration tests (requires DB running)
	go test ./internal/adapter/... -race -v -timeout 60s -tags integration

.PHONY: test-coverage
test-coverage: ## Run tests with HTML coverage report
	@echo "$(CYAN)▶ Analyzing coverage...$(RESET)"
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "$(GREEN)✓ Report at coverage.html$(RESET)"

.PHONY: test-coverage-check
test-coverage-check: test-coverage ## Enforce coverage >= 80%
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 80" | bc) -eq 1 ]; then \
		echo "$(YELLOW)⚠ Coverage below 80% ($$COVERAGE%)$(RESET)" && exit 1; \
	else \
		echo "$(GREEN)✓ Coverage OK: $$COVERAGE%$(RESET)"; \
	fi

.PHONY: mocks
mocks: ## Generate mocks with mockgen
	go generate ./...
	@echo "$(GREEN)✓ Mocks generated$(RESET)"

# ==============================================================================
# CODE QUALITY
# ==============================================================================

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Auto-fix linting issues
	golangci-lint run --fix ./...

.PHONY: fmt
fmt: ## Format code with gofmt + goimports
	gofmt -w .
	goimports -w .
	@echo "$(GREEN)✓ Code formatted$(RESET)"

.PHONY: fmt-check
fmt-check: ## Verify code is properly formatted
	@DIFF=$$(gofmt -l .); \
	if [ -n "$$DIFF" ]; then \
		echo "$(YELLOW)⚠ Unformatted files:$(RESET)"; echo "$$DIFF"; exit 1; \
	fi
	@echo "$(GREEN)✓ Format OK$(RESET)"

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: check
check: fmt-check vet lint test-coverage-check ## Run full quality pipeline
	@echo "$(GREEN)✓ All checks passed$(RESET)"

# ==============================================================================
# DATABASE
# ==============================================================================

.PHONY: db-up
db-up: ## Start PostgreSQL in Docker
	$(DOCKER_COMP) up -d postgres
	@echo "$(GREEN)✓ PostgreSQL running$(RESET)"

.PHONY: db-down
db-down: ## Stop PostgreSQL
	$(DOCKER_COMP) stop postgres

.PHONY: db-reset
db-reset: ## Drop and recreate the database
	$(DOCKER_COMP) down -v postgres && $(MAKE) db-up

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	migrate -path ./internal/adapter/outbound/persistence/postgres/migrations \
		-database "$$DATABASE_URL" up
	@echo "$(GREEN)✓ Migrations applied$(RESET)"

.PHONY: migrate-down
migrate-down: ## Revert last migration
	migrate -path ./internal/adapter/outbound/persistence/postgres/migrations \
		-database "$$DATABASE_URL" down 1

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create NAME=name)
	@[ "$(NAME)" ] || (echo "$(YELLOW)Usage: make migrate-create NAME=migration_name$(RESET)" && exit 1)
	migrate create -ext sql \
		-dir ./internal/adapter/outbound/persistence/postgres/migrations \
		-seq $(NAME)
	@echo "$(GREEN)✓ Migration '$(NAME)' created$(RESET)"

# ==============================================================================
# DOCKER
# ==============================================================================

.PHONY: docker-up
docker-up: ## Start full stack (API + DB)
	$(DOCKER_COMP) up -d
	@echo "$(GREEN)✓ Stack up$(RESET)"

.PHONY: docker-up-build
docker-up-build: ## Rebuild and start stack
	$(DOCKER_COMP) up -d --build

.PHONY: docker-down
docker-down: ## Stop and remove containers
	$(DOCKER_COMP) down

.PHONY: docker-logs
docker-logs: ## Tail API logs
	$(DOCKER_COMP) logs -f api

# ==============================================================================
# UTILITIES
# ==============================================================================

.PHONY: env
env: ## Create .env from .env.example
	@[ -f .env ] \
		&& echo "$(YELLOW)⚠ .env already exists$(RESET)" \
		|| (cp .env.example .env && echo "$(GREEN)✓ .env created$(RESET)")

.PHONY: swagger
swagger: ## Generate Swagger docs
	swag init -g $(CMD_PATH)/main.go -o ./docs/swagger
	@echo "$(GREEN)✓ Swagger at ./docs/swagger$(RESET)"

.PHONY: sec
sec: ## Run security scan with gosec
	@which gosec > /dev/null 2>&1 || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

.PHONY: ci
ci: deps fmt-check vet lint test-coverage-check ## Run full CI pipeline
	@echo "$(GREEN)✓ CI passed$(RESET)"
