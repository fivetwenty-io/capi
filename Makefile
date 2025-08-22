# Colors for terminal output
GREEN  := \033[1;32m
YELLOW := \033[1;33m
BLUE   := \033[1;34m
CYAN   := \033[1;36m
WHITE  := \033[1;37m
RESET  := \033[0m

# Default target - show help
.DEFAULT_GOAL := help

# Variables
# Git version information
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION ?= "dev/$(GIT_BRANCH)/$(GIT_SHA)"
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_FILES := $(shell find . -name '*.go' -type f -not -path "./vendor/*")

##@ General

.PHONY: help
help: ## Display this help message
	@echo "$(BLUE)capi-client Go Library Makefile$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(CYAN)<target>$(RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Testing & Quality

.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(RESET)"
	@go test -v $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Tests complete$(RESET)"

.PHONY: test-short
test-short: ## Run tests in short mode
	@echo "$(GREEN)Running short tests...$(RESET)"
	@go test -short $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Short tests complete$(RESET)"

.PHONY: test-race
test-race: ## Run tests with race condition detection
	@echo "$(GREEN)Running tests with race detection...$(RESET)"
	@go test -race $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Race tests complete$(RESET)"

.PHONY: coverage
coverage: ## Generate test coverage report
	@echo "$(GREEN)Generating coverage report...$(RESET)"
	@go test -coverprofile=coverage.out $(shell go list ./... | grep -v vendor)
	@go tool cover -func=coverage.out
	@echo "$(GREEN)✓ Coverage report generated$(RESET)"

.PHONY: coverage-html
coverage-html: coverage ## Generate and open HTML coverage report
	@echo "$(GREEN)Opening HTML coverage report...$(RESET)"
	@go tool cover -html=coverage.out

.PHONY: test-all
test-all: test coverage ## Run all tests and generate coverage report
	@echo "$(GREEN)✓ All tests and coverage complete$(RESET)"

.PHONY: report
report: coverage-html ## Alias for coverage-html (backwards compatibility)

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Benchmarks complete$(RESET)"

##@ Code Quality

.PHONY: fmt
fmt: ## Format all Go source files
	@echo "$(GREEN)Formatting code...$(RESET)"
	@go fmt $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Code formatted$(RESET)"

.PHONY: vet
vet: ## Run go vet on all source files
	@echo "$(GREEN)Running go vet...$(RESET)"
	@go vet $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Vet analysis complete$(RESET)"

.PHONY: lint
lint: fmt vet ## Run fmt and vet

.PHONY: govulncheck
govulncheck: ## Run vulnerability check on dependencies
	@echo "$(GREEN)Checking for vulnerabilities...$(RESET)"
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing govulncheck...$(RESET)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	@govulncheck $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Vulnerability check complete$(RESET)"

.PHONY: gosec
gosec: ## Run security scanner on source code
	@echo "$(GREEN)Running security scan (gosec)...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing gosec...$(RESET)"; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	}
	@gosec -quiet -fmt text ./...
	@echo "$(GREEN)✓ Security scan complete$(RESET)"

.PHONY: staticcheck
staticcheck: ## Run staticcheck static analysis
	@echo "$(GREEN)Running staticcheck...$(RESET)"
	@command -v staticcheck >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing staticcheck...$(RESET)"; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	}
	@staticcheck $(shell go list ./... | grep -v vendor)
	@echo "$(GREEN)✓ Staticcheck analysis complete$(RESET)"

.PHONY: trivy
trivy: ## Run Trivy container and dependency scanner
	@echo "$(GREEN)Running Trivy scan...$(RESET)"
	@command -v trivy >/dev/null 2>&1 || { \
		echo "$(YELLOW)Trivy not found. Please install it:$(RESET)"; \
		echo "$(CYAN)  brew install trivy$(RESET) (macOS)"; \
		echo "$(CYAN)  apt-get install trivy$(RESET) (Debian/Ubuntu)"; \
		echo "$(CYAN)  Or visit: https://aquasecurity.github.io/trivy$(RESET)"; \
		exit 1; \
	}
	@trivy fs --scanners vuln,misconfig,secret --severity HIGH,CRITICAL --skip-dirs vendor .
	@echo "$(GREEN)✓ Trivy scan complete$(RESET)"

.PHONY: security
security: govulncheck gosec trivy ## Run all security scans (govulncheck, gosec, trivy)
	@echo "$(GREEN)✓ All security scans complete$(RESET)"

.PHONY: check
check: lint vet staticcheck test ## Run basic checks (lint, vet, staticcheck, test)
	@echo "$(GREEN)✓ Basic checks passed$(RESET)"

.PHONY: check-all
check-all: lint vet test-all ## Run all checks (lint, vet, tests with coverage)
	@echo "$(GREEN)✓ All checks passed$(RESET)"

##@ Documentation

.PHONY: docs
docs: ## Generate Go documentation
	@echo "$(GREEN)Generating documentation...$(RESET)"
	@go doc -all ./pkg/capi
	@echo "$(GREEN)✓ Documentation generated$(RESET)"

.PHONY: godoc
godoc: ## Start local godoc server
	@echo "$(GREEN)Starting godoc server on http://localhost:6060$(RESET)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(RESET)"
	@command -v godoc >/dev/null 2>&1 || { \
		echo "$(YELLOW)Installing godoc...$(RESET)"; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	}
	@godoc -http=:6060

##@ Examples

.PHONY: examples
examples: ## Build example programs
	@echo "$(GREEN)Building examples...$(RESET)"
	@if [ -d "examples" ]; then \
		for example in examples/*/; do \
			if [ -f "$$example/main.go" ]; then \
				echo "Building $$example..."; \
				go build -o "$$example/example" "$$example/main.go"; \
			fi \
		done; \
		echo "$(GREEN)✓ Examples built$(RESET)"; \
	else \
		echo "$(YELLOW)No examples directory found$(RESET)"; \
	fi

.PHONY: run-example
run-example: ## Run a specific example (use EXAMPLE=<name>)
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "$(RED)Please specify an example: make run-example EXAMPLE=<name>$(RESET)"; \
		exit 1; \
	fi
	@if [ -f "examples/$(EXAMPLE)/main.go" ]; then \
		echo "$(GREEN)Running example: $(EXAMPLE)$(RESET)"; \
		go run examples/$(EXAMPLE)/main.go; \
	else \
		echo "$(RED)Example not found: $(EXAMPLE)$(RESET)"; \
		exit 1; \
	fi

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts and test cache
	@echo "$(YELLOW)Cleaning up...$(RESET)"
	@rm -f coverage.out coverage.html test.cov capi
	@rm -rf artifacts/
	@go clean -testcache
	@if [ -d "examples" ]; then \
		find examples -name "example" -type f -delete; \
	fi
	@echo "$(GREEN)✓ Cleanup complete$(RESET)"

##@ Release

.PHONY: tag
tag: ## Create a new version tag (use VERSION=vX.Y.Z)
	@echo "$(BLUE)Creating new tag...$(RESET)"
	@echo "Checking that VERSION was defined in the calling environment"
	@test -n "$(VERSION)" || { echo "$(RED)ERROR: VERSION not set. Use: make tag VERSION=vX.Y.Z$(RESET)"; exit 1; }
	@echo "$(GREEN)Creating tag $(VERSION)...$(RESET)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "$(GREEN)✓ Tag created. Push with: git push origin $(VERSION)$(RESET)"

.PHONY: release-notes
release-notes: ## Generate release notes between tags
	@echo "$(BLUE)Generating release notes...$(RESET)"
	@if [ -z "$(FROM)" ] || [ -z "$(TO)" ]; then \
		echo "$(YELLOW)Usage: make release-notes FROM=v1.0.0 TO=v2.0.0$(RESET)"; \
		echo "$(YELLOW)Showing last 10 commits instead:$(RESET)"; \
		git log --oneline -10; \
	else \
		echo "$(GREEN)Changes from $(FROM) to $(TO):$(RESET)"; \
		git log --oneline $(FROM)..$(TO); \
	fi

.PHONY: version
version: ## Display the current version
	@echo "$(CYAN)Version: $(VERSION)$(RESET)"

##@ Dependencies

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(GREEN)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies ready$(RESET)"

.PHONY: deps-update
deps-update: ## Update all dependencies to latest versions
	@echo "$(GREEN)Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

.PHONY: deps-tidy
deps-tidy: ## Clean up go.mod and go.sum
	@echo "$(GREEN)Tidying dependencies...$(RESET)"
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(RESET)"

.PHONY: deps-graph
deps-graph: ## Show dependency graph
	@echo "$(GREEN)Generating dependency graph...$(RESET)"
	@go mod graph
	@echo "$(GREEN)✓ Dependency graph complete$(RESET)"

##@ Build

.PHONY: cli
cli: ## Build the CLI binary
	@echo "$(GREEN)Building CLI...$(RESET)"
	@go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_SHA) -X main.date=$(BUILD_TIME)" -o capi cmd/capi/main.go
	@echo "$(GREEN)✓ CLI built as 'capi'$(RESET)"

##@ Development

.PHONY: setup
setup: deps ## Initial project setup
	@echo "$(GREEN)Setting up project...$(RESET)"
	@go mod download
	@echo "$(GREEN)✓ Project setup complete$(RESET)"

.PHONY: ci
ci: lint vet test-race coverage security ## Run all CI checks
	@echo "$(GREEN)✓ All CI checks passed$(RESET)"

# Include all phony targets
.PHONY: help test test-short test-race test-all coverage coverage-html report benchmark fmt vet lint \
        govulncheck gosec staticcheck trivy security check check-all docs godoc examples run-example \
        clean tag release-notes version deps deps-update deps-tidy deps-graph cli setup ci
