-- Loading spinner/indicator utilities
-- Rule #26: Module-scoped state for spinners

local M = {}

-- Active spinners
local active_spinners = {}
local spinner_id_counter = 0

local constants = require("spicy.constants")

local function resolve_frames(style)
  if style == "slash" then
    return constants.spinner_frames_slash
  end

  return constants.spinner_frames_braille
end

--- Start a spinner
--- @param message string Message to display
--- @param opts table|nil Options:
---   - position: string Position (statusline, float, cmdline)
---   - style: string Spinner style (braille, slash)
--- @return number spinner_id
function M.start(message, opts)
  opts = opts or {}

  spinner_id_counter = spinner_id_counter + 1
  local id = spinner_id_counter

  local frames = resolve_frames(opts.style)
  local spinner = {
    id = id,
    message = message or "Loading...",
    frame = 1,
    position = opts.position or "cmdline",
    timer = nil,
    frames = frames,
  }

  -- Start animation timer
  spinner.timer = vim.loop.new_timer()
  spinner.timer:start(
    0,
    100, -- Update every 100ms
    vim.schedule_wrap(function()
      if not active_spinners[id] then
        return
      end

      -- Update frame
      spinner.frame = (spinner.frame % #spinner.frames) + 1
      local frame = spinner.frames[spinner.frame]

      -- Display based on position
      if spinner.position == "cmdline" then
        -- Already wrapped in vim.schedule_wrap, safe to call
        vim.api.nvim_echo(
          { { frame .. " " .. spinner.message, "Normal" } },
          false,
          {}
        )
      end
      -- TODO: Add float and statusline positions
    end)
  )

  active_spinners[id] = spinner
  return id
end

--- Update spinner message
--- @param id number Spinner ID
--- @param message string New message
function M.update_message(id, message)
  local spinner = active_spinners[id]
  if spinner then
    spinner.message = message
  end
end

--- Stop a spinner
--- @param id number Spinner ID
function M.stop(id)
  local spinner = active_spinners[id]
  if not spinner then
    return
  end

  -- Stop timer
  if spinner.timer then
    spinner.timer:stop()
    spinner.timer:close()
  end

  -- Clear display (must use vim.schedule in callbacks)
  if spinner.position == "cmdline" then
    vim.schedule(function()
      vim.api.nvim_echo({ { "", "Normal" } }, false, {})
    end)
  end

  active_spinners[id] = nil
end

--- Stop all spinners
function M.stop_all()
  for id, _ in pairs(active_spinners) do
    M.stop(id)
  end
end

return M
