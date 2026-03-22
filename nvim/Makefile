.PHONY: test test-file lint format check install clean

# Run all tests
test:
	@echo "Running all tests..."
	nvim --headless -u tests/minimal_init.lua \
		-c "PlenaryBustedDirectory tests/ { minimal_init = 'tests/minimal_init.lua' }" \
		-c "quit"

# Run a specific test file
# Usage: make test-file FILE=tests/config_spec.lua
test-file:
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make test-file FILE=tests/config_spec.lua"; \
		exit 1; \
	fi
	@echo "Running $(FILE)..."
	nvim --headless -u tests/minimal_init.lua \
		-c "PlenaryBustedFile $(FILE)" \
		-c "quit"

# Run tests with coverage (if luacov is installed)
test-coverage:
	@echo "Running tests with coverage..."
	nvim --headless -u tests/minimal_init.lua \
		-c "lua require('luacov')" \
		-c "PlenaryBustedDirectory tests/ { minimal_init = 'tests/minimal_init.lua' }" \
		-c "quit"
	@if command -v luacov >/dev/null 2>&1; then \
		luacov; \
		echo "Coverage report generated: luacov.report.out"; \
	else \
		echo "luacov not installed - skipping coverage report"; \
	fi

# Lint Lua files (requires luacheck)
lint:
	@if command -v luacheck >/dev/null 2>&1; then \
		echo "Running luacheck..."; \
		luacheck lua/ tests/ --globals vim; \
	else \
		echo "luacheck not installed"; \
		echo "Install: luarocks install luacheck"; \
	fi

# Format Lua files (requires stylua)
format:
	@if command -v stylua >/dev/null 2>&1; then \
		echo "Formatting with stylua..."; \
		stylua lua/ tests/; \
	else \
		echo "stylua not installed"; \
		echo "Install: cargo install stylua"; \
	fi

# Check formatting without modifying
format-check:
	@if command -v stylua >/dev/null 2>&1; then \
		echo "Checking formatting..."; \
		stylua --check lua/ tests/; \
	else \
		echo "stylua not installed"; \
	fi

# Run all checks (lint + test)
check: lint test

# Install test dependencies
install-deps:
	@echo "Installing test dependencies..."
	@if [ ! -d "$(HOME)/.local/share/nvim/site/pack/vendor/start/plenary.nvim" ]; then \
		git clone --depth=1 https://github.com/nvim-lua/plenary.nvim \
			$(HOME)/.local/share/nvim/site/pack/vendor/start/plenary.nvim; \
	fi
	@echo "Dependencies installed!"

# Clean generated files
clean:
	rm -f luacov.*.out
	rm -rf .spicy/
	find . -name "*.test" -delete

# Help
help:
	@echo "Spicy.nvim Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  test           - Run all tests"
	@echo "  test-file      - Run specific test file (make test-file FILE=...)"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Lint Lua files with luacheck"
	@echo "  format         - Format Lua files with stylua"
	@echo "  format-check   - Check formatting without modifying"
	@echo "  check          - Run lint + test"
	@echo "  install-deps   - Install test dependencies"
	@echo "  clean          - Clean generated files"
	@echo "  help           - Show this help"
