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

-- Build base args with model/verbose and optional history
function M.base_args(command_name, opts, include_history)
  opts = opts or {}
  local args = {}

  local model = opts.model or config.get("models." .. command_name)
  add_model(args, model)

  local verbose = opts.verbose or config.get("verbose")
  add_verbose(args, verbose)

  if include_history ~= false then
    add_history(args, true)
  end

  return args
end

return M
