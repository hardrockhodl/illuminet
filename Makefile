# IllumiNET Makefile
#
# The default target builds the illuminet binary with build metadata
# injected via -ldflags. Cross-compilation is supported by overriding
# GOOS and GOARCH on the command line:
#
#   make build GOOS=linux GOARCH=arm64

SHELL := /bin/sh

MODULE      := github.com/hardrockhodl/illuminet
BINARY      := illuminet
CMD_PKG     := ./cmd/illuminet
VERSION_PKG := $(MODULE)/pkg/version

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Commit=$(COMMIT) \
	-X $(VERSION_PKG).Date=$(DATE)

GO        ?= go
GOBUILD   := $(GO) build -trimpath -ldflags '$(LDFLAGS)'
GOTEST    := $(GO) test
GOVET     := $(GO) vet
GOFMT     := gofmt

BIN_DIR := bin
BIN_OUT := $(BIN_DIR)/$(BINARY)
ifeq ($(GOOS),windows)
	BIN_OUT := $(BIN_OUT).exe
endif

.PHONY: all build test test-race lint fmt vet clean install run help

all: build

build:
	@mkdir -p $(BIN_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -o $(BIN_OUT) $(CMD_PKG)

test:
	$(GOTEST) ./...

test-race:
	$(GOTEST) -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

fmt:
	$(GOFMT) -w .

vet:
	$(GOVET) ./...

clean:
	rm -rf $(BIN_DIR) dist coverage.out coverage.html

install:
	$(GO) install -trimpath -ldflags '$(LDFLAGS)' $(CMD_PKG)

run:
	$(GO) run $(CMD_PKG)

help:
	@echo "Targets:"
	@echo "  build      Build the illuminet binary into $(BIN_DIR)/ (default)."
	@echo "  test       Run go test ./..."
	@echo "  test-race  Run go test with -race and coverage."
	@echo "  lint       Run golangci-lint."
	@echo "  fmt        Run gofmt -w on the tree."
	@echo "  vet        Run go vet ./..."
	@echo "  clean      Remove build artifacts."
	@echo "  install    go install with ldflags-injected metadata."
	@echo "  run        Run the binary via go run."
	@echo ""
	@echo "Override GOOS and GOARCH to cross-compile."
