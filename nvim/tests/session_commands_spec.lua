-- @req NVIM-SESSION-001
describe("session command variants", function()
  local original_spicy
  local original_commands

  before_each(function()
    original_spicy = package.loaded["spicy"]
    original_commands = package.loaded["spicy.commands"]
    vim.g.loaded_spicy = nil
    vim.cmd("runtime plugin/spicy.lua")
  end)

  after_each(function()
    package.loaded["spicy"] = original_spicy
    package.loaded["spicy.commands"] = original_commands
    vim.g.loaded_spicy = nil
  end)

  it("forwards session mode and arguments for ask and tutor", function()
    local calls = {}
    package.loaded["spicy"] = {
      ask = function(question, opts)
        table.insert(calls, { "ask", question, opts.session })
      end,
      tutor = function(topic, opts)
        table.insert(calls, { "tutor", topic, opts and opts.session or false })
      end,
    }
    package.loaded["spicy.commands"] = {
      ask_visual = function(opts)
        table.insert(calls, { "ask_visual", opts.question, opts.session })
      end,
    }

    vim.cmd("SpicyAsk hello world")
    vim.cmd("SpicyAskSession hello world")
    vim.cmd("SpicyTutor subject")
    vim.cmd("SpicyTutorSession subject")

    assert.same({
      { "ask", "hello world", false },
      { "ask", "hello world", true },
      { "tutor", "subject", false },
      { "tutor", "subject", true },
    }, calls)
  end)

  it("forwards session mode and range for explain", function()
    local calls = {}
    vim.api.nvim_buf_set_lines(0, 0, -1, false, { "first", "second" })
    package.loaded["spicy.commands"] = {
      explain = function(opts)
        table.insert(calls, { opts.session, opts.range.start_line, opts.range.end_line })
      end,
    }

    vim.cmd("1,2SpicyExplain")
    vim.cmd("1,2SpicyExplainSession")

    assert.same({ { false, 1, 2 }, { true, 1, 2 } }, calls)
  end)

  it("keeps visual selection routes for session commands", function()
    local calls = {}
    vim.api.nvim_buf_set_lines(0, 0, -1, false, { "first", "second" })
    package.loaded["spicy.commands"] = {
      ask_visual = function(opts)
        table.insert(calls, { "ask", opts.question, opts.session })
      end,
      explain = function(opts)
        table.insert(calls, { "explain", opts.range.start_line, opts.session })
      end,
    }

    vim.cmd("1,2SpicyAskSession question")
    vim.cmd("1,2SpicyExplainSession")

    assert.same({
      { "ask", "question", true },
      { "explain", 1, true },
    }, calls)
  end)
end)
