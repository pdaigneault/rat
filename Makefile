# rat — speed-reading TUI for markdown and text.
#
# Common tasks. Run `make help` for the list.

BIN_DIR     := bin
BINARY      := $(BIN_DIR)/rat
PKG         := ./...
MAIN        := .
INSTALL_DIR ?= $(shell go env GOBIN)
ifeq ($(INSTALL_DIR),)
INSTALL_DIR := $(shell go env GOPATH)/bin
endif

# Inject version info at build time (falls back gracefully outside a git repo).
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

GO      ?= go
SAMPLE  := testdata/sample.md

.DEFAULT_GOAL := build

## build: compile the binary into ./bin/rat
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BINARY) $(MAIN)

## run: build and play the sample document
.PHONY: run
run: build
	./$(BINARY) $(SAMPLE)

## test: run the unit test suite
.PHONY: test
test:
	$(GO) test $(PKG)

## cover: run tests and open a coverage summary
.PHONY: cover
cover:
	$(GO) test -coverprofile=coverage.out $(PKG)
	$(GO) tool cover -func=coverage.out | tail -n 1

## vet: run go vet
.PHONY: vet
vet:
	$(GO) vet $(PKG)

## fmt: format all Go sources in place
.PHONY: fmt
fmt:
	gofmt -w .

## fmt-check: fail if any source is not gofmt-clean
.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-clean:"; echo "$$unformatted"; exit 1; \
	fi

## tidy: sync go.mod / go.sum
.PHONY: tidy
tidy:
	$(GO) mod tidy

## check: fmt-check + vet + test (what CI should run)
.PHONY: check
check: fmt-check vet test

## install: go install into GOBIN (or GOPATH/bin)
.PHONY: install
install:
	$(GO) install -ldflags '$(LDFLAGS)' $(MAIN)

## clean: remove build artefacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR) coverage.out

## help: list available targets
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F': ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
