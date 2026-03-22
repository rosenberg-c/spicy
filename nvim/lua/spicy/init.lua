-- Main entry point for spicy.nvim
-- Rule #15: Keep module boundaries clear
-- Rule #29: Prefer local scope

local M = {}

-- Lazy-load modules
local config = nil
local commands = nil

--- Setup spicy.nvim with user configuration
--- @param opts table|nil User configuration options
function M.setup(opts)
  config = require("spicy.config")
  config.setup(opts)
end

--- Ask a question and display answer
--- @param question string|nil The question to ask
--- @param opts table|nil Optional configuration
function M.ask(question, opts)
  if not commands then
    commands = require("spicy.commands")
  end
  return commands.ask(question, opts)
end

--- Generate a tutorial
--- @param topic string|nil The topic for the tutorial
--- @param opts table|nil Optional configuration
function M.tutor(topic, opts)
  if not commands then
    commands = require("spicy.commands")
  end
  return commands.tutor(topic, opts)
end

--- Explain code
--- @param opts table|nil Optional configuration
function M.explain(opts)
  if not commands then
    commands = require("spicy.commands")
  end
  return commands.explain(opts)
end

--- Generate git commit message
--- @param opts table|nil Optional configuration
function M.gitmessage(opts)
  if not commands then
    commands = require("spicy.commands")
  end
  return commands.gitmessage(opts)
end

--- Update selection with ctx-edit
--- @param opts table|nil Optional configuration
function M.ctx_edit(opts)
  if not commands then
    commands = require("spicy.commands")
  end
  return commands.ctx_edit(opts)
end

return M
