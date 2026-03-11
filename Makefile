DEFAULT := help

GO ?= go
CMD_PKG := ./cmd/tutor

# Output binary path. Override with `make BIN=...` or `make BINDIR=...`.
BINDIR ?= bin
BIN ?= $(BINDIR)/tutor

# Install location (used by `make install` and `make binlink`).
INSTALL_DIR ?= $(HOME)/.local/bin
INSTALL_BIN ?= $(INSTALL_DIR)/tutor

.PHONY: help fmt vet test check build run clean binlink install

help:
	@printf "%s\n" \
	"Targets:" \
	"  make fmt    - format Go code (modifies files)" \
	"  make vet    - run go vet" \
	"  make test   - run go test" \
	"  make check  - run vet + test" \
	"  make build  - build ./cmd/tutor into $(BIN)" \
	"  make run    - run the tutor" \
	"  make clean  - remove built binary"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: vet test

build:
	@mkdir -p "$(dir $(BIN))"
	rm -f "$(HOME)/.local/bin/tutor"
	$(GO) build -o "$(BIN)" $(CMD_PKG)

run:
	$(GO) run $(CMD_PKG)

clean:
	rm -f "$(BIN)"

binlink: build
	ln -sf "$(abspath $(BIN))" "$(INSTALL_BIN)"

install: build
	cp -f "$(BIN)" "$(INSTALL_BIN)"

