-- Tests for spicy.config module
local config = require("spicy.config")

describe("spicy.config", function()
  before_each(function()
    -- Reset config before each test
    config.reset()
  end)

  describe("setup", function()
    it("should load default config", function()
      -- @req NVIM-CONFIG-001
      config.setup()
      local cfg = config.get_all()

      assert.is_not_nil(cfg)
      assert.is_not_nil(cfg.models)
      assert.is_not_nil(cfg.ui)
    end)

    it("should merge user config with defaults", function()
      config.setup({
        models = {
          ask = "custom-model",
        },
      })

      assert.equals("custom-model", config.get("models.ask"))
      -- Other defaults should still be there
      assert.is_not_nil(config.get("models.tutor"))
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
