SPECSYNC ?= specsync
INSTALL_DIR ?= $(HOME)/.local/bin
SPICY_CONFIG_DIR ?= $(HOME)/.config/spicy

.PHONY: help
help:
	@printf "%s\n" \
	"Targets:" \
	"  make test      - run CLI + Neovim tests" \
	"  make test-cli  - run CLI tests" \
	"  make test-nvim - run Neovim plugin tests" \
	"  make install   - install CLI tools and Neovim plugin" \
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
install: install-cli install-nvim

.PHONY: install-nvim
install-nvim:
	@mkdir -p "$(SPICY_CONFIG_DIR)"
	@test ! -d "$(SPICY_CONFIG_DIR)/nvim" || test -L "$(SPICY_CONFIG_DIR)/nvim"
	ln -sfn "$(CURDIR)/nvim" "$(SPICY_CONFIG_DIR)/nvim"

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
