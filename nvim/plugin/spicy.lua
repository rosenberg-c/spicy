-- Spicy.nvim plugin entry point
-- Keep this file minimal - only register commands
-- Defer actual loading to lua modules

-- Rule #26: Avoid hidden global state
if vim.g.loaded_spicy then
  return
end
vim.g.loaded_spicy = 1

-- Check Neovim version
if vim.fn.has("nvim-0.8") ~= 1 then
  vim.notify(
    "spicy.nvim requires Neovim 0.8 or higher",
    vim.log.levels.ERROR
  )
  return
end

-- SpicyAsk command
vim.api.nvim_create_user_command("SpicyAsk", function(opts)
  -- Defer loading until command is used
  local spicy = require("spicy")

  -- Handle range for visual selection
  if opts.range > 0 then
    spicy.ask_visual()
  else
    -- Join args as question
    local question = table.concat(opts.fargs, " ")
    spicy.ask(question)
  end
end, {
  nargs = "*",
  range = true,
  desc = "Ask a question using spicy AI",
})

-- SpicyTutor command
vim.api.nvim_create_user_command("SpicyTutor", function(opts)
  local spicy = require("spicy")
  local topic = table.concat(opts.fargs, " ")
  spicy.tutor(topic)
end, {
  nargs = "*",
  desc = "Generate a tutorial using spicy AI",
})

-- SpicyExplain command
vim.api.nvim_create_user_command("SpicyExplain", function(opts)
  local commands = require("spicy.commands")

  local options = {}

  -- Handle range/visual selection
  if opts.range > 0 then
    options.range = {
      start_line = opts.line1,
      end_line = opts.line2,
    }
  end

  if options.range then
    commands.explain(options)
  else
    commands.explain_visual(options)
  end
end, {
  nargs = "*",
  range = true,
  desc = "Explain code using spicy AI",
})

-- SpicyGitmessage command
vim.api.nvim_create_user_command("SpicyGitmessage", function(opts)
  local spicy = require("spicy")

  local options = {}

  -- First arg might be prefix (feat, fix, etc.)
  if #opts.fargs > 0 then
    options.prefix = opts.fargs[1]
  end

  -- Rest is hint
  if #opts.fargs > 1 then
    options.hint = table.concat(
      { unpack(opts.fargs, 2) },
      " "
    )
  end

  spicy.gitmessage(options)
end, {
  nargs = "*",
  desc = "Generate git commit message using spicy AI",
})


-- SpicyCtxEdit command
vim.api.nvim_create_user_command("SpicyCtxEdit", function(opts)
  local commands = require("spicy.commands")

  local options = {}

  if opts.range > 0 then
    options.range = {
      start_line = opts.line1,
      end_line = opts.line2,
    }
  end

  if options.range then
    commands.ctx_edit(options)
  else
    commands.ctx_edit_visual(options)
  end
end, {
  nargs = "*",
  range = true,
  desc = "Edit selection using spicy AI",
})
