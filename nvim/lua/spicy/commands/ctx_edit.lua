-- SpicyCtxEdit command implementation
-- Rule #5: Assume operations are slow and fragile
-- Rule #20: Return results and errors consistently

local M = {}

local job = require("spicy.utils.job")
local config = require("spicy.config")
local input_ui = require("spicy.ui.input")
local helpers = require("spicy.utils.helpers")
local cli = require("spicy.utils.cli")
local inline_spinner = require("spicy.utils.inline_spinner")
local jsonutil = require("spicy.utils.json")

--- Build the spicy ctx-edit command
--- @param prompt string Instruction for update
--- @param context string Selected context
--- @param opts table|nil Options (model, verbose)
--- @return string, table Command and args
local function build_command(prompt, context, opts)
  opts = opts or {}

  local bin = config.get("bin.ctx_edit") or "ctx-edit"
  local args = cli.base_args("ctx_edit", opts, false)

  -- JSON output for reliable parsing
  table.insert(args, "--json")

  -- Add prompt and context
  table.insert(args, "--prompt")
  table.insert(args, prompt)
  table.insert(args, "--context")
  table.insert(args, context)

  return bin, args
end

local function apply_update(range, updated_text)
  local replacement = vim.split(updated_text, "\n", { plain = true })
  if range.start_col and range.end_col then
    vim.api.nvim_buf_set_text(
      range.bufnr,
      range.start_line - 1,
      range.start_col - 1,
      range.end_line - 1,
      range.end_col,
      replacement
    )
    return
  end

  vim.api.nvim_buf_set_lines(
    range.bufnr,
    range.start_line - 1,
    range.end_line,
    false,
    replacement
  )
end

--- Execute ctx-edit command
--- @param prompt string Instruction for update
--- @param selection string Selected context
--- @param range table Range info (bufnr, start_line, end_line)
--- @param opts table|nil Options
--- @param callback function|nil Callback(updated_text, err)
function M.execute(prompt, selection, range, opts, callback)
  opts = opts or {}

  local cmd, args = build_command(prompt, selection, opts)

  if not job.command_exists(cmd) then
    local err = ("Command not found: %s"):format(cmd)
    helpers.error(err)
    if callback then
      callback(nil, err)
    end
    return
  end

  if config.get("verbose") then
    helpers.info(("Running: %s %s"):format(cmd, table.concat(args, " ")))
  end

  local inline_spinner = nil
  local show_spinner = config.get("ui.ctx_edit.show_spinner")
  if show_spinner == nil then
    show_spinner = true
  end
  if show_spinner then
    inline_spinner = inline_spinner.start(
      range.bufnr,
      range.start_line,
      range.end_line
    )
  end

  job.run(cmd, args, {
    timeout = opts.timeout or config.get("timeout"),
    on_exit = function(stdout, stderr, code)
      if inline_spinner then
        vim.schedule(function()
          inline_spinner.stop()
        end)
      end

      if code ~= 0 then
        local err_msg = table.concat(stderr, "\n")
        helpers.error(
          ("ctx-edit failed (exit code %d): %s"):format(code, err_msg)
        )
        if callback then
          callback(nil, err_msg)
        end
        return
      end

    local raw_output = table.concat(stdout, "\n")
    raw_output = raw_output:gsub("%z", "")
    raw_output = raw_output:gsub("\r", "")

    local payload, err = jsonutil.decode_loose(raw_output)
    if err then
      if config.get("verbose") and raw_output ~= "" then
        helpers.info("ctx-edit raw stdout: " .. raw_output)
      end
      helpers.error("Failed to parse ctx-edit output: " .. err)
      if callback then
        callback(nil, err)
      end
      return
    end

    if not payload.updated_text then
      local err_msg = "missing updated_text in response"
      helpers.error("Failed to parse ctx-edit output: " .. err_msg)
      if callback then
        callback(nil, err_msg)
      end
      return
    end

      local updated_text = payload.updated_text
      if helpers.trim(updated_text) == "" then
        helpers.warn("ctx-edit returned empty update")
        if callback then
          callback(nil, "empty update")
        end
        return
      end

      vim.schedule(function()
        apply_update(range, updated_text)
        helpers.info("Selection updated")
        if callback then
          callback(updated_text, nil)
        end
      end)
    end,
    on_timeout = function()
      if inline_spinner then
        vim.schedule(function()
          inline_spinner.stop()
        end)
      end
      helpers.error("ctx-edit timed out")
      if callback then
        callback(nil, "timeout")
      end
    end,
  })
end

--- Update selection with ctx-edit (main entry point)
--- @param opts table|nil Options:
---   - range: {start_line, end_line}
---   - buffer: number
---   - on_complete: function Completion callback
function M.ctx_edit(opts)
  opts = opts or {}

  if not opts.range then
    helpers.error("No selection range provided")
    return
  end

  local bufnr = opts.buffer or 0
  if opts.range and not opts.selection_text then
    local selection, start_line, start_col, end_line, end_col = helpers.get_visual_selection()
    if selection and start_line == opts.range.start_line and end_line == opts.range.end_line then
      opts.selection_text = selection
      opts.range.start_col = start_col
      opts.range.end_col = end_col
    end
  end

  local selection = opts.selection_text
  if not selection then
    selection = helpers.get_lines_in_range(
      bufnr,
      opts.range.start_line,
      opts.range.end_line
    )
  end

  if not selection or selection == "" then
    helpers.error("No selection content")
    return
  end

  input_ui.prompt({
    prompt = "Edit selection: ",
    default = "",
  }, function(prompt)
    if not prompt or prompt == "" then
      helpers.info("Edit cancelled")
      return
    end

    local range = {
      bufnr = bufnr,
      start_line = opts.range.start_line,
      start_col = opts.range.start_col,
      end_line = opts.range.end_line,
      end_col = opts.range.end_col,
    }

    M.execute(prompt, selection, range, opts, opts.on_complete)
  end)
end

--- Update visual selection with ctx-edit
--- @param opts table|nil Options
function M.ctx_edit_visual(opts)
  opts = opts or {}

  local selection, start_line, start_col, end_line, end_col = helpers.get_visual_selection()
  if not selection then
    helpers.error("No visual selection")
    return
  end

  opts.selection_text = selection
  opts.range = {
    start_line = start_line,
    start_col = start_col,
    end_line = end_line,
    end_col = end_col,
  }

  M.ctx_edit(opts)
end

return M
