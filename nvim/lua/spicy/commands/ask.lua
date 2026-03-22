-- SpicyAsk command implementation
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

--- Build the spicy ask command
--- @param question string The question to ask
--- @param opts table|nil Options (model, verbose, context)
--- @return table, table Command and args
local function build_command(question, opts)
  opts = opts or {}

  local bin = config.get("bin.ask") or "ask"
  local args = cli.base_args("ask", opts, true)

  -- Add question
  table.insert(args, question)

  return bin, args
end

--- Display answer in configured output
--- @param answer string The answer to display
--- @param opts table|nil Options
local function display_answer(answer, opts)
  opts = opts or {}

  local output_mode = opts.output or config.get("ui.ask.output") or "float"

  if output_mode == "float" then
    local float_opts = config.get("ui.ask.float_opts") or {}
    local syntax = config.get("ui.ask.syntax") or "markdown"

    local bufnr, winid = float.create({
      width = float_opts.width or 0.8,
      height = float_opts.height or 0.6,
      border = float_opts.border or "rounded",
      title = float_opts.title or " Spicy Ask ",
      title_pos = float_opts.title_pos or "center",
      content = vim.split(answer, "\n", { plain = true }),
      syntax = syntax,
    })

    -- Store for potential reuse
    return bufnr, winid
  elseif output_mode == "buffer" then
    -- Create new buffer with answer
    local bufnr = helpers.create_scratch_buffer()
    vim.api.nvim_buf_set_lines(
      bufnr,
      0,
      -1,
      false,
      vim.split(answer, "\n", { plain = true })
    )
    vim.bo[bufnr].syntax = "markdown"
    vim.cmd("buffer " .. bufnr)
    return bufnr, nil
  elseif output_mode == "split" then
    vim.cmd("new")
    local bufnr = vim.api.nvim_get_current_buf()
    vim.api.nvim_buf_set_lines(
      bufnr,
      0,
      -1,
      false,
      vim.split(answer, "\n", { plain = true })
    )
    vim.bo[bufnr].syntax = "markdown"
    vim.bo[bufnr].buftype = "nofile"
    return bufnr, nil
  end
end

--- Execute ask command
--- @param question string The question to ask
--- @param opts table|nil Options
--- @param callback function|nil Callback(answer, err)
function M.execute(question, opts, callback)
  opts = opts or {}

  -- Build command
  local cmd, args = build_command(question, opts)

  -- Debug: Show what we're running (only if verbose)
  if config.get("verbose") then
    local cmd_str = cmd .. " " .. table.concat(args, " ")
    helpers.info("Running: " .. cmd_str)
  end

  -- Check if command exists
  if not job.command_exists(cmd) then
    local err = ("Command not found: %s"):format(cmd)
    helpers.error(err)
    if callback then
      callback(nil, err)
    end
    return
  end

  -- Start spinner
  local show_spinner = config.get("ui.ask.show_spinner")
  local spinner_id = nil
  if show_spinner then
    spinner_id = spinner.start("Asking question...")
  end

  -- Run command
  local stdout_lines = {}
  local stderr_lines = {}

  job.run(cmd, args, {
    timeout = opts.timeout or config.get("timeout"),
    on_stdout = function(_, data)
      if data then
        table.insert(stdout_lines, data)
      end
    end,
    on_stderr = function(_, data)
      if data then
        table.insert(stderr_lines, data)
      end
    end,
    on_exit = function(stdout, stderr, code)
      -- Stop spinner
      if spinner_id then
        spinner.stop(spinner_id)
      end

      -- Rule #5: Propagate errors with context
      if code ~= 0 then
        local err_msg = table.concat(stderr, "\n")
        helpers.error(
          ("Ask command failed (exit code %d): %s"):format(code, err_msg)
        )
        if callback then
          callback(nil, err_msg)
        end
        return
      end

      -- Get answer
      local answer = table.concat(stdout, "\n")
      answer = helpers.trim(answer)

      if answer == "" then
        local err_detail = ("Empty answer. " ..
          "Exit code: %d, " ..
          "Stdout lines: %d, " ..
          "Stderr lines: %d"):format(
            code,
            #stdout,
            #stderr
          )
        vim.schedule(function()
          helpers.warn(err_detail)

          -- Debug: show what we got
          if #stderr > 0 then
            helpers.info("Stderr: " .. table.concat(stderr, "\n"))
          end
        end)

        if callback then
          callback(nil, "empty answer")
        end
        return
      end

      -- Display answer (must be scheduled - we're in a callback)
      vim.schedule(function()
        display_answer(answer, opts)

        if callback then
          callback(answer, nil)
        end
      end)
    end,
    on_timeout = function()
      if spinner_id then
        spinner.stop(spinner_id)
      end
      helpers.error("Ask command timed out")
      if callback then
        callback(nil, "timeout")
      end
    end,
  })
end

--- Ask a question (main entry point)
--- @param question string|nil The question (nil to prompt)
--- @param opts table|nil Options:
---   - model: string Model to use
---   - verbose: boolean Verbose output
---   - context: string Additional context
---   - output: string Output mode
---   - on_complete: function Completion callback
function M.ask(question, opts)
  opts = opts or {}

  -- If question not provided, prompt for it
  if not question or question == "" then
    input_ui.prompt({
      prompt = "Ask a question: ",
      default = "",
    }, function(input)
      if not input or input == "" then
        helpers.info("Ask cancelled")
        return
      end

      M.execute(input, opts, opts.on_complete)
    end)
    return
  end

  -- Add context if provided
  local full_question = question
  if opts.context then
    full_question = ("%s\n\nContext:\n%s"):format(question, opts.context)
  end

  M.execute(full_question, opts, opts.on_complete)
end

--- Ask about visual selection
--- @param opts table|nil Options
function M.ask_visual(opts)
  opts = opts or {}

  -- Get visual selection
  local selection, start_line, _, end_line, _ = helpers.get_visual_selection()

  if not selection then
    helpers.error("No visual selection")
    return
  end

  -- Prompt for question about the selection
  input_ui.prompt({
    prompt = "Ask about selection: ",
    default = "",
  }, function(question)
    if not question or question == "" then
      helpers.info("Ask cancelled")
      return
    end

    -- Add selection as context
    opts.context = ("Selected code (lines %d-%d):\n%s"):format(
      start_line,
      end_line,
      selection
    )

    M.execute(question, opts, opts.on_complete)
  end)
end

return M
