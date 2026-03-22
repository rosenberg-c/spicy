-- SpicyExplain command implementation
-- Rule #5: Assume operations are slow and fragile
-- Rule #20: Return results and errors consistently

local M = {}

local job = require("spicy.utils.job")
local config = require("spicy.config")
local float = require("spicy.ui.float")
local spinner = require("spicy.ui.spinner")
local helpers = require("spicy.utils.helpers")
local fs = require("spicy.utils.fs")

--- Build the spicy explain command
--- @param code string The code to explain
--- @param opts table|nil Options (model, verbose, language)
--- @return string, table Command and args
local function build_command(code, opts)
  opts = opts or {}

  local bin = config.get("bin.explain") or "explain"
  local args = {}

  -- Add model flag
  local model = opts.model or config.get("models.explain")
  if model then
    table.insert(args, "-m")
    table.insert(args, model)
  end

  -- Add verbose flag
  local verbose = opts.verbose or config.get("verbose")
  if verbose then
    table.insert(args, "-v")
  end

  -- Add language flag
  if opts.language then
    table.insert(args, "-l")
    table.insert(args, opts.language)
  end

  -- Add history flag
  table.insert(args, "--history")

  -- Add --no-save flag (we handle saving ourselves)
  table.insert(args, "--no-save")

  if opts.snippet then
    table.insert(args, "--snippet")
  end

  if opts.context_file then
    table.insert(args, "--context-file")
    table.insert(args, opts.context_file)
  elseif opts.context then
    table.insert(args, "--context")
    table.insert(args, opts.context)
  end

  return bin, args, code
end

--- Execute explain command
--- @param code string Code to explain
--- @param opts table|nil Options
--- @param callback function|nil Callback(explanation, err)
function M.execute(code, opts, callback)
  opts = opts or {}

  local context_tmp_file = nil
  if opts.context and not opts.context_file then
    context_tmp_file = vim.fn.tempname() .. "-context.txt"
    local ok, err = fs.write_file(context_tmp_file, opts.context)
    if not ok then
      helpers.error("Failed to create context file: " .. err)
      return
    end
    opts.context_file = context_tmp_file
    opts.context = nil
  end

  -- Build command (stdin will be used for code)
  local cmd, args, stdin_data = build_command(code, opts)

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
  local spinner_id = spinner.start("Generating explanation...")

  -- Write code to temp file (explain command expects file or stdin)
  local tmp_file = vim.fn.tempname() .. ".txt"
  local write_ok, write_err = fs.write_file(tmp_file, stdin_data)
  if not write_ok then
    if spinner_id then
      spinner.stop(spinner_id)
    end
    helpers.error("Failed to create temp file: " .. write_err)
    return
  end

  -- Add temp file as argument
  table.insert(args, tmp_file)

  -- Run command
  job.run(cmd, args, {
    timeout = opts.timeout or config.get("timeout"),
    on_exit = function(stdout, stderr, code_exit)
      -- Cleanup temp files
      vim.loop.fs_unlink(tmp_file)
      if context_tmp_file then
        vim.loop.fs_unlink(context_tmp_file)
      end

      -- Stop spinner
      if spinner_id then
        spinner.stop(spinner_id)
      end

      -- Check for errors
      if code_exit ~= 0 then
        local err_msg = table.concat(stderr, "\n")
        helpers.error(
          ("Explain command failed (exit code %d): %s"):format(
            code_exit,
            err_msg
          )
        )
        if callback then
          callback(nil, err_msg)
        end
        return
      end

      -- Get explanation
      local explanation = table.concat(stdout, "\n")
      explanation = helpers.trim(explanation)

      if explanation == "" then
        helpers.warn("Received empty explanation")
        if callback then
          callback(nil, "empty explanation")
        end
        return
      end

      -- Handle output
      vim.schedule(function()
        -- Save to file if auto_save enabled
        local auto_save = opts.auto_save
        if auto_save == nil then
          auto_save = config.get("ui.explain.auto_save")
        end

        local filepath = nil
        if auto_save then
          local save_dir = config.get("ui.explain.save_dir")
          save_dir = fs.expand(save_dir)

          -- Create directory if needed
          if not fs.exists(save_dir) then
            fs.mkdir(save_dir)
          end

          -- Generate filename
          local source_file = opts.source_name or "code"
          local filename = source_file .. "-explanation.md"
          filepath = fs.join(save_dir, filename)

          local ok, err = fs.write_file(filepath, explanation)
          if ok then
            helpers.info(("Explanation saved to: %s"):format(filepath))
          else
            helpers.error(("Failed to save: %s"):format(err))
          end
        end

        -- Display in buffer
        local bufnr = helpers.create_scratch_buffer()
        vim.api.nvim_buf_set_lines(
          bufnr,
          0,
          -1,
          false,
          vim.split(explanation, "\n", { plain = true })
        )
        vim.bo[bufnr].filetype = "markdown"
        vim.cmd("buffer " .. bufnr)

        if callback then
          callback(explanation, nil)
        end
      end)
    end,
  })
