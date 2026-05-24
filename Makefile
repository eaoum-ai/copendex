.PHONY: help test build check tidy clean

GO ?= go
BIN ?= copendex
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/gomod

help:
	@printf "Repository development commands\n\n"
	@printf "  make help    Show this help message\n"
	@printf "  make test    Run Go tests\n"
	@printf "  make build   Build the copendex CLI\n"
	@printf "  make check   Run tests and build\n"
	@printf "  make tidy    Update Go module files\n"
	@printf "  make clean   Remove local build artifacts\n"

test:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" ./scripts/test.sh

build:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) build -o $(BIN) ./cmd/copendex

check: test build

tidy:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) mod tidy

clean:
	rm -rf "$(BIN)" .cache
