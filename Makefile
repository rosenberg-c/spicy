.PHONY: help test test-cli test-nvim install install-cli install-gio install-cli-gio sync-test-matrix

SPECSYNC ?= specsync

help:
	@printf "%s\n" \
	"Targets:" \
	"  make test      - run CLI + Neovim tests" \
	"  make test-cli  - run CLI tests" \
	"  make test-nvim - run Neovim plugin tests" \
	"  make install   - install CLI tools" \
	"  make install-gio - install CLI tools + Gio askwrapper" \
	"  make sync-test-matrix - regenerate docs/TEST_MATRIX.md"

test: test-cli test-nvim

test-cli:
	$(MAKE) -C cli test

test-nvim:
	$(MAKE) -C nvim test

install: install-cli

install-cli:
	$(MAKE) -C cli install-all

sync-test-matrix:
	$(SPECSYNC) -apply -config docs/test-matrix.config.json

install-gio: install-cli-gio

install-cli-gio:
	$(MAKE) -C cli install-all
	$(MAKE) -C cli install-askwrapper-gio
