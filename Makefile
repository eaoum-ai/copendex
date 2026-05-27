.PHONY: help test build install uninstall check tidy clean

GO ?= go
BIN ?= copendex
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/gomod
GOPATH ?= $(shell $(GO) env GOPATH)
BINDIR ?= $(GOPATH)/bin

help:
	@printf "Repository development commands\n\n"
	@printf "  make help    Show this help message\n"
	@printf "  make test    Run Go tests\n"
	@printf "  make build   Build the copendex CLI\n"
	@printf "  make install Install the copendex CLI into BINDIR\n"
	@printf "  make uninstall Remove the copendex CLI from BINDIR\n"
	@printf "  make check   Run tests and build\n"
	@printf "  make tidy    Update Go module files\n"
	@printf "  make clean   Remove local build artifacts\n"
	@printf "\n"
	@printf "Variables:\n"
	@printf "  BINDIR=%s\n" "$(BINDIR)"

test:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" ./scripts/test.sh

build:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) build -o $(BIN) ./cmd/copendex

install:
	mkdir -p "$(BINDIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) build -o "$(BINDIR)/$(BIN)" ./cmd/copendex
	@printf "Installed %s\n" "$(BINDIR)/$(BIN)"

uninstall:
	rm -f "$(BINDIR)/$(BIN)"
	@printf "Removed %s\n" "$(BINDIR)/$(BIN)"

check: test build

tidy:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) mod tidy

clean:
	rm -rf "$(BIN)" .cache
