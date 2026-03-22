.PHONY: help test test-cli test-nvim

help:
	@printf "%s\n" \
	"Targets:" \
	"  make test      - run CLI + Neovim tests" \
	"  make test-cli  - run CLI tests" \
	"  make test-nvim - run Neovim plugin tests"

test: test-cli test-nvim

test-cli:
	$(MAKE) -C cli test

test-nvim:
	$(MAKE) -C nvim test
