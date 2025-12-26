# Krakenv - Environment Variable Management Tool
# "When envs get complex, release the krakenv"

.PHONY: all build test lint fmt clean install help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
BINARY_NAME=krakenv
BINARY_PATH=./cmd/krakenv

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Default target
all: lint test build

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_PATH)

## test: Run all tests
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-short: Run tests without race detector (faster)
test-short:
	$(GOTEST) -v -coverprofile=coverage.out ./...

## coverage: Show test coverage
coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linters
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOFMT) -s -w .
	$(GOCMD) mod tidy

## clean: Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/

## install: Install binary to GOPATH/bin
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

## deps: Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## check: Run all checks (lint + test)
check: lint test
	@echo "All checks passed!"

## run: Run the application
run: build
	./$(BINARY_NAME)

## help: Show this help
help:
	@echo "Krakenv - Environment Variable Management Tool"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

