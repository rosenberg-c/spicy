-- Configuration management for spicy.nvim
-- Follows Rule #3: Return values, not references
-- Follows Rule #7: Dependency injection for testability

local M = {}

-- Default configuration
-- Rule #17: Keep table shapes consistent
local DEFAULT_CONFIG = {
  -- CLI binary configuration
  bin = {
    ask = "ask",
    tutor = "tutor",
    explain = "explain",
    gitmessage = "gitmessage",
    ctx_edit = "v-edit",
  },

  -- Default models for each command
  models = {
    ask = "openai/gpt-5.3-codex",
    tutor = "openai/gpt-5.3-codex",
    explain = "openai/gpt-5.3-codex",
    gitmessage = "openai/gpt-5.3-codex",
    ctx_edit = "openai/gpt-5.3-codex",
  },

  -- UI configuration per command
  ui = {
    ask = {
      output = "float",
      float_opts = {
        relative = "editor",
        width = 0.8,
        height = 0.6,
        border = "rounded",
        title = " Spicy Ask ",
        title_pos = "center",
      },
      syntax = "markdown",
      show_spinner = true,
      auto_close = false,
    },

    tutor = {
      output = "buffer",
      auto_save = true,
      save_dir = vim.fn.expand("~/tutorials"),
      auto_open = true,
      markdown_preview = false,
    },

    explain = {
      output = "buffer",
      auto_save = true,
      save_dir = vim.fn.expand("~/explanations"),
      side_by_side = false,
      show_line_numbers = true,
      context_max_chars = 3000,
      context_surround_lines = 80,
    },

    gitmessage = {
      output = "float",
      auto_copy = true,
      float_opts = {
        relative = "editor",
        width = 0.6,
        height = 0.4,
        border = "rounded",
        title = " Git Commit Message ",
      },
      conventional_commits = true,
      show_diff = true,
      auto_insert = false,
    },

    ctx_edit = {
      show_spinner = true,
    },
  },

  -- Behavior
  verbose = false,
  timeout = 300000, -- 5 minutes in milliseconds

  -- History
  history = {
    enabled = true,
    max_entries = 100,
    save_to_file = true,
    file_path = vim.fn.stdpath("data") .. "/spicy_history.json",
  },

  -- Telescope integration
  telescope = {
    enabled = true,
    theme = "dropdown",
  },

  -- Statusline integration
  statusline = {
    enabled = false,
    show_running = true,
    show_last_result = false,
  },
}

-- Current configuration (private)
local current_config = nil

--- Deep copy a table
--- Rule #3: Return values, not references
--- @param tbl table The table to copy
--- @return table A deep copy of the table
local function deep_copy(tbl)
  if type(tbl) ~= "table" then
    return tbl
  end

  local copy = {}
  for k, v in pairs(tbl) do
    copy[k] = deep_copy(v)
  end

  return copy
end

--- Deep merge two tables
--- Rule #28: Avoid mutating inputs
--- @param base table Base configuration
--- @param override table Override configuration
--- @return table Merged configuration
local function deep_merge(base, override)
  local result = deep_copy(base)

  if type(override) ~= "table" then
    return override
  end

  for k, v in pairs(override) do
    if type(v) == "table" and type(result[k]) == "table" then
      result[k] = deep_merge(result[k], v)
    else
      result[k] = deep_copy(v)
    end
  end

  return result
end

--- Get a value from config using dot notation
--- @param key string Dot-separated key (e.g., "ui.ask.output")
--- @return any The configuration value or nil
function M.get(key)
  if not current_config then
    current_config = deep_copy(DEFAULT_CONFIG)
  end

  local parts = vim.split(key, ".", { plain = true })
  local value = current_config

  for _, part in ipairs(parts) do
    if type(value) ~= "table" then
      return nil
    end
    value = value[part]
  end

  return value
end

--- Set a value in config using dot notation
--- Rule #28: Avoid mutating inputs - creates new config
--- @param key string Dot-separated key (e.g., "ui.ask.output")
--- @param val any The value to set
function M.set(key, val)
  if not current_config then
    current_config = deep_copy(DEFAULT_CONFIG)
  end

  local parts = vim.split(key, ".", { plain = true })
  local tbl = current_config

  -- Navigate to parent table
  for i = 1, #parts - 1 do
    local part = parts[i]
    if type(tbl[part]) ~= "table" then
      tbl[part] = {}
    end
    tbl = tbl[part]
  end

  -- Set the final value
  tbl[parts[#parts]] = val
end

--- Get the full default configuration
--- Rule #3: Return copy, not reference
--- @return table A copy of the default configuration
function M.get_default()
  return deep_copy(DEFAULT_CONFIG)
end

--- Get the current full configuration
--- Rule #3: Return copy, not reference
--- @return table A copy of the current configuration
function M.get_all()
  if not current_config then
    current_config = deep_copy(DEFAULT_CONFIG)
  end
  return deep_copy(current_config)
end

--- Setup configuration with user options
--- Rule #28: Avoid mutating inputs
--- @param opts table|nil User configuration options
function M.setup(opts)
  opts = opts or {}
  current_config = deep_merge(DEFAULT_CONFIG, opts)
end

--- Reset configuration to defaults
function M.reset()
  current_config = deep_copy(DEFAULT_CONFIG)
end

--- Validate a configuration table
--- Rule #23: Validate decoded data
--- @param cfg table Configuration to validate
--- @return boolean, string|nil True if valid, or false with error message
function M.validate(cfg)
  if type(cfg) ~= "table" then
    return false, "configuration must be a table"
  end

  -- Validate bin paths if provided
  if cfg.bin and type(cfg.bin) ~= "table" then
    return false, "bin configuration must be a table"
  end

  -- Validate models
  if cfg.models and type(cfg.models) ~= "table" then
    return false, "models configuration must be a table"
  end

  -- Validate timeout
  if cfg.timeout and type(cfg.timeout) ~= "number" then
    return false, "timeout must be a number"
  end

  if cfg.timeout and cfg.timeout < 0 then
    return false, "timeout must be positive"
  end

  return true, nil
end

return M
