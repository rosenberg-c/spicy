-- SpicyAsk command implementation
-- Rule #5: Assume operations are slow and fragile
-- Rule #20: Return results and errors consistently

local M = {}

local config = require("spicy.config")
local float = require("spicy.ui.float")
local input_ui = require("spicy.ui.input")
local helpers = require("spicy.utils.helpers")
local cli = require("spicy.utils.cli")
local runner = require("spicy.utils.runner")
local fs = require("spicy.utils.fs")

--- Build the spicy ask command
--- @param question string The question to ask
--- @param opts table|nil Options (model, verbose, context)
--- @return table, table Command and args
local function build_command(question, opts)
  opts = opts or {}

  local bin = config.get("bin.ask") or "ask"
  local args = cli.base_args("ask", opts, true)

  -- Add context. A context file is preferred; otherwise inline context is
  -- passed. The two are mutually exclusive on the CLI.
  if opts.context_file then
    table.insert(args, "--context-file")
    table.insert(args, opts.context_file)
  elseif opts.context then
    table.insert(args, "--context")
    table.insert(args, opts.context)
  end

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

  -- Materialize inline context to a temp file so selected code is transported
  -- safely (avoid passing large/multi-line context as a single shell arg).
  local context_tmp_file = nil
  if opts.context and not opts.context_file then
    context_tmp_file = vim.fn.tempname() .. "-context.txt"
    local ok, err = fs.write_file(context_tmp_file, opts.context)
    if not ok then
      helpers.error("Failed to create context file: " .. tostring(err))
      if callback then
        callback(nil, "failed to write context file")
      end
      return
    end
    opts.context_file = context_tmp_file
    opts.context = nil
  end

  local function cleanup_context_file()
    if context_tmp_file then
      vim.loop.fs_unlink(context_tmp_file)
    end
  end

  -- Build command
  local cmd, args = build_command(question, opts)

  local show_spinner = config.get("ui.ask.show_spinner")
  local spinner_message = nil
  if show_spinner then
    spinner_message = "Asking question..."
  end

  runner.run(cmd, args, {
    timeout = opts.timeout or config.get("timeout"),
    spinner_message = spinner_message,
    on_exit = function(stdout, stderr, code)
      cleanup_context_file()

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
      cleanup_context_file()
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

  -- Context (when provided) is transported to the CLI via a context file in
  -- execute, so the question itself stays clean and free of inlined selection.
  local function finalize(q)
    M.execute(q, opts, opts.on_complete)
  end

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

      finalize(input)
    end)
    return
  end

  finalize(question)
end

--- Ask about visual selection
--- @param opts table|nil Options:
---   - question: string Optional question (skips the prompt)
---   - on_complete: function Completion callback
function M.ask_visual(opts)
  opts = opts or {}

  -- Get visual selection
  local selection, start_line, _, end_line, _ = helpers.get_visual_selection()

  if not selection then
    helpers.error("No visual selection")
    return
  end

  -- Add selection as context
  opts.context = ("Selected code (lines %d-%d):\n%s"):format(
    start_line,
    end_line,
    selection
  )

  -- If a question was provided, ask directly
  if opts.question and opts.question ~= "" then
    M.ask(opts.question, opts)
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

    M.ask(question, opts)
  end)
end

return M
