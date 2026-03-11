DEFAULT := help

GO ?= go

# Output binary directory
BINDIR ?= bin

# Install location
INSTALL_DIR ?= $(HOME)/.local/bin

# Commands to build
CMDS := tutor gitmessage

.PHONY: help fmt vet test check build build-all clean install install-all
.PHONY: build-tutor build-gitmessage install-tutor install-gitmessage

help:
	@printf "%s\n" \
	"Targets:" \
	"  make fmt              - format Go code (modifies files)" \
	"  make vet              - run go vet" \
	"  make test             - run go test" \
	"  make check            - run vet + test" \
	"  make build-all        - build all commands (tutor, gitmessage)" \
	"  make build-tutor      - build tutor only" \
	"  make build-gitmessage - build gitmessage only" \
	"  make install-all      - install all commands to $(INSTALL_DIR)" \
	"  make install-tutor    - install tutor only" \
	"  make install-gitmessage - install gitmessage only" \
	"  make clean            - remove all built binaries"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: vet test

# Build all commands
build-all: build-tutor build-gitmessage

# Alias for build-all
build: build-all

# Build individual commands
build-tutor:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/tutor" ./cmd/tutor

build-gitmessage:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/gitmessage" ./cmd/gitmessage

# Install all commands
install-all: build-all
	cp -f "$(BINDIR)/tutor" "$(INSTALL_DIR)/tutor"
	cp -f "$(BINDIR)/gitmessage" "$(INSTALL_DIR)/gitmessage"

# Alias for install-all
install: install-all

# Install individual commands
install-tutor: build-tutor
	cp -f "$(BINDIR)/tutor" "$(INSTALL_DIR)/tutor"

install-gitmessage: build-gitmessage
	cp -f "$(BINDIR)/gitmessage" "$(INSTALL_DIR)/gitmessage"

# Clean all binaries
clean:
	rm -rf "$(BINDIR)"
