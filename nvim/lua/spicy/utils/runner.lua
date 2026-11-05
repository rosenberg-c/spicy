-- Shared command runner for CLI invocations
local M = {}

local job = require("spicy.utils.job")
local config = require("spicy.config")
local spinner = require("spicy.ui.spinner")
local helpers = require("spicy.utils.helpers")

--- Run a command with standardized logging, spinner, and timeout handling
--- @param cmd string Command to run
--- @param args table Command arguments
--- @param opts table|nil Options:
---   - timeout: number Timeout in ms
---   - spinner_message: string|nil Spinner message
---   - on_stdout: function|nil
---   - on_stderr: function|nil
---   - on_exit: function(stdout, stderr, code)
---   - on_timeout: function|nil
--- @return table|nil, string|nil Job or error
function M.run(cmd, args, opts)
  opts = opts or {}

  if not job.command_exists(cmd) then
    local err = ("Command not found: %s"):format(cmd)
    helpers.error(err)
    if opts.on_exit then
      opts.on_exit({}, { err }, 127)
    end
    return nil, err
  end

  if config.get("verbose") then
    helpers.info(("Running: %s %s"):format(cmd, table.concat(args, " ")))
  end

  local spinner_id = nil
  if opts.spinner_message and opts.spinner_message ~= "" then
    spinner_id = spinner.start(opts.spinner_message)
  end

  local function stop_spinner()
    if spinner_id then
      spinner.stop(spinner_id)
    end
  end

  local result, err = job.run(cmd, args, {
    timeout = opts.timeout or config.get("timeout"),
    on_stdout = opts.on_stdout,
    on_stderr = opts.on_stderr,
    on_exit = function(stdout, stderr, code)
      stop_spinner()
      if opts.on_exit then
        opts.on_exit(stdout, stderr, code)
      end
    end,
    on_timeout = function()
      stop_spinner()
      if opts.on_timeout then
        opts.on_timeout()
      else
        helpers.error("Command timed out")
        if opts.on_exit then
          opts.on_exit({}, { "timeout" }, 1)
        end
      end
    end,
  })

  -- Spawn failed synchronously (job.run returned nil). Stop the spinner and
  -- surface the error so the UI never stays stuck.
  if not result then
    stop_spinner()
    local err_msg = err or "failed to start command"
    helpers.error(err_msg)
    if opts.on_exit then
      opts.on_exit({}, { err_msg }, 127)
    end
    return nil, err_msg
  end

  return result, err
end

return M
