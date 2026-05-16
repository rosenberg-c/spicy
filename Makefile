.PHONY: help test test-cli test-nvim install install-cli sync-test-matrix

SPECSYNC ?= specsync

help:
	@printf "%s\n" \
	"Targets:" \
	"  make test      - run CLI + Neovim tests" \
	"  make test-cli  - run CLI tests" \
	"  make test-nvim - run Neovim plugin tests" \
	"  make install   - install CLI tools" \
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
