DEFAULT := help

GO ?= go

# Output binary directory
BINDIR ?= bin

# Install location
INSTALL_DIR ?= $(HOME)/.local/bin

# Commands to build
CMDS := tutor gitmessage explain ask history ctx-edit

.PHONY: help fmt vet test check build build-all clean install install-all
.PHONY: build-tutor build-gitmessage build-explain build-ask build-history build-ctx-edit
.PHONY: install-tutor install-gitmessage install-explain install-ask
.PHONY: install-history install-ctx-edit

help:
	@printf "%s\n" \
	"Targets:" \
	"  make fmt              - format Go code (modifies files)" \
	"  make vet              - run go vet" \
	"  make test             - run go test" \
	"  make check            - run vet + test" \
	"  make build-all        - build all commands" \
	"  make build-tutor      - build tutor only" \
	"  make build-gitmessage - build gitmessage only" \
	"  make build-explain    - build explain only" \
	"  make build-ask        - build ask only" \
	"  make build-history    - build history only" \
	"  make build-ctx-edit   - build ctx-edit only" \
	"  make install-all      - install all commands to $(INSTALL_DIR)" \
	"  make install-tutor    - install tutor only" \
	"  make install-gitmessage - install gitmessage only" \
	"  make install-explain  - install explain only" \
	"  make install-ask      - install ask only" \
	"  make install-history  - install history only" \
	"  make install-ctx-edit - install ctx-edit only" \
	"  make clean            - remove all built binaries"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: vet test

# Build all commands
build-all: build-tutor build-gitmessage build-explain build-ask build-history build-ctx-edit

# Alias for build-all
build: build-all

# Build individual commands
build-tutor:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/tutor" ./cmd/tutor

build-gitmessage:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/gitmessage" ./cmd/gitmessage

build-explain:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/explain" ./cmd/explain

build-ask:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/ask" ./cmd/ask

build-history:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/shistory" ./cmd/history

build-ctx-edit:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/ctx-edit" ./cmd/ctx-edit

# Install all commands
install-all: build-all
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/tutor"
	rm -f "$(INSTALL_DIR)/gitmessage"
	rm -f "$(INSTALL_DIR)/explain"
	rm -f "$(INSTALL_DIR)/ask"
	rm -f "$(INSTALL_DIR)/shistory"
	rm -f "$(INSTALL_DIR)/ctx-edit"
	cp "$(BINDIR)/tutor" "$(INSTALL_DIR)/tutor"
	cp "$(BINDIR)/gitmessage" "$(INSTALL_DIR)/gitmessage"
	cp "$(BINDIR)/explain" "$(INSTALL_DIR)/explain"
	cp "$(BINDIR)/ask" "$(INSTALL_DIR)/ask"
	cp "$(BINDIR)/shistory" "$(INSTALL_DIR)/shistory"
	cp "$(BINDIR)/ctx-edit" "$(INSTALL_DIR)/ctx-edit"

# Alias for install-all
install: install-all

# Install individual commands
install-tutor: build-tutor
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/tutor"
	cp "$(BINDIR)/tutor" "$(INSTALL_DIR)/tutor"

install-gitmessage: build-gitmessage
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/gitmessage"
	cp "$(BINDIR)/gitmessage" "$(INSTALL_DIR)/gitmessage"

install-explain: build-explain
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/explain"
	cp "$(BINDIR)/explain" "$(INSTALL_DIR)/explain"

install-ask: build-ask
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/ask"
	cp "$(BINDIR)/ask" "$(INSTALL_DIR)/ask"

install-history: build-history
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/shistory"
	cp "$(BINDIR)/shistory" "$(INSTALL_DIR)/shistory"

install-ctx-edit: build-ctx-edit
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/ctx-edit"
	cp "$(BINDIR)/ctx-edit" "$(INSTALL_DIR)/ctx-edit"

# Clean all binaries
clean:
	rm -rf "$(BINDIR)"
