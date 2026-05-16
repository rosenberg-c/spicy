.PHONY: help test test-cli test-nvim install install-cli install-gio install-cli-gio sync-test-matrix link-agent-docs

SPECSYNC ?= specsync

help:
	@printf "%s\n" \
	"Targets:" \
	"  make test      - run CLI + Neovim tests" \
	"  make test-cli  - run CLI tests" \
	"  make test-nvim - run Neovim plugin tests" \
	"  make install   - install CLI tools" \
	"  make install-gio - install CLI tools + Gio askwrapper" \
	"  make link-agent-docs - symlink shared agent rules" \
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

link-agent-docs:
	ln -snf ../agent/AGENT.md AGENT.md
	ln -snf ../../agent/docs/RULES.md docs/RULES.md
	ln -snf ../../agent/docs/RULES_GO.md docs/RULES_GO.md
	@echo "Linked AGENT.md, docs/RULES.md, and docs/RULES_GO.md from ../agent"
