-- SpicyGitmessage command implementation
-- Rule #5: Assume operations are slow and fragile
-- Rule #20: Return results and errors consistently

local M = {}

local job = require("spicy.utils.job")
local config = require("spicy.config")
local float = require("spicy.ui.float")
local input_ui = require("spicy.ui.input")
local spinner = require("spicy.ui.spinner")
local helpers = require("spicy.utils.helpers")
local cli = require("spicy.utils.cli")

--- Build the spicy gitmessage command
--- @param opts table|nil Options (model, verbose, hint, prefix)
--- @return string, table Command and args
local function build_command(opts)
  opts = opts or {}

  local bin = config.get("bin.gitmessage") or "gitmessage"
  local args = cli.base_args("gitmessage", opts, true)

  -- Add copy flag
  if opts.copy or config.get("ui.gitmessage.auto_copy") then
    table.insert(args, "-c")
  end

  -- Add prefix if provided
  if opts.prefix then
    table.insert(args, opts.prefix)
  end

  -- Add hint if provided
  if opts.hint then
    table.insert(args, opts.hint)
  end

  return bin, args
end

--- Execute gitmessage command
--- @param opts table|nil Options
--- @param callback function|nil Callback(message, err)
function M.execute(opts, callback)
  opts = opts or {}

  -- Build command
  local cmd, args = build_command(opts)

  -- Check if command exists
  if not job.command_exists(cmd) then
    local err = ("Command not found: %s"):format(cmd)
    helpers.error(err)
    if callback then
      callback(nil, err)
    end
    return
  end

  -- Debug output
  if config.get("verbose") then
    helpers.info(("Running: %s %s"):format(cmd, table.concat(args, " ")))
  end

  -- Start spinner
  local spinner_id = spinner.start("Generating commit message...")

  -- Run command
  job.run(cmd, args, {
    timeout = opts.timeout or config.get("timeout"),
    on_exit = function(stdout, stderr, code)
      -- Stop spinner
      if spinner_id then
        spinner.stop(spinner_id)
      end

      -- Check for errors
      if code ~= 0 then
        local err_msg = table.concat(stderr, "\n")
        helpers.error(
          ("Gitmessage command failed (exit code %d): %s"):format(code, err_msg)
        )
        if callback then
          callback(nil, err_msg)
        end
        return
      end

      -- Get message
      local message = table.concat(stdout, "\n")
      message = helpers.trim(message)

      if message == "" then
        helpers.warn("Received empty commit message")
        if callback then
          callback(nil, "empty message")
        end
        return
      end

      -- Handle output
      vim.schedule(function()
        -- Display in floating window
        local output_mode = opts.output or config.get("ui.gitmessage.output")

        if output_mode == "float" then
          local float_opts = config.get("ui.gitmessage.float_opts") or {}

          float.create({
            width = float_opts.width or 0.6,
            height = float_opts.height or 0.4,
            border = float_opts.border or "rounded",
            title = float_opts.title or " Git Commit Message ",
            content = vim.split(message, "\n", { plain = true }),
            syntax = "gitcommit",
          })
        else
          -- Print to command line
          helpers.info("Commit message: " .. message)
        end

        if callback then
          callback(message, nil)
        end
      end)
    end,
  })
end

--- Generate git commit message (main entry point)
--- @param opts table|nil Options:
---   - prefix: string Prefix (feat, fix, etc.)
---   - hint: string Hint for the message
function M.gitmessage(opts)
  opts = opts or {}
  M.execute(opts, opts.on_complete)
end

return M
