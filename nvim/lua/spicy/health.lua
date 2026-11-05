-- Health check for spicy.nvim
-- Rule #8: Document public APIs when needed

local M = {}

--- Check plugin health
--- Called by :checkhealth spicy
function M.check()
  -- Get health module (different APIs for nvim 0.9 vs 0.10+)
  local health = vim.health or require("health")

  health.start("spicy.nvim")

  -- Check Neovim version
  local nvim_version = vim.version()
  if nvim_version.major == 0 and nvim_version.minor < 8 then
    health.error(
      "Neovim version too old",
      "Requires Neovim 0.8+. Current: " .. vim.inspect(nvim_version)
    )
  else
    health.ok(("Neovim version: %s.%s.%s"):format(
      nvim_version.major,
      nvim_version.minor,
      nvim_version.patch
    ))
  end

  -- Check for plenary.nvim (required)
  local has_plenary = pcall(require, "plenary")
  if has_plenary then
    health.ok("plenary.nvim found")
  else
    health.error(
      "plenary.nvim not found (required)",
      {
        "Install plenary.nvim:",
        "https://github.com/nvim-lua/plenary.nvim",
      }
    )
  end

  -- Check for spicy CLI
  local job = require("spicy.utils.job")
  local config = require("spicy.config")

  local spicy_bin = config.get("bin.path") or "spicy"
  local spicy_path = job.command_path(spicy_bin)

  if spicy_path then
    health.ok(("spicy CLI found: %s"):format(spicy_path))
  else
    health.error(
      ("spicy CLI not found: %s"):format(spicy_bin),
      {
        "Install spicy CLI:",
        "https://github.com/user/spicy",
        "Or configure bin.path in setup()",
      }
    )
    return
  end

  -- Check individual commands
  local commands = { "ask", "tutor", "explain", "gitmessage" }
  for _, cmd in ipairs(commands) do
    local cmd_path = config.get("bin." .. cmd)
    if cmd_path then
      if job.command_exists(cmd_path) then
        health.ok(("%s command found: %s"):format(cmd, cmd_path))
      else
        health.warn(
          ("%s command not found: %s"):format(cmd, cmd_path),
          "Using default: " .. spicy_bin .. " " .. cmd
        )
      end
    else
      health.info(("%s using default binary"):format(cmd))
    end
  end

  -- Check optional dependencies
  local has_telescope = pcall(require, "telescope")
  if has_telescope then
    health.ok("telescope.nvim found (optional)")
  else
    health.info(
      "telescope.nvim not found (optional)",
      "Install for enhanced history browsing"
    )
  end

  local has_nui = pcall(require, "nui")
  if has_nui then
    health.ok("nui.nvim found (optional)")
  else
    health.info(
      "nui.nvim not found (optional)",
      "Install for enhanced UI components"
    )
  end

  -- Check configuration
  local cfg = config.get_all()
  local valid, err = config.validate(cfg)
  if valid then
    health.ok("Configuration is valid")
  else
    health.error(
      "Configuration is invalid",
      err
    )
  end

  -- Check directory permissions
  local history = config.get("history")
  if history and history.enabled and history.save_to_file then
    local fs = require("spicy.utils.fs")
    local history_dir = vim.fn.fnamemodify(history.file_path, ":h")

    if fs.exists(history_dir) then
      health.ok(("History directory exists: %s"):format(history_dir))
    else
      local ok, mkdir_err = fs.mkdir(history_dir)
      if ok then
        health.ok(("Created history directory: %s"):format(history_dir))
      else
        health.warn(
          ("Cannot create history directory: %s"):format(history_dir),
          mkdir_err
        )
      end
    end
  end

  -- Check tutorial save directory
  local tutor_dir = config.get("ui.tutor.save_dir")
  if tutor_dir then
    local fs = require("spicy.utils.fs")
    tutor_dir = fs.expand(tutor_dir)

    if fs.exists(tutor_dir) then
      health.ok(("Tutorial directory exists: %s"):format(tutor_dir))
    else
      health.info(
        ("Tutorial directory does not exist: %s"):format(tutor_dir),
        "Will be created when first tutorial is saved"
      )
    end
  end

  -- Check explain save directory
  local explain_dir = config.get("ui.explain.save_dir")
  if explain_dir then
    local fs = require("spicy.utils.fs")
    explain_dir = fs.expand(explain_dir)

    if fs.exists(explain_dir) then
      health.ok(("Explanation directory exists: %s"):format(explain_dir))
    else
      health.info(
        ("Explanation directory does not exist: %s"):format(explain_dir),
        "Will be created when first explanation is saved"
      )
    end
  end
end

return M
