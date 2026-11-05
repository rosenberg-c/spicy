-- Tests for spicy.config module
local fs = require("spicy.utils.fs")
local config

-- vim.fn.chdir returns the previous cwd on success, so verify via getcwd.
local function chdir(dir)
  vim.fn.chdir(dir)
  assert.equals(vim.fn.resolve(dir), vim.fn.resolve(vim.fn.getcwd()))
end

describe("spicy.config", function()
  local home_dir
  local notify_calls
  local original_notify
  local original_os_homedir
  local original_cwd

  local function reload_config()
    package.loaded["spicy.config"] = nil
    return require("spicy.config")
  end

  local function config_file_path()
    return fs.join(home_dir, ".config", "spicy", "config.json")
  end

  local function write_config_file(content)
    local path = config_file_path()

    if type(content) == "table" then
      content = vim.fn.json_encode(content)
    end

    local ok, err = fs.write_file(path, content)
    assert.is_true(ok, err)

    return path
  end

  before_each(function()
    home_dir = vim.fn.tempname()
    vim.fn.mkdir(home_dir, "p")

    -- Run with the cwd inside the temp home so local config discovery
    -- (walk-up from cwd) is deterministic and finds no .spicy.config.json
    -- unless a test creates one.
    original_cwd = vim.fn.getcwd()
    chdir(home_dir)

    notify_calls = {}
    original_notify = vim.notify
    vim.notify = function(msg, level, opts)
      table.insert(notify_calls, {
        msg = msg,
        level = level,
        opts = opts,
      })
    end

    original_os_homedir = vim.loop.os_homedir
    vim.loop.os_homedir = function()
      return home_dir
    end

    config = reload_config()
  end)

  after_each(function()
    package.loaded["spicy.config"] = nil

    if original_notify then
      vim.notify = original_notify
    end

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

  describe("setup", function()
    it("should load default config when config.json is absent", function()
      -- @req NVIM-CONFIG-001, NVIM-CONFIG-004
      config.setup()
      local cfg = config.get_all()

      assert.is_not_nil(cfg)
      assert.is_not_nil(cfg.models)
      assert.is_not_nil(cfg.ui)
      assert.equals(0, #notify_calls)
    end)

    it("should merge config.json with defaults", function()
      -- @req NVIM-CONFIG-001
      write_config_file({
        verbose = true,
        models = {
          ask = "file-model",
        },
      })

      config.setup()

      assert.is_true(config.get("verbose"))
      assert.equals("file-model", config.get("models.ask"))
      assert.is_not_nil(config.get("models.tutor"))
    end)

    it("should use config.json default_model when command model is absent", function()
      -- @req NVIM-CONFIG-001, NVIM-CONFIG-005
      write_config_file({
        default_model = "file-default",
      })

      config.setup()

      assert.equals("file-default", config.get_model("ask"))
      assert.equals("file-default", config.get_model("tutor"))
    end)

    it("should use default_model when command model is not overridden", function()
      -- @req NVIM-CONFIG-005
      config.setup({
        default_model = "shared-model",
      })

      assert.equals("shared-model", config.get_model("ask"))
      assert.equals("shared-model", config.get_model("tutor"))
    end)

    it("should prefer explicit command models over default_model", function()
      -- @req NVIM-CONFIG-005
      config.setup({
        default_model = "shared-model",
        models = {
          explain = "explain-model",
        },
      })

      assert.equals("shared-model", config.get_model("ask"))
      assert.equals("explain-model", config.get_model("explain"))
    end)

    it("should let setup options override config.json", function()
      -- @req NVIM-CONFIG-001
      write_config_file({
        verbose = true,
        models = {
          ask = "file-model",
        },
      })

      config.setup({
        verbose = false,
        models = {
          ask = "setup-model",
        },
      })

      assert.is_false(config.get("verbose"))
      assert.equals("setup-model", config.get("models.ask"))
      assert.is_not_nil(config.get("models.tutor"))
    end)

    it("should let setup default_model override config.json", function()
      -- @req NVIM-CONFIG-001, NVIM-CONFIG-005
      write_config_file({
        default_model = "file-default",
        models = {
          ask = "file-ask",
        },
      })

      config.setup({
        default_model = "setup-default",
      })

      assert.equals("setup-default", config.get("default_model"))
      assert.equals("setup-default", config.get_model("tutor"))
      assert.equals("file-ask", config.get_model("ask"))
    end)

    it("should handle nested config overrides", function()
      config.setup({
        ui = {
          ask = {
            output = "buffer",
          },
        },
      })

      assert.equals("buffer", config.get("ui.ask.output"))
      -- Other UI settings should still exist
      assert.is_not_nil(config.get("ui.tutor"))
    end)
  end)

  describe("local config", function()
    local repo_dir

    before_each(function()
      repo_dir = vim.fn.tempname()
      vim.fn.mkdir(repo_dir, "p")
    end)

    after_each(function()
      if home_dir and vim.fn.isdirectory(home_dir) == 1 then
        chdir(home_dir)
      end

      if repo_dir and vim.fn.isdirectory(repo_dir) == 1 then
        vim.fn.delete(repo_dir, "rf")
      end
    end)

    local function write_local_config(content)
      local path = fs.join(repo_dir, ".spicy.config.json")

      if type(content) == "table" then
        content = vim.fn.json_encode(content)
      end

      local ok, err = fs.write_file(path, content)
      assert.is_true(ok, err)

      return path
    end

    it("should let local default_model override global file config", function()
      -- @req NVIM-CONFIG-006
      write_config_file({
        default_model = "file-default",
      })

      chdir(repo_dir)
      write_local_config({
        default_model = "local-default",
      })

      config.setup()

      assert.equals("local-default", config.get("default_model"))
      assert.equals("local-default", config.get_model("ask"))
      assert.equals(0, #notify_calls)
    end)

    it("should override global models per key while preserving global-only keys", function()
      -- @req NVIM-CONFIG-006
      write_config_file({
        models = {
          ask = "file-ask",
          tutor = "file-tutor",
        },
      })

      chdir(repo_dir)
      write_local_config({
        models = {
          ask = "local-ask",
        },
      })

      config.setup()

      assert.equals("local-ask", config.get("models.ask"))
      assert.equals("file-tutor", config.get("models.tutor"))
      assert.equals("local-ask", config.get_model("ask"))
      assert.equals("file-tutor", config.get_model("tutor"))
    end)

    it("should apply global file config when no local config is found", function()
      -- @req NVIM-CONFIG-006
      write_config_file({
        default_model = "file-default",
        models = {
          ask = "file-ask",
        },
      })

      chdir(repo_dir)

      config.setup()

      assert.equals("file-default", config.get("default_model"))
      assert.equals("file-ask", config.get("models.ask"))
      assert.equals("file-ask", config.get_model("ask"))
      assert.equals(0, #notify_calls)
    end)

    it("should ignore invalid local JSON and keep global config", function()
      -- @req NVIM-CONFIG-006
      write_config_file({
        default_model = "file-default",
        models = {
          ask = "file-ask",
        },
      })

      chdir(repo_dir)
      write_local_config("{not valid json")

      config.setup()

      assert.equals("file-default", config.get("default_model"))
      assert.equals("file-ask", config.get("models.ask"))
      assert.equals(1, #notify_calls)
      assert.matches("%.spicy%.config%.json", notify_calls[1].msg)
    end)

    it("should ignore local config that fails validation and keep global config", function()
      -- @req NVIM-CONFIG-006
      write_config_file({
        default_model = "file-default",
      })

      chdir(repo_dir)
      write_local_config({
        default_model = false,
      })

      config.setup()

      assert.equals("file-default", config.get("default_model"))
      assert.equals(1, #notify_calls)
      assert.matches("default_model.*must be a string", notify_calls[1].msg)
    end)

    it("should discover local config by walking up from a subdirectory", function()
      -- @req NVIM-CONFIG-006
      write_config_file({
        default_model = "file-default",
      })

      write_local_config({
        default_model = "local-default",
      })

      local nested = fs.join(repo_dir, "src", "deep")
      vim.fn.mkdir(nested, "p")
      assert.equals(1, vim.fn.isdirectory(nested))
      chdir(nested)

      config.setup()

      assert.equals("local-default", config.get("default_model"))
      assert.equals("local-default", config.get_model("ask"))
    end)

    it("should prefer local models over local default_model for that command", function()
      -- @req NVIM-CONFIG-005, NVIM-CONFIG-006
      write_config_file({
        default_model = "file-default",
      })

      chdir(repo_dir)
      write_local_config({
        default_model = "local-default",
        models = {
          ask = "local-ask",
        },
      })

      config.setup()

      assert.equals("local-ask", config.get_model("ask"))
      assert.equals("local-default", config.get_model("tutor"))
    end)

    it("should let setup options override local config", function()
      -- @req NVIM-CONFIG-006
      chdir(repo_dir)
      write_local_config({
        default_model = "local-default",
        models = {
          ask = "local-ask",
        },
      })

      config.setup({
        default_model = "setup-default",
        models = {
          ask = "setup-ask",
        },
      })

      assert.equals("setup-default", config.get("default_model"))
      assert.equals("setup-ask", config.get("models.ask"))
      assert.equals("setup-ask", config.get_model("ask"))
    end)
  end)

  describe("get", function()
    it("should get top-level values", function()
      config.setup()
      local models = config.get("models")
      assert.is_table(models)
    end)

    it("should get nested values with dot notation", function()
      -- @req NVIM-CONFIG-002
      config.setup()
      local output = config.get("ui.ask.output")
      assert.equals("float", output)
    end)

    it("should return nil for non-existent keys", function()
      config.setup()
      local value = config.get("does.not.exist")
      assert.is_nil(value)
    end)

    it("should return nil for invalid paths", function()
      config.setup()
      local value = config.get("models.ask.invalid")
      assert.is_nil(value)
    end)

    it("should lazily load config.json without setup", function()
      -- @req NVIM-CONFIG-004
      write_config_file({
        verbose = true,
        models = {
          ask = "lazy-model",
        },
      })

      assert.is_true(config.get("verbose"))
      assert.equals("lazy-model", config.get("models.ask"))
      assert.equals("lazy-model", config.get_all().models.ask)
    end)
  end)

  describe("set", function()
    it("should set top-level values", function()
      config.setup()
      config.set("verbose", true)
      assert.is_true(config.get("verbose"))
    end)

    it("should set nested values with dot notation", function()
      config.setup()
      config.set("ui.ask.output", "split")
      assert.equals("split", config.get("ui.ask.output"))
    end)

    it("should create intermediate tables if needed", function()
      config.setup()
      config.set("new.nested.value", 123)
      assert.equals(123, config.get("new.nested.value"))
    end)
  end)

  describe("get_default", function()
    it("should return a copy of defaults", function()
      local defaults = config.get_default()
      assert.is_table(defaults)

      -- Modifying returned value shouldn't affect actual defaults
      defaults.models.ask = "modified"
      local fresh_defaults = config.get_default()
      assert.is_not.equals("modified", fresh_defaults.models.ask)
    end)
  end)

  describe("get_all", function()
    it("should return current config", function()
      config.setup({ verbose = true })
      local all = config.get_all()

      assert.is_true(all.verbose)
    end)

    it("should return a copy, not reference", function()
      config.setup()
      local cfg1 = config.get_all()
      cfg1.verbose = true

      local cfg2 = config.get_all()
      assert.is_not_true(cfg2.verbose)
    end)
  end)

  describe("validate", function()
    it("should accept valid config", function()
      local valid, err = config.validate({
        default_model = "test",
        models = { ask = "test" },
        timeout = 5000,
      })

      assert.is_true(valid)
      assert.is_nil(err)
    end)

    it("should reject non-table config", function()
      local valid, err = config.validate("not a table")

      assert.is_false(valid)
      assert.is_string(err)
      assert.matches("must be a table", err)
    end)

    it("should reject invalid bin type", function()
      local valid, err = config.validate({
        bin = "not a table",
      })

      assert.is_false(valid)
      assert.matches("bin.*must be a table", err)
    end)

    it("should reject invalid models type", function()
      -- @req NVIM-CONFIG-003
      local valid, err = config.validate({
        models = "not a table",
      })

      assert.is_false(valid)
      assert.matches("models.*must be a table", err)
    end)

    it("should reject invalid default_model type", function()
      -- @req NVIM-CONFIG-003, NVIM-CONFIG-005
      local valid, err = config.validate({
        default_model = false,
      })

      assert.is_false(valid)
      assert.matches("default_model.*must be a string", err)
    end)

    it("should reject non-numeric timeout", function()
      local valid, err = config.validate({
        timeout = "not a number",
      })

      assert.is_false(valid)
      assert.matches("timeout.*must be a number", err)
    end)

    it("should reject negative timeout", function()
      local valid, err = config.validate({
        timeout = -100,
      })

      assert.is_false(valid)
      assert.matches("timeout.*positive", err)
    end)

    it("should ignore invalid JSON config files safely", function()
      -- @req NVIM-CONFIG-004
      write_config_file("{not valid json")

      config.setup()

      assert.equals("openai/gpt-5.3-codex", config.get("models.ask"))
      assert.equals(1, #notify_calls)
      assert.matches("config%.json", notify_calls[1].msg)
    end)

    it("should ignore invalid config structure safely", function()
      -- @req NVIM-CONFIG-003, NVIM-CONFIG-004
      write_config_file({
        models = "not a table",
      })

      config.setup()

      assert.equals("openai/gpt-5.3-codex", config.get("models.ask"))
      assert.equals(1, #notify_calls)
      assert.matches("models.*must be a table", notify_calls[1].msg)
    end)
  end)

  describe("reset", function()
    it("should reset config to defaults", function()
      config.setup({ verbose = true })
      assert.is_true(config.get("verbose"))

      config.reset()
      assert.is_false(config.get("verbose"))
    end)
  end)
end)
