-- Configuration management for spicy.nvim
-- Follows Rule #3: Return values, not references
-- Follows Rule #7: Dependency injection for testability

local M = {}
local fs = require("spicy.utils.fs")

local LOCAL_CONFIG_FILE_NAME = ".spicy.config.json"

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

  -- Default model for commands without explicit overrides
  default_model = "openai/gpt-5.3-codex",

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
local current_model_overrides = nil
local warned_messages = {}

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

local function warn_once(msg)
  if warned_messages[msg] then
    return
  end

  warned_messages[msg] = true
  vim.notify(msg, vim.log.levels.WARN, {
    title = "Spicy",
  })
end

local function is_list(value)
  if type(vim.islist) == "function" then
    return vim.islist(value)
  end

  return vim.tbl_islist(value)
end

local function decode_json(content)
  local ok, decoded = pcall(vim.fn.json_decode, content)

  if (not ok or type(decoded) ~= "table")
    and type(vim.json) == "table"
    and type(vim.json.decode) == "function"
  then
    ok, decoded = pcall(vim.json.decode, content)
  end

  if not ok or type(decoded) ~= "table" or is_list(decoded) then
    return nil, "config.json must contain a JSON object"
  end

  return decoded, nil
end

local function get_config_file_path()
  local home = vim.loop.os_homedir()

  if not home or home == "" then
    return nil
  end

  return fs.join(home, ".config", "spicy", "config.json")
end

--- Walk up from start_dir to the filesystem root and return the first
--- existing regular file named .spicy.config.json, or nil if none is found.
--- @param start_dir? string Directory to start from (default: vim.fn.getcwd())
--- @return string|nil Path to the local config file
local function find_local_config_path(start_dir)
  start_dir = start_dir or vim.fn.getcwd()

  if not start_dir or start_dir == "" then
    return nil
  end

  local dir = start_dir

  while true do
    local path = fs.join(dir, LOCAL_CONFIG_FILE_NAME)
    if fs.is_file(path) then
      return path
    end

    local parent = vim.fn.fnamemodify(dir, ":h")
    if parent == "" then
      parent = "/"
    end

    if parent == dir then
      return nil
    end

    dir = parent
  end
end

--- Read, decode, and validate a single config file.
--- Unreadable files, invalid JSON, and validation failures warn once and
--- return nil so the remaining config layers still apply.
--- @param path string Path to the config file
--- @return table|nil Decoded configuration
local function load_config_file(path)
  local content, err = fs.read_file(path)
  if not content then
    warn_once(("Ignoring unreadable config file %s: %s"):format(
      path,
      err or "unknown error"
    ))
    return nil
  end

  local decoded, decode_err = decode_json(content)
  if not decoded then
    warn_once(("Ignoring invalid config file %s: %s"):format(
      path,
      decode_err
    ))
    return nil
  end

  local valid, validation_err = M.validate(decoded)
  if not valid then
    warn_once(("Ignoring invalid config file %s: %s"):format(
      path,
      validation_err
    ))
    return nil
  end

  return decoded
end

local function load_file_config()
  local path = get_config_file_path()

  if not path or not fs.exists(path) then
    return nil
  end

  return load_config_file(path)
end

local function load_local_config()
  local path = find_local_config_path()

  if not path then
    return nil
  end

  return load_config_file(path)
end

local function build_config(opts)
  local cfg = deep_copy(DEFAULT_CONFIG)
  local file_config = load_file_config()
  local local_config = load_local_config()
  local model_overrides = {}

  local function collect_model_overrides(source)
    if type(source) ~= "table" or type(source.models) ~= "table" then
      return
    end

    for command_name, _ in pairs(source.models) do
      model_overrides[command_name] = true
    end
  end

  if file_config then
    collect_model_overrides(file_config)
    cfg = deep_merge(cfg, file_config)
  end

  if local_config then
    collect_model_overrides(local_config)
    cfg = deep_merge(cfg, local_config)
  end

  if opts ~= nil then
    collect_model_overrides(opts)
    cfg = deep_merge(cfg, opts)
  end

  return cfg, model_overrides
end

local function ensure_loaded()
  if not current_config then
    current_config, current_model_overrides = build_config()
  end

  return current_config
end

--- Get a value from config using dot notation
--- @param key string Dot-separated key (e.g., "ui.ask.output")
--- @return any The configuration value or nil
function M.get(key)
  local cfg = ensure_loaded()

  local parts = vim.split(key, ".", { plain = true })
  local value = cfg

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
  ensure_loaded()

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

  if key == "models" and type(val) == "table" then
    current_model_overrides = {}

    for command_name, _ in pairs(val) do
      current_model_overrides[command_name] = true
    end
  elseif #parts == 2 and parts[1] == "models" then
    current_model_overrides = current_model_overrides or {}
    current_model_overrides[parts[2]] = val ~= nil or nil
  end
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
  return deep_copy(ensure_loaded())
end

--- Get the effective model for a command.
--- Command-specific model entries only override default_model when they were
--- provided by a config file (local .spicy.config.json or global
--- config.json), setup(opts), or config.set.
--- @param command_name string
--- @return string|nil
function M.get_model(command_name)
  local cfg = ensure_loaded()

  if current_model_overrides and current_model_overrides[command_name] then
    return cfg.models and cfg.models[command_name] or nil
  end

  if cfg.default_model ~= nil then
    return cfg.default_model
  end

  return cfg.models and cfg.models[command_name] or nil
end

--- Setup configuration with user options
--- Rule #28: Avoid mutating inputs
--- @param opts table|nil User configuration options
function M.setup(opts)
  opts = opts or {}
  current_config, current_model_overrides = build_config(opts)
end

--- Reset configuration to defaults
function M.reset()
  current_config = nil
  current_model_overrides = nil
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

  if cfg.default_model ~= nil and type(cfg.default_model) ~= "string" then
    return false, "default_model must be a string"
  end

  if cfg.ui and type(cfg.ui) ~= "table" then
    return false, "ui configuration must be a table"
  end

  if cfg.history and type(cfg.history) ~= "table" then
    return false, "history configuration must be a table"
  end

  if cfg.telescope and type(cfg.telescope) ~= "table" then
    return false, "telescope configuration must be a table"
  end

  if cfg.statusline and type(cfg.statusline) ~= "table" then
    return false, "statusline configuration must be a table"
  end

  if cfg.verbose ~= nil and type(cfg.verbose) ~= "boolean" then
    return false, "verbose must be a boolean"
  end

  -- Validate timeout
  if cfg.timeout ~= nil and type(cfg.timeout) ~= "number" then
    return false, "timeout must be a number"
  end

  if cfg.timeout ~= nil and cfg.timeout < 0 then
    return false, "timeout must be positive"
  end

  return true, nil
end

return M
