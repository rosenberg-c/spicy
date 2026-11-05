GO ?= go
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: build
build:
	$(GO) build ./cmd/imgwalker

.PHONY: test
test:
	$(GO) test ./...

.PHONY: install
install:
	@mkdir -p "$(INSTALL_DIR)"
	$(GO) build -o "$(INSTALL_DIR)/imgwalker" ./cmd/imgwalker
