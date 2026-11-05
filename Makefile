SPECSYNC ?= specsync
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: help
help:
	@printf "%s\n" \
	"Targets:" \
	"  make test      - run CLI + Neovim tests" \
	"  make test-cli  - run CLI tests" \
	"  make test-nvim - run Neovim plugin tests" \
	"  make install   - install CLI tools" \
	"  make install-imgwalker - install imgwalker GUI app" \
	"  make install-gio - install CLI tools + Gio askwrapper" \
	"  make sync-test-matrix - regenerate docs/TEST_MATRIX.md"

.PHONY: test
test: test-cli test-nvim

.PHONY: test-cli
test-cli:
	$(MAKE) -C cli test

.PHONY: test-nvim
test-nvim:
	$(MAKE) -C nvim test

.PHONY: install
install: install-cli

.PHONY: install-cli
install-cli:
	$(MAKE) -C cli install-all INSTALL_DIR="$(INSTALL_DIR)"

.PHONY: install-imgwalker
install-imgwalker:
	$(MAKE) -C imgwalker install INSTALL_DIR="$(INSTALL_DIR)"

.PHONY: sync-test-matrix
sync-test-matrix:
	$(SPECSYNC) -apply -config docs/test-matrix.config.json

.PHONY: install-gio
install-gio: install-cli-gio

.PHONY: install-cli-gio
install-cli-gio:
	$(MAKE) -C cli install-all INSTALL_DIR="$(INSTALL_DIR)"
	$(MAKE) -C cli install-askwrapper-gio INSTALL_DIR="$(INSTALL_DIR)"
