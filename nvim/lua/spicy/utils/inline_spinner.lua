-- Inline spinner for buffer ranges
local M = {}

local INLINE_SPINNER_FRAMES = { "-", "\\", "|", "/" }
local inline_spinner_ns = vim.api.nvim_create_namespace("spicy.inline_spinner")

--- Start an inline spinner over a line range
--- @param bufnr number
--- @param start_line number
--- @param end_line number
--- @param message string|nil
--- @return table spinner handle
function M.start(bufnr, start_line, end_line, message)
  local frame = 1
  local timer = vim.loop.new_timer()
  local top_line = math.max(start_line - 1, 0)
  local bottom_line = math.max(end_line - 1, 0)
  local text_message = message or "Updating selection..."

  local function render()
    local glyph = INLINE_SPINNER_FRAMES[frame]
    local text = glyph .. " " .. text_message
    local opts = {
      virt_text = { { text, "Comment" } },
      virt_text_pos = "eol",
    }

    vim.api.nvim_buf_set_extmark(
      bufnr,
      inline_spinner_ns,
      top_line,
      0,
      opts
    )
    vim.api.nvim_buf_set_extmark(
      bufnr,
      inline_spinner_ns,
      bottom_line,
      0,
      opts
    )
  end

  local function tick()
    frame = (frame % #INLINE_SPINNER_FRAMES) + 1
    vim.api.nvim_buf_clear_namespace(bufnr, inline_spinner_ns, 0, -1)
    render()
  end

  vim.schedule(render)
  timer:start(0, 120, vim.schedule_wrap(tick))

  return {
    stop = function()
      timer:stop()
      timer:close()
      vim.api.nvim_buf_clear_namespace(bufnr, inline_spinner_ns, 0, -1)
    end,
  }
end

return M
