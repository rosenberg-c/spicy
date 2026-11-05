GO ?= go

# Output binary directory
BINDIR ?= bin

# Install location
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: help
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
	"  make build-askwrapper - build askwrapper only" \
	"  make build-askwrapper-gio - build askwrapper with Gio UI" \
	"  make install-all      - install all commands to $(INSTALL_DIR)" \
	"  make install-tutor    - install tutor only" \
	"  make install-gitmessage - install gitmessage only" \
	"  make install-explain  - install explain only" \
	"  make install-ask      - install ask only" \
	"  make install-history  - install record only" \
	"  make install-v-edit   - install v-edit only" \
	"  make install-askwrapper - install askwrapper only" \
	"  make install-askwrapper-gio - install askwrapper with Gio UI" \
	"  make clean            - remove all built binaries"

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test ./...

.PHONY: check
check: vet test

# Build all commands
.PHONY: build-all
build-all: build-tutor build-gitmessage build-explain build-ask build-history build-v-edit build-askwrapper

# Alias for build-all
.PHONY: build
build: build-all

# Build individual commands
.PHONY: build-tutor
build-tutor:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/tutor" ./cmd/tutor

.PHONY: build-gitmessage
build-gitmessage:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/gitmessage" ./cmd/gitmessage

.PHONY: build-explain
build-explain:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/explain" ./cmd/explain

.PHONY: build-ask
build-ask:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/ask" ./cmd/ask

.PHONY: build-history
build-history:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/record" ./cmd/history

.PHONY: build-v-edit
build-v-edit:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/v-edit" ./cmd/ctx-edit

.PHONY: build-askwrapper
build-askwrapper:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/askwrapper" ./cmd/askwrapper

.PHONY: build-askwrapper-gio
build-askwrapper-gio:
	@mkdir -p "$(BINDIR)"
	$(GO) build -tags gio -o "$(BINDIR)/askwrapper" ./cmd/askwrapper

.PHONY: build-ctx-edit
build-ctx-edit: build-v-edit

# Install all commands
.PHONY: install-all
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
	rm -f "$(INSTALL_DIR)/askwrapper"
	cp "$(BINDIR)/tutor" "$(INSTALL_DIR)/tutor"
	cp "$(BINDIR)/gitmessage" "$(INSTALL_DIR)/gitmessage"
	cp "$(BINDIR)/explain" "$(INSTALL_DIR)/explain"
	cp "$(BINDIR)/ask" "$(INSTALL_DIR)/ask"
	cp "$(BINDIR)/record" "$(INSTALL_DIR)/record"
	cp "$(BINDIR)/v-edit" "$(INSTALL_DIR)/v-edit"
	cp "$(BINDIR)/askwrapper" "$(INSTALL_DIR)/askwrapper"

# Alias for install-all
.PHONY: install
install: install-all

# Install individual commands
.PHONY: install-tutor
install-tutor: build-tutor
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/tutor"
	cp "$(BINDIR)/tutor" "$(INSTALL_DIR)/tutor"

.PHONY: install-gitmessage
install-gitmessage: build-gitmessage
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/gitmessage"
	cp "$(BINDIR)/gitmessage" "$(INSTALL_DIR)/gitmessage"

.PHONY: install-explain
install-explain: build-explain
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/explain"
	cp "$(BINDIR)/explain" "$(INSTALL_DIR)/explain"

.PHONY: install-ask
install-ask: build-ask
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/ask"
	cp "$(BINDIR)/ask" "$(INSTALL_DIR)/ask"

.PHONY: install-history
install-history: build-history
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/shistory"
	rm -f "$(INSTALL_DIR)/spicy-record"
	rm -f "$(INSTALL_DIR)/record"
	cp "$(BINDIR)/record" "$(INSTALL_DIR)/record"

.PHONY: install-v-edit
install-v-edit: build-v-edit
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/ctx-edit"
	rm -f "$(INSTALL_DIR)/v-edit"
	cp "$(BINDIR)/v-edit" "$(INSTALL_DIR)/v-edit"

.PHONY: install-askwrapper
install-askwrapper: build-askwrapper
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/askwrapper"
	cp "$(BINDIR)/askwrapper" "$(INSTALL_DIR)/askwrapper"

.PHONY: install-askwrapper-gio
install-askwrapper-gio: build-askwrapper-gio
	@mkdir -p "$(INSTALL_DIR)"
	rm -f "$(INSTALL_DIR)/askwrapper"
	cp "$(BINDIR)/askwrapper" "$(INSTALL_DIR)/askwrapper"

.PHONY: install-ctx-edit
install-ctx-edit: install-v-edit

# Clean all binaries
.PHONY: clean
clean:
	rm -rf "$(BINDIR)"