end

--- Explain code (main entry point)
--- @param opts table|nil Options:
---   - range: {start_line, end_line}
---   - buffer: number
---   - language: string
function M.explain(opts)
  opts = opts or {}

  local code
  local source_name

  if opts.range and not opts.selection_text then
    local selection, start_line, start_col, end_line, end_col = helpers.get_visual_selection()
    if selection and start_line == opts.range.start_line and end_line == opts.range.end_line then
      opts.selection_text = selection
      opts.range.start_col = start_col
      opts.range.end_col = end_col
    end
  end

  -- Determine what to explain
  if opts.selection_text then
    code = opts.selection_text
    source_name = helpers.get_filename()
    opts.source_name = source_name
  elseif opts.range then
    -- Explain range
    local bufnr = opts.buffer or 0
    code = helpers.get_lines_in_range(
      bufnr,
      opts.range.start_line,
      opts.range.end_line
    )
    source_name = helpers.get_filename()
    opts.source_name = source_name
    opts.snippet = true
    local surround = config.get("ui.explain.context_surround_lines")
    if not surround or surround < 0 then
      surround = 80
    end
    if surround > 0 then
      local total_lines = vim.api.nvim_buf_line_count(bufnr)
      local start_line = math.max(1, opts.range.start_line - surround)
      local end_line = math.min(total_lines, opts.range.end_line + surround)
      local context = helpers.get_lines_in_range(bufnr, start_line, end_line)
      context = (
        "Context lines %d-%d of %d (around selection).\n\n%s"
      ):format(start_line, end_line, total_lines, context)
      local max_chars = config.get("ui.explain.context_max_chars")
      if max_chars == nil or max_chars <= 0 then
        max_chars = 3000
      end
      if #context > max_chars then
        context = context:sub(1, max_chars)
          .. "\n\n[Context truncated to max chars]\n"
      end
      opts.context = context
    end
  else
    -- Explain whole buffer
    local bufnr = opts.buffer or 0
    code = helpers.get_buffer_content(bufnr)
    source_name = helpers.get_filename()
    opts.source_name = source_name
  end

  -- Auto-detect language if not specified
  if not opts.language then
    opts.language = helpers.get_filetype()
  end

  if not code or code == "" then
    helpers.error("No code to explain")
    return
  end

  if config.get("verbose") then
    local ctx_len = 0
    if opts.context then
      ctx_len = #opts.context
    end
    helpers.info(
      ("Explain selection len=%d, context len=%d"):format(#code, ctx_len)
    )
  end

  M.execute(code, opts, opts.on_complete)
end

--- Explain visual selection
--- @param opts table|nil Options
function M.explain_visual(opts)
  opts = opts or {}

  -- Get visual selection
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

  M.explain(opts)
end

return M
