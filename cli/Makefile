DEFAULT := help

GO ?= go

# Output binary directory
BINDIR ?= bin

# Install location
INSTALL_DIR ?= $(HOME)/.local/bin

# Commands to build
CMDS := tutor gitmessage explain ask history v-edit

.PHONY: help fmt vet test check build build-all clean install install-all
.PHONY: build-tutor build-gitmessage build-explain build-ask build-history build-v-edit build-ctx-edit
.PHONY: install-tutor install-gitmessage install-explain install-ask
.PHONY: install-history install-v-edit install-ctx-edit

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
	"  make build-history    - build record only" \
	"  make build-v-edit     - build v-edit only" \
	"  make install-all      - install all commands to $(INSTALL_DIR)" \
	"  make install-tutor    - install tutor only" \
	"  make install-gitmessage - install gitmessage only" \
	"  make install-explain  - install explain only" \
	"  make install-ask      - install ask only" \
	"  make install-history  - install record only" \
	"  make install-v-edit   - install v-edit only" \
	"  make clean            - remove all built binaries"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: vet test

# Build all commands
build-all: build-tutor build-gitmessage build-explain build-ask build-history build-v-edit

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
	$(GO) build -o "$(BINDIR)/record" ./cmd/history

build-v-edit:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/v-edit" ./cmd/ctx-edit

build-ctx-edit: build-v-edit

# Install all commands
install-all: build-all
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/tutor"
	rm -f "$(INSTALL_DIR)/gitmessage"
	rm -f "$(INSTALL_DIR)/explain"
	rm -f "$(INSTALL_DIR)/ask"
	rm -f "$(INSTALL_DIR)/shistory"
	rm -f "$(INSTALL_DIR)/spicy-record"
	rm -f "$(INSTALL_DIR)/record"
	rm -f "$(INSTALL_DIR)/ctx-edit"
	rm -f "$(INSTALL_DIR)/v-edit"
	cp "$(BINDIR)/tutor" "$(INSTALL_DIR)/tutor"
	cp "$(BINDIR)/gitmessage" "$(INSTALL_DIR)/gitmessage"
	cp "$(BINDIR)/explain" "$(INSTALL_DIR)/explain"
	cp "$(BINDIR)/ask" "$(INSTALL_DIR)/ask"
	cp "$(BINDIR)/record" "$(INSTALL_DIR)/record"
	cp "$(BINDIR)/v-edit" "$(INSTALL_DIR)/v-edit"

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
	rm -f "$(INSTALL_DIR)/spicy-record"
	rm -f "$(INSTALL_DIR)/record"
	cp "$(BINDIR)/record" "$(INSTALL_DIR)/record"

install-v-edit: build-v-edit
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/ctx-edit"
	rm -f "$(INSTALL_DIR)/v-edit"
	cp "$(BINDIR)/v-edit" "$(INSTALL_DIR)/v-edit"

install-ctx-edit: install-v-edit

# Clean all binaries
clean:
	rm -rf "$(BINDIR)"
