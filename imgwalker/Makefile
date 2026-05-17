GO ?= go
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: build test install

build:
	$(GO) build ./cmd/imgwalker

test:
	$(GO) test ./...

install:
	@mkdir -p "$(INSTALL_DIR)"
	$(GO) build -o "$(INSTALL_DIR)/imgwalker" ./cmd/imgwalker
