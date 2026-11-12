# Go projects should have a Makefile for commond build-related tasks.
# See help target for a list of targets and their descriptions.
#
# Debug related targets are commented out until they can be tested.

# setup defaults
SHELL := $(shell which bash)
CWD_DIR := $(shell pwd)
GITHUB_API_URL ?= https://api.github.com
# DLV_BIN := $(shell go env GOPATH)/bin/dlv
LINT_FILES := ./...

# provide extra information when format fails
define goformat
	files="$$(go fmt ./...)"; \
	if [ -n "$${files}" ]; then \
		echo "❌ ERROR: go files are not properly formatted:"; \
		echo "$$files"; \
		echo ""; \
		echo "run the 'go fmt ./..' command or configure your editor"; \
		exit 1; \
	fi;
endef

# # install dlv if it is not already installed
# define dlv
# 	cat /proc/sys/kernel/yama/ptrace_scope | grep 0 || \
# 		echo 0 | sudo tee /proc/sys/kernel/yama/ptrace_scope; \
# 	echo "Checking if '$(DLV_BIN)' exist"; \
# 	test -f "$(DLV_BIN)" || \
# 		echo "Installing dlv..." && \
# 		go install github.com/go-delve/delve/cmd/dlv@latest && \
# 		echo "Installed dlv";
# endef

# NOTE: Targets defined with .PHONY are not files, they execute commands.

# clean up the go modules files
.PHONY: tidy
tidy:
	@echo "==> starting tidy"
	go mod tidy

# go format this project
.PHONY: format
format: 
	@echo "==> starting format"
	@$(call goformat)

# run some go test
.PHONY: test
test: 
	@echo "==> starting test"
	go test ./...

# runs linter for all files
.PHONY: lint-all
lint-all:
	@echo "==> starting lint for directory: ${LINT_FILES}"
	golangci-lint run ${LINT_FILES}

# runs linter for only files with diffs from origin/main (useful for PRs)
.PHONY: lint
lint:
	@echo "==> starting lint for changed files"
	golangci-lint run --whole-files --new-from-rev=origin/main

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  tidy       - clean up the go modules files"
	@echo "  format     - go format this project"
	@echo "  test       - run some go test"
	@echo "  lint-all   - runs linter for all files (optional pass in LINT_FILES=path_to_dir_or_file_to_check)"
	@echo "  lint       - runs linter for only files with diffs from origin/main (useful for PRs)"
	@echo "  license-categories        - regenerate npm license category data"
	@echo "  verify-license-categories - verify checked-in npm data is current"
	@echo "  test-license-categories   - run npm tests for the license-categories package"
	@echo "  help       - this help message"

.PHONY: license-categories
license-categories:
	go run ./cmd/license-categories -out packages/license-categories/src/generated.ts

.PHONY: verify-license-categories
verify-license-categories:
	@tmp=$$(mktemp); \
	trap 'rm -f "$$tmp"' 0; \
	go run ./cmd/license-categories -out "$$tmp" && \
	diff -I '^// go-spdx: ' -u packages/license-categories/src/generated.ts "$$tmp"

.PHONY: test-license-categories
test-license-categories:
	cd packages/license-categories && npm test
