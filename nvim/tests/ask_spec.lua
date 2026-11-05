-- Tests for spicy.commands.ask module (context transport via context file)
local ask
local input_ui
local runner
local fs

local scratch_bufnr = nil

local function clear_marks()
  vim.fn.setpos("'<", { 0, 0, 0, 0 })
  vim.fn.setpos("'>", { 0, 0, 0, 0 })
end

local function set_marks(start_line, end_line)
  vim.fn.setpos("'<", { scratch_bufnr, start_line, 0, 0 })
  vim.fn.setpos("'>", { scratch_bufnr, end_line, 999, 0 })
end

-- vim.fn.chdir returns the previous cwd on success, so verify via getcwd.
local function chdir(dir)
  vim.fn.chdir(dir)
  assert.equals(vim.fn.resolve(dir), vim.fn.resolve(vim.fn.getcwd()))
end

describe("spicy.commands.ask", function()
  local original_cwd
  local test_dir

  before_each(function()
    package.loaded["spicy.commands.ask"] = nil

    ask = require("spicy.commands.ask")
    input_ui = require("spicy.ui.input")
    runner = require("spicy.utils.runner")
    fs = require("spicy.utils.fs")

    -- Run with the cwd inside a temp dir so local config discovery
    -- (walk-up from cwd) is deterministic and finds no .spicy.config.json.
    test_dir = vim.fn.tempname()
    vim.fn.mkdir(test_dir, "p")
    original_cwd = vim.fn.getcwd()
    chdir(test_dir)

    scratch_bufnr = vim.api.nvim_create_buf(true, false)
    vim.api.nvim_buf_set_lines(
      scratch_bufnr,
      0,
      -1,
      false,
      { "alpha", "beta", "gamma" }
    )
    vim.api.nvim_set_current_buf(scratch_bufnr)

    set_marks(1, 3)
  end)

  after_each(function()
    clear_marks()

    if scratch_bufnr and vim.api.nvim_buf_is_valid(scratch_bufnr) then
      vim.api.nvim_buf_delete(scratch_bufnr, { force = true })
    end
    scratch_bufnr = nil

    if original_cwd then
      vim.fn.chdir(original_cwd)
    end

    if test_dir and vim.fn.isdirectory(test_dir) == 1 then
      vim.fn.delete(test_dir, "rf")
    end
    test_dir = nil

    package.loaded["spicy.commands.ask"] = nil
  end)

  it("should pass the question and context through to execute without inlining", function()
    -- @req NVIM-ASK-001
    local recorded = {}
    local original_execute = ask.execute

    ask.execute = function(question, opts, _callback)
      recorded.question = question
      recorded.opts = opts
    end

    local context = "Selected code (lines 1-3):\nalpha\nbeta\ngamma"
    ask.ask("What does this do?", { context = context })

    assert.equals("What does this do?", recorded.question)
    assert.equals(context, recorded.opts.context)
    assert.is_nil(recorded.opts.context_file)

    ask.execute = original_execute
  end)

  it("should preserve the question and context when prompting", function()
    -- @req NVIM-ASK-001
    local recorded = {}
    local prompt_cb = nil
    local original_execute = ask.execute
    local original_prompt = input_ui.prompt

    ask.execute = function(question, opts)
      recorded.question = question
      recorded.opts = opts
    end
    input_ui.prompt = function(_, callback)
      prompt_cb = callback
    end

    local context = "Selected code (lines 1-3):\nalpha\nbeta\ngamma"
    ask.ask(nil, { context = context })

    assert.is_function(prompt_cb)
    prompt_cb("What does this do?")

    assert.equals("What does this do?", recorded.question)
    assert.equals(context, recorded.opts.context)

    ask.execute = original_execute
    input_ui.prompt = original_prompt
  end)

  it("should set the visual selection as context when a question is provided", function()
    -- @req NVIM-ASK-001
    local recorded = nil
    local original_ask = ask.ask

    ask.ask = function(question, opts)
      recorded = { question = question, opts = opts }
    end

    ask.ask_visual({ question = "Explain this" })

    assert.equals("Explain this", recorded.question)
    assert.equals(
      "Selected code (lines 1-3):\nalpha\nbeta\ngamma",
      recorded.opts.context
    )

    ask.ask = original_ask
  end)

  it("should prompt for the question and keep the selection context", function()
    -- @req NVIM-ASK-001
    local recorded = nil
    local prompt_cb = nil
    local original_ask = ask.ask
    local original_prompt = input_ui.prompt

    ask.ask = function(question, opts)
      recorded = { question = question, opts = opts }
    end
    input_ui.prompt = function(_, callback)
      prompt_cb = callback
    end

    ask.ask_visual({})

    assert.is_function(prompt_cb)
    prompt_cb("Explain this")

    assert.equals("Explain this", recorded.question)
    assert.equals(
      "Selected code (lines 1-3):\nalpha\nbeta\ngamma",
      recorded.opts.context
    )

    ask.ask = original_ask
    input_ui.prompt = original_prompt
  end)

  it("should report an error and not ask when there is no visual selection", function()
    -- @req NVIM-ASK-001
    clear_marks()

    local ask_called = false
    local prompt_called = false
    local original_ask = ask.ask
    local original_prompt = input_ui.prompt

    ask.ask = function()
      ask_called = true
    end
    input_ui.prompt = function()
      prompt_called = true
    end

    ask.ask_visual({ question = "Explain this" })

    assert.is_false(ask_called)
    assert.is_false(prompt_called)

    ask.ask = original_ask
    input_ui.prompt = original_prompt
  end)

  it("should materialize inline context to a --context-file arg and clean it up", function()
    -- @req NVIM-ASK-001
    local recorded = nil
    local original_run = runner.run
    local original_notify = vim.notify

    runner.run = function(_cmd, args, opts)
      recorded = { args = args }

      local ctx_idx = nil
      for i, a in ipairs(args) do
        if a == "--context-file" then
          ctx_idx = i
          break
        end
      end
      assert.is_not_nil(
        ctx_idx,
        "expected --context-file in args: " .. table.concat(args, " ")
      )
      recorded.ctx_file = args[ctx_idx + 1]
      assert.is_not_nil(recorded.ctx_file)

      recorded.file_existed = fs.exists(recorded.ctx_file)
      local content, read_err = fs.read_file(recorded.ctx_file)
      recorded.file_content = content
      recorded.file_read_err = read_err

      -- Simulate a failed exit: this triggers cleanup of the temp file and
      -- returns before any display/schedule path.
      if opts.on_exit then
        opts.on_exit({}, { "boom" }, 1)
      end
    end

    vim.notify = function() end

    ask.execute("What does this do?", { context = "alpha\nbeta\ngamma" })

    assert.is_nil(recorded.file_read_err)
    assert.is_true(recorded.file_existed, "context file should exist during run")
    assert.equals("alpha\nbeta\ngamma", recorded.file_content)

    -- After on_exit ran, the temp file must be cleaned up.
    assert.is_false(
      fs.exists(recorded.ctx_file),
      "context file should be removed after exit"
    )

    runner.run = original_run
    vim.notify = original_notify
  end)
end)
