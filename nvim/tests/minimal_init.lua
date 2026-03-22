-- Minimal init for running tests
-- This sets up the test environment without loading user config

-- Add current directory to runtimepath
vim.opt.rtp:append(".")

-- Add plenary to runtimepath
local plenary_dir = vim.fn.stdpath("data") .. "/plenary.nvim"
if vim.fn.isdirectory(plenary_dir) == 0 then
  print("Installing plenary.nvim for testing...")
  vim.fn.system({
    "git",
    "clone",
    "--depth=1",
    "https://github.com/nvim-lua/plenary.nvim",
    plenary_dir,
  })
end
vim.opt.rtp:append(plenary_dir)

-- Load plenary
vim.cmd("runtime plugin/plenary.vim")

-- Disable swap files and backup
vim.opt.swapfile = false
vim.opt.backup = false
vim.opt.writebackup = false

-- Set test-specific settings
vim.g.loaded_spicy = nil -- Allow reloading
