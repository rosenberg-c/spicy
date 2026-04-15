-- Shared CLI argument helpers
local M = {}

local config = require("spicy.config")
local fs = require("spicy.utils.fs")

local MODEL_ENV_KEY = "SPICY_MODEL"

local function add_model(args, model)
  if model then
    table.insert(args, "-m")
    table.insert(args, model)
  end
end

local function add_verbose(args, verbose)
  if verbose then
    table.insert(args, "-v")
  end
end

local function add_history(args, enabled)
  if enabled then
    table.insert(args, "--history")
  end
end

local function trim(s)
  return (s:gsub("^%s+", ""):gsub("%s+$", ""))
end

local function trim_inline_comment(value)
  if value == "" then
    return ""
  end

  if value:sub(1, 1) == '"' or value:sub(1, 1) == "'" then
    return value
  end

  local idx = value:find(" #", 1, true)
  if idx then
    return trim(value:sub(1, idx - 1))
  end

  return value
end

local function lookup_model_in_env_file(path)
  local content, err = fs.read_file(path)
  if not content or err then
    return nil
  end

  for line in content:gmatch("[^\r\n]+") do
    local candidate = trim(line)
    if candidate ~= "" and candidate:sub(1, 1) ~= "#" then
      candidate = candidate:gsub("^export%s+", "")
      local key, value = candidate:match("^([^=]+)=(.*)$")
      if key and value then
        key = trim(key)
        if key == MODEL_ENV_KEY then
          value = trim(value)
          value = trim_inline_comment(value)
          local first = value:sub(1, 1)
          local last = value:sub(-1)
          if (first == '"' and last == '"') or (first == "'" and last == "'") then
            value = value:sub(2, -2)
          end
          value = trim(value)
          if value ~= "" then
            return value
          end
        end
      end
    end
  end

  return nil
end

local function lookup_env_model()
  local env_model = vim.env[MODEL_ENV_KEY]
  if env_model and trim(env_model) ~= "" then
    return trim(env_model)
  end

  local local_env_model = lookup_model_in_env_file(vim.fn.getcwd() .. "/.env")
  if local_env_model then
    return local_env_model
  end

  local home = vim.loop.os_homedir()
  if not home or home == "" then
    return nil
  end

  return lookup_model_in_env_file(home .. "/.config/spicy/.env")
end

local function resolve_model(command_name, opts)
  if opts.model and trim(opts.model) ~= "" then
    return trim(opts.model)
  end

  local env_model = lookup_env_model()
  if env_model then
    return env_model
  end

  return config.get("models." .. command_name)
end

-- Build base args with model/verbose and optional history
function M.base_args(command_name, opts, include_history)
  opts = opts or {}
  local args = {}

  local model = resolve_model(command_name, opts)
  add_model(args, model)

  local verbose = opts.verbose or config.get("verbose")
  add_verbose(args, verbose)

  if include_history ~= false then
    add_history(args, true)
  end

  return args
end

return M
