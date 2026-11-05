-- Tests for spicy.utils.cli module
local cli
local config

describe("spicy.utils.cli", function()
  local home_dir
  local original_os_homedir
  local original_cwd

  local function reload_modules()
    package.loaded["spicy.config"] = nil
    package.loaded["spicy.utils.cli"] = nil

    config = require("spicy.config")
    cli = require("spicy.utils.cli")
  end

  -- vim.fn.chdir returns the previous cwd on success, so verify via getcwd.
  local function chdir(dir)
    vim.fn.chdir(dir)
    assert.equals(vim.fn.resolve(dir), vim.fn.resolve(vim.fn.getcwd()))
  end

  before_each(function()
    home_dir = vim.fn.tempname()

    vim.fn.mkdir(home_dir, "p")

    original_os_homedir = vim.loop.os_homedir
    vim.loop.os_homedir = function()
      return home_dir
    end

    -- Run with the cwd inside the temp home so local config discovery
    -- (walk-up from cwd) is deterministic and finds no .spicy.config.json.
    original_cwd = vim.fn.getcwd()
    chdir(home_dir)

    reload_modules()
  end)

  after_each(function()
    package.loaded["spicy.config"] = nil
    package.loaded["spicy.utils.cli"] = nil

    if original_os_homedir then
      vim.loop.os_homedir = original_os_homedir
    end

    if original_cwd then
      vim.fn.chdir(original_cwd)
    end

    if home_dir and vim.fn.isdirectory(home_dir) == 1 then
      vim.fn.delete(home_dir, "rf")
    end
  end)

  it("should use default_model when no command override exists", function()
    -- @req NVIM-CONFIG-005
    config.setup({
      default_model = "shared-model",
    })

    assert.same({ "-m", "shared-model", "--history" }, cli.base_args("ask"))
  end)

  it("should let opts.model override config models", function()
    -- @req NVIM-CONFIG-005
    config.setup({
      default_model = "shared-model",
      models = {
        ask = "ask-model",
      },
    })

    assert.same(
      { "-m", "opts-model", "--history" },
      cli.base_args("ask", { model = "opts-model" })
    )
  end)

  it("should let command specific config override default_model", function()
    -- @req NVIM-CONFIG-005
    config.setup({
      default_model = "shared-model",
      models = {
        ask = "ask-model",
      },
    })

    assert.same({ "-m", "ask-model", "--history" }, cli.base_args("ask"))
  end)
end)
