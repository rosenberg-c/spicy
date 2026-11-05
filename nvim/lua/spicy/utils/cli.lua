-- Shared CLI argument helpers
local M = {}

local config = require("spicy.config")

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

local function resolve_model(command_name, opts)
  if opts.model and trim(opts.model) ~= "" then
    return trim(opts.model)
  end

  return config.get_model(command_name)
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
