-- SpicyCtxEdit command implementation
-- Rule #5: Assume operations are slow and fragile
-- Rule #20: Return results and errors consistently

local M = {}

local job = require("spicy.utils.job")
local config = require("spicy.config")
local input_ui = require("spicy.ui.input")
local helpers = require("spicy.utils.helpers")

local INLINE_SPINNER_FRAMES = { "-", "\\", "|", "/" }
local inline_spinner_ns = vim.api.nvim_create_namespace("spicy.ctx_edit")

--- Build the spicy ctx-edit command
--- @param prompt string Instruction for update
--- @param context string Selected context
--- @param opts table|nil Options (model, verbose)
--- @return string, table Command and args
local function build_command(prompt, context, opts)
  opts = opts or {}

  local bin = config.get("bin.ctx_edit") or "ctx-edit"
  local args = {}

  -- Add model flag
  local model = opts.model or config.get("models.ctx_edit")
  if model then
    table.insert(args, "-m")
    table.insert(args, model)
  end

  -- Add verbose flag
  local verbose = opts.verbose or config.get("verbose")
  if verbose then
    table.insert(args, "-v")
  end

  -- JSON output for reliable parsing
  table.insert(args, "--json")

  -- Add prompt and context
  table.insert(args, "--prompt")
  table.insert(args, prompt)
  table.insert(args, "--context")
  table.insert(args, context)

  return bin, args
end

local function decode_response(stdout)
  local output = table.concat(stdout, "\n")
  output = helpers.trim(output)
  output = output:gsub("%z", "")
  output = output:gsub("\r", "")

  if output == "" then
    return nil, "empty response", output
  end

  local ok, decoded = pcall(vim.fn.json_decode, output)
  if not ok or type(decoded) ~= "table" then
    if type(vim.json) == "table" and type(vim.json.decode) == "function" then
      local json_ok, json_decoded = pcall(vim.json.decode, output)
      if json_ok and type(json_decoded) == "table" then
        decoded = json_decoded
        ok = true
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    local lines = vim.split(output, "\n", { plain = true })
    for i = #lines, 1, -1 do
      local line = helpers.trim(lines[i])
      if line ~= "" then
        local line_ok, line_decoded = pcall(vim.fn.json_decode, line)
        if line_ok then
          if type(line_decoded) == "table" then
            decoded = line_decoded
            ok = true
            break
          end
          if type(line_decoded) == "string" then
            local nested_ok, nested_decoded = pcall(vim.fn.json_decode, line_decoded)
            if nested_ok and type(nested_decoded) == "table" then
              decoded = nested_decoded
              ok = true
              break
            end
          end
        end
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    local first = output:find("{", 1, true)
    local last = nil
    for i = #output, 1, -1 do
      if output:sub(i, i) == "}" then
        last = i
        break
      end
    end
    if first and last and last > first then
      local candidate = output:sub(first, last)
      local cand_ok, cand_decoded = pcall(vim.fn.json_decode, candidate)
      if cand_ok and type(cand_decoded) == "table" then
        decoded = cand_decoded
        ok = true
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    local stripped = helpers.trim(output)
    local quote = stripped:sub(1, 1)
    if (quote == '"' or quote == "'") and stripped:sub(-1) == quote then
      local unquoted = stripped:sub(2, -2)
      local uq_ok, uq_decoded = pcall(vim.fn.json_decode, unquoted)
      if uq_ok and type(uq_decoded) == "table" then
        decoded = uq_decoded
        ok = true
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    return nil, "invalid json response", output
  end

  if not decoded.updated_text then
    return nil, "missing updated_text in response", output
  end

  return decoded, nil, output
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

local function start_inline_spinner(bufnr, start_line, end_line)
  local frame = 1
  local timer = vim.loop.new_timer()
  local top_line = math.max(start_line - 1, 0)
  local bottom_line = math.max(end_line - 1, 0)

  local function render()
    local glyph = INLINE_SPINNER_FRAMES[frame]
    local text = glyph .. " Updating selection..."
    local opts = {
      virt_text = { { text, "Comment" } },
      virt_text_pos = "eol",
    }

    vim.api.nvim_buf_set_extmark(
      bufnr,
      inline_spinner_ns,
      top_line,
      0,
      opts
    )
    vim.api.nvim_buf_set_extmark(
      bufnr,
      inline_spinner_ns,
      bottom_line,
      0,
      opts
    )
  end

  local function tick()
    frame = (frame % #INLINE_SPINNER_FRAMES) + 1
    vim.api.nvim_buf_clear_namespace(bufnr, inline_spinner_ns, 0, -1)
    render()
  end

  vim.schedule(render)
  timer:start(0, 120, vim.schedule_wrap(tick))

  return {
    stop = function()
      timer:stop()
      timer:close()
      vim.api.nvim_buf_clear_namespace(bufnr, inline_spinner_ns, 0, -1)
    end,
  }
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
    inline_spinner = start_inline_spinner(
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

      local payload, err, raw_output = decode_response(stdout)
      if err then
        if config.get("verbose") and raw_output and raw_output ~= "" then
          helpers.info("ctx-edit raw stdout: " .. raw_output)
        end
        helpers.error("Failed to parse ctx-edit output: " .. err)
        if callback then
          callback(nil, err)
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
