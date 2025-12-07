GOLANGCI_LINT_VERSION = v2.7.1

GOFILES = $(shell find . -type f -name '*.go' -not -path "./pkg/api/*" -not -path "./vendor/*" -not -path "./.gocache/*" -not -path "./.git/*")

# GOPRIVATE_PATTERN = your.domain.com/your/repo

BUF_VERSION = v1.61.0

# Debug mode is off by default
debug ?= 0
arch = amd64
ifeq ($(debug), 1)
	arch = arm64
endif

core_scale ?= 1

COMPOSE_FILES := -f compose.yaml
# ifneq ($(strip $(core_scale)),1)
# COMPOSE_FILES += -f compose.core-scaled.yaml
# endif

COMPOSE_SCALE :=
ifneq ($(strip $(core_scale)),1)
COMPOSE_SCALE := --scale core=$(core_scale)
endif

test:
	RUN_INTEGRATION_TEST=yes go test -v -timeout 5m -cover ./...

test-race:
	go test -race ./...

bench:
	go test -bench=. -benchtime=3s -benchmem -run=^-memprofile=mem.out ./...

update-deps:
	go get -v -u ./...
	go mod tidy

prepare-toolchain: prepare-proto-tools

prepare-proto-tools:
	@echo "Installing buf $(BUF_VERSION)..."
	$(call check_command_exists,buf) || go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)

	@echo "Installing protoc-gen-go..."
	$(call check_command_exists,protoc-gen-go) || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

	@echo "Installing protoc-gen-go-grpc..."
	$(call check_command_exists,protoc-gen-go-grpc) || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

	@echo "Installing protoc-gen-openapi..."
	$(call check_command_exists,protoc-gen-openapi) || go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest

proto-update:
	buf dep update

proto-lint:
	buf lint

proto-breaking:
	buf breaking --against '.git#branch=main'

proto-generate:
	buf generate

proto-format: ## Format proto files
	buf format -w

proto: proto-update proto-lint proto-generate proto-format

prepare-toolchain:
	$(call check_command_exists,docker) || (echo "Docker is missing, install it before starting to code." && exit 1)

	$(call check_command_exists,git) || (echo "git is not present on the system, install it before starting to code." && exit 1)

	$(call check_command_exists,go) || (echo "golang is not present on the system, download and install it at https://go.dev/dl" && exit 1)

	$(call check_command_exists,gitversion) || (echo "${GITVERSION_NOT_INSTALLED}" && exit 1)

	@echo "Installing gci...\n"
	$(call check_command_exists,gci) || go install github.com/daixiang0/gci@latest

	@echo "Installing gofumpt...\n"
	$(call check_command_exists,gofumpt) || go install mvdan.cc/gofumpt@latest

	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)...\n"
	$(call check_command_exists,golangci-lint) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b "$(go env GOPATH)/bin" $(GOLANGCI_LINT_VERSION)

	@echo "Installing staticcheck...\n"
	$(call check_command_exists,staticcheck) || go install honnef.co/go/tools/cmd/staticcheck@latest

	@echo "Installing govulncheck...\n"
	$(call check_command_exists,govulncheck) || go install golang.org/x/vuln/cmd/govulncheck@latest

	@echo "Installing gosec...\n"
	$(call check_command_exists,gosec) || go install github.com/securego/gosec/v2/cmd/gosec@latest

	@ifdef GOPRIVATE_PATTERN
		@echo "Checking if GOPRIVATE is set correctly and contains $(GOPRIVATE_PATTERN)\n"
		go env GOPRIVATE | grep $(GOPRIVATE_PATTERN) || (echo "GOPRIVATE does not contain $(GOPRIVATE_PATTERN), setting GOPRIVATE" && go env -w GOPRIVATE=$(GOPRIVATE_PATTERN))
	endif

	@echo "Checking if pre-commit is installed..."
	pre-commit --version || (echo "pre-commit is not installed, install it with 'pip install pre-commit'" && exit 1)

	@echo "Initializing pre-commit..."
	pre-commit validate-config || pre-commit install && pre-commit install-hooks

	@echo "Installing Atlas..."
	$(call check_command_exists,atlas) || (echo "atlas is not present on the system, install it from brew or run 'curl -sSf https://atlasgo.sh | sh' " && exit 1)

update-toolchain:
	@echo "Updating buf to latest..."
	go install github.com/bufbuild/buf/cmd/buf@latest && echo "buf version: " && buf --version

	@echo "Updating protoc-gen-go..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

	@echo "Updating protoc-gen-go-grpc..."
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

	@echo "Updating protoc-gen-openapi..."
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest

	@echo "Updating gci...\n"
	go install github.com/daixiang0/gci@latest

	@echo "Updating gofumpt...\n"
	go install mvdan.cc/gofumpt@latest

	@echo "Updating govulncheck...\n"
	go install golang.org/x/vuln/cmd/govulncheck@latest

	@echo "Updating gosec...\n"
	go install github.com/securego/gosec/v2/cmd/gosec@latest

	@echo "Updating staticcheck...\n"
	go install honnef.co/go/tools/cmd/staticcheck@latest

lint: prepare-toolchain proto-lint proto-format
	@echo "Running gci..."
	@for file in ${GOFILES}; do \
		gci write -s standard -s default -s blank -s dot -s "prefix(#PROJECT)" -s localmodule --skip-vendor --skip-generated $$file; \
	done

	@echo "\nRunning gofumpt..."
	gofumpt -l -w ${GOFILES}

	@echo "\nRunning staticcheck..."
	staticcheck ./...

	@echo "\nRunning golangci-lint $(GOLANGCI_LINT_VERSION)..."
	golangci-lint run -v --fix ./...

vet:
	@echo "Running go vet..."

	$(call check_command_exists,shadow) || go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest

	@for file in ${GOFILES}; do \
		go vet -vettool=$(shell which shadow) $$file; \
	done

sec:
	@echo "Running govulncheck..."
	govulncheck ./...

	@echo "\nRunning gosec..."
	gosec -exclude-generated ./...

# check_command_exists is a helper function that checks if a command exists.
define check_command_exists
@which $(1) > /dev/null 2>&1 || (echo "$(1) command not found" && exit 1)
endef

ifeq ($(call check_command_exists,$(1)),false)
  $(error "$(1) command not found")
endif

# help prints a list of available targets and their descriptions.
help:
	@echo "Available targets:"
	@echo
	@echo "Development commands:"
	@echo "  prepare-toolchain\t\tInstall and configure all required development tools"
	@echo "  update-toolchain\t\tUpdate all development tools to their latest versions"
	@echo
	@echo "Testing commands:"
	@echo "  test\t\t\t\tRun all tests in the project"
	@echo
	@echo "Code quality commands:"
	@echo "  lint\t\t\t\tRun all linters (gci, gofumpt, staticcheck, golangci-lint)"
	@echo "  vet\t\t\t\tRun go vet and shadow analysis"
	@echo "  sec\t\t\t\tRun security analysis (govulncheck, gosec)"
	@echo
	@echo "  update-deps\t\t\tUpdate all dependencies and tidy go.mod"
	@echo
	@echo
	@echo "For more information, see the project README."
.PHONY: prepare-toolchain update-toolchain test bench vet update-deps lint sec help
