.PHONY: help test integration smoke benchmark build install uninstall check tidy clean

GO ?= go
BIN ?= cosha
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/gomod
GOPATH ?= $(shell $(GO) env GOPATH)
BINDIR ?= $(GOPATH)/bin

help:
	@printf "Repository development commands\n\n"
	@printf "  make help    Show this help message\n"
	@printf "  make test    Run Go tests\n"
	@printf "  make integration Run compiled CLI integration tests\n"
	@printf "  make smoke   Clone public Java test repos and run smoke checks\n"
	@printf "  make benchmark Clone public Java test repos and write benchmark reports\n"
	@printf "  make build   Build the cosha CLI\n"
	@printf "  make install Install the cosha CLI into BINDIR\n"
	@printf "  make uninstall Remove the cosha CLI from BINDIR\n"
	@printf "  make check   Run tests, integration tests, and build\n"
	@printf "  make tidy    Update Go module files\n"
	@printf "  make clean   Remove local build artifacts\n"
	@printf "\n"
	@printf "Variables:\n"
	@printf "  BINDIR=%s\n" "$(BINDIR)"

test:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" ./scripts/test.sh

integration:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" ./scripts/integration.sh

smoke: build
	COSHA_BIN="$(CURDIR)/$(BIN)" GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" ./scripts/test_repositories.sh smoke

benchmark: build
	COSHA_BIN="$(CURDIR)/$(BIN)" GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" ./scripts/test_repositories.sh benchmark

build:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) build -o $(BIN) ./cmd/cosha

install:
	mkdir -p "$(BINDIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) build -o "$(BINDIR)/$(BIN)" ./cmd/cosha
	@printf "Installed %s\n" "$(BINDIR)/$(BIN)"

uninstall:
	rm -f "$(BINDIR)/$(BIN)"
	@printf "Removed %s\n" "$(BINDIR)/$(BIN)"

check: test integration build

tidy:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" $(GO) mod tidy

clean:
	rm -rf "$(BIN)" .cache
