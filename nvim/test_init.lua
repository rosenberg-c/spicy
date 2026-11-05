-- Minimal test config for spicy.nvim
-- Usage: nvim -u test_init.lua

-- Add plugin to runtimepath
vim.opt.rtp:append(vim.fn.getcwd())

-- Install plenary if not already present
local plenary_path = vim.fn.stdpath("data") .. "/lazy/plenary.nvim"
if vim.fn.isdirectory(plenary_path) == 0 then
  print("Installing plenary.nvim...")
  vim.fn.system({
    "git",
    "clone",
    "--depth=1",
    "https://github.com/nvim-lua/plenary.nvim",
    plenary_path,
  })
end
vim.opt.rtp:append(plenary_path)

-- Load spicy
require("spicy").setup({
  -- Test config
  models = {
    ask = "openai/gpt-5.3-codex",
  },
  keymaps = {
    enabled = true,
  },
})

-- Print status
print("Spicy.nvim loaded!")
print("Try: :SpicyAsk what is a closure")
print("Or: :checkhealth spicy")
