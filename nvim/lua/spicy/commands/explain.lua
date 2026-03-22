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

  return bin, args, code
end

--- Execute explain command
--- @param code string Code to explain
--- @param opts table|nil Options
--- @param callback function|nil Callback(explanation, err)
function M.execute(code, opts, callback)
  opts = opts or {}

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
      -- Cleanup temp file
      vim.loop.fs_unlink(tmp_file)

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

  -- Determine what to explain
  if opts.range then
    -- Explain range
    local bufnr = opts.buffer or 0
    code = helpers.get_lines_in_range(
      bufnr,
      opts.range.start_line,
      opts.range.end_line
    )
    source_name = helpers.get_filename()
    opts.source_name = source_name
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

  M.execute(code, opts, opts.on_complete)
end

--- Explain visual selection
--- @param opts table|nil Options
function M.explain_visual(opts)
  opts = opts or {}

  -- Get visual selection
  local selection, start_line, end_line = helpers.get_visual_selection()

  if not selection then
    helpers.error("No visual selection")
    return
  end

  opts.range = {
    start_line = start_line,
    end_line = end_line,
  }

  M.explain(opts)
end

return M
