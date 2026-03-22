-- Commands module - exports all command implementations
-- Rule #15: Keep module boundaries clear

local M = {}

--- Ask a question
--- @param question string|nil Question to ask
--- @param opts table|nil Options
function M.ask(question, opts)
  local ask = require("spicy.commands.ask")
  return ask.ask(question, opts)
end

--- Ask about visual selection
--- @param opts table|nil Options
function M.ask_visual(opts)
  local ask = require("spicy.commands.ask")
  return ask.ask_visual(opts)
end

--- Generate a tutorial
--- @param topic string|nil Tutorial topic
--- @param opts table|nil Options
function M.tutor(topic, opts)
  local tutor = require("spicy.commands.tutor")
  return tutor.tutor(topic, opts)
end

--- Explain code
--- @param opts table|nil Options
function M.explain(opts)
  local explain = require("spicy.commands.explain")
  return explain.explain(opts)
end

--- Explain visual selection
--- @param opts table|nil Options
function M.explain_visual(opts)
  local explain = require("spicy.commands.explain")
  return explain.explain_visual(opts)
end

--- Generate git commit message
--- @param opts table|nil Options
function M.gitmessage(opts)
  local gitmessage = require("spicy.commands.gitmessage")
  return gitmessage.gitmessage(opts)
end

--- Update selection with ctx-edit
--- @param opts table|nil Options
function M.ctx_edit(opts)
  local ctx_edit = require("spicy.commands.ctx_edit")
  return ctx_edit.ctx_edit(opts)
end

--- Update visual selection with ctx-edit
--- @param opts table|nil Options
function M.ctx_edit_visual(opts)
  local ctx_edit = require("spicy.commands.ctx_edit")
  return ctx_edit.ctx_edit_visual(opts)
end

return M
