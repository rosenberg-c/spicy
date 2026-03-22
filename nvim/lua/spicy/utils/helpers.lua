-- General utility functions
-- Rule #15: Avoid large generic utils.lua - but these are truly generic

local M = {}

--- Get visual selection text
--- @return string|nil, number|nil, number|nil text, start_line, end_line
function M.get_visual_selection()
  -- Get visual selection marks
  local start_pos = vim.fn.getpos("'<")
  local end_pos = vim.fn.getpos("'>")

  local start_line = start_pos[2]
  local end_line = end_pos[2]

  if start_line == 0 or end_line == 0 then
    return nil, nil, nil
  end

  -- Get lines in selection
  local lines = vim.api.nvim_buf_get_lines(
    0,
    start_line - 1,
    end_line,
    false
  )

  if #lines == 0 then
    return nil, nil, nil
  end

  return table.concat(lines, "\n"), start_line, end_line
end

--- Get current buffer content
--- @param bufnr number|nil Buffer number (0 or nil for current)
--- @return string Buffer content
function M.get_buffer_content(bufnr)
  bufnr = bufnr or 0
  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  return table.concat(lines, "\n")
end

--- Get lines from buffer in range
--- @param bufnr number|nil Buffer number (0 or nil for current)
--- @param start_line number Start line (1-indexed)
--- @param end_line number End line (1-indexed)
--- @return string Content of lines in range
function M.get_lines_in_range(bufnr, start_line, end_line)
  bufnr = bufnr or 0
  local lines = vim.api.nvim_buf_get_lines(
    bufnr,
    start_line - 1,
    end_line,
    false
  )
  return table.concat(lines, "\n")
end

--- Get current filetype
--- @return string Filetype
function M.get_filetype()
  return vim.bo.filetype
end

--- Get current filename
--- @return string Filename
function M.get_filename()
  return vim.fn.expand("%:t")
end

--- Get current file path
--- @return string Full file path
function M.get_filepath()
  return vim.fn.expand("%:p")
end

--- Check if buffer is empty
--- @param bufnr number|nil Buffer number (0 or nil for current)
--- @return boolean True if buffer is empty
function M.is_buffer_empty(bufnr)
  bufnr = bufnr or 0
  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  return #lines == 0 or (#lines == 1 and lines[1] == "")
end

--- Trim whitespace from string
--- @param str string String to trim
--- @return string Trimmed string
function M.trim(str)
  return str:match("^%s*(.-)%s*$")
end

--- Split string by delimiter
--- @param str string String to split
--- @param delimiter string Delimiter
--- @return table Array of parts
function M.split(str, delimiter)
  return vim.split(str, delimiter, { plain = true })
end

--- Check if string starts with prefix
--- @param str string String to check
--- @param prefix string Prefix to check for
--- @return boolean True if string starts with prefix
function M.starts_with(str, prefix)
  return str:sub(1, #prefix) == prefix
end

--- Check if string ends with suffix
--- @param str string String to check
--- @param suffix string Suffix to check for
--- @return boolean True if string ends with suffix
function M.ends_with(str, suffix)
  return str:sub(-#suffix) == suffix
end

--- Notify user with message
--- @param msg string Message to display
--- @param level string|nil Log level (info, warn, error)
function M.notify(msg, level)
  level = level or "info"

  local levels = {
    info = vim.log.levels.INFO,
    warn = vim.log.levels.WARN,
    error = vim.log.levels.ERROR,
  }

  vim.notify(msg, levels[level] or vim.log.levels.INFO, {
    title = "Spicy",
  })
end

--- Show error message
--- @param msg string Error message
function M.error(msg)
  M.notify(msg, "error")
end

--- Show warning message
--- @param msg string Warning message
function M.warn(msg)
  M.notify(msg, "warn")
end

--- Show info message
--- @param msg string Info message
function M.info(msg)
  M.notify(msg, "info")
end

--- Create a scratch buffer
--- @return number Buffer number
function M.create_scratch_buffer()
  local bufnr = vim.api.nvim_create_buf(false, true)
  vim.bo[bufnr].bufhidden = "wipe"
  return bufnr
end

--- Debounce a function
--- Rule #1: Never rely on loop variable identity
--- @param func function Function to debounce
--- @param delay number Delay in milliseconds
--- @return function Debounced function
function M.debounce(func, delay)
  local timer = nil

  return function(...)
    local args = { ... }

    if timer then
      timer:stop()
      timer:close()
    end

    timer = vim.defer_fn(function()
      func(unpack(args))
    end, delay)
  end
end

return M
