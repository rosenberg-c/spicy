-- Floating window utilities
-- Rule #3: Return values, not shared references

local M = {}

--- Calculate centered window position and size
--- @param width number|float Width (absolute or percentage 0-1)
--- @param height number|float Height (absolute or percentage 0-1)
--- @return table Window options {row, col, width, height}
local function calc_window_size(width, height)
  local screen_width = vim.o.columns
  local screen_height = vim.o.lines

  -- Convert percentage to absolute
  local win_width = width
  local win_height = height

  if width > 0 and width <= 1 then
    win_width = math.floor(screen_width * width)
  end

  if height > 0 and height <= 1 then
    win_height = math.floor(screen_height * height)
  end

  -- Calculate centered position
  local row = math.floor((screen_height - win_height) / 2)
  local col = math.floor((screen_width - win_width) / 2)

  return {
    row = row,
    col = col,
    width = win_width,
    height = win_height,
  }
end

--- Create a floating window
--- @param opts table Options:
---   - width: number Width (absolute or 0-1 for percentage)
---   - height: number Height (absolute or 0-1 for percentage)
---   - title: string Window title
---   - title_pos: string Title position (left, center, right)
---   - border: string Border style (none, single, double, rounded, solid, shadow)
---   - relative: string Relative to (editor, win, cursor)
---   - enter: boolean Enter window on creation
---   - content: table|nil Initial content lines
---   - syntax: string|nil Syntax highlighting
--- @return number, number bufnr, winid
function M.create(opts)
  opts = opts or {}

  -- Default options
  local width = opts.width or 0.8
  local height = opts.height or 0.6
  local border = opts.border or "rounded"
  local title = opts.title or ""
  local title_pos = opts.title_pos or "center"
  local relative = opts.relative or "editor"
  local enter = opts.enter
  if enter == nil then
    enter = true
  end

  -- Calculate window dimensions
  local size = calc_window_size(width, height)

  -- Create buffer
  local bufnr = vim.api.nvim_create_buf(false, true)
  vim.bo[bufnr].bufhidden = "wipe"

  -- Set buffer options
  if opts.syntax then
    vim.bo[bufnr].syntax = opts.syntax
  end

  -- Set content if provided
  if opts.content then
    vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, opts.content)
  end

  -- Configure window
  local win_opts = {
    relative = relative,
    row = size.row,
    col = size.col,
    width = size.width,
    height = size.height,
    style = "minimal",
    border = border,
  }

  -- Add title if provided
  if title ~= "" then
    win_opts.title = title
    win_opts.title_pos = title_pos
  end

  -- Create window
  local winid = vim.api.nvim_open_win(bufnr, enter, win_opts)

  -- Set window options
  vim.wo[winid].wrap = true
  vim.wo[winid].linebreak = true
  vim.wo[winid].cursorline = false

  -- Set up close keybindings
  local close_keys = { "q", "<Esc>" }
  for _, key in ipairs(close_keys) do
    vim.api.nvim_buf_set_keymap(
      bufnr,
      "n",
      key,
      ("<cmd>lua vim.api.nvim_win_close(%d, true)<CR>"):format(winid),
      { nowait = true, noremap = true, silent = true }
    )
  end

  return bufnr, winid
end

--- Update floating window content
--- @param bufnr number Buffer number
--- @param content table|string Content lines or string
function M.update_content(bufnr, content)
  if not vim.api.nvim_buf_is_valid(bufnr) then
    return
  end

  -- Convert string to lines
  local lines = content
  if type(content) == "string" then
    lines = vim.split(content, "\n", { plain = true })
  end

  -- Make buffer modifiable temporarily
  vim.bo[bufnr].modifiable = true

  -- Set content
  vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, lines)

  -- Make buffer non-modifiable again
  vim.bo[bufnr].modifiable = false
end

--- Append content to floating window
--- @param bufnr number Buffer number
--- @param content table|string Content lines or string
function M.append_content(bufnr, content)
  if not vim.api.nvim_buf_is_valid(bufnr) then
    return
  end

  -- Convert string to lines
  local lines = content
  if type(content) == "string" then
    lines = vim.split(content, "\n", { plain = true })
  end

  -- Make buffer modifiable temporarily
  vim.bo[bufnr].modifiable = true

  -- Get current line count
  local line_count = vim.api.nvim_buf_line_count(bufnr)

  -- Append content
  vim.api.nvim_buf_set_lines(bufnr, line_count, line_count, false, lines)

  -- Make buffer non-modifiable again
  vim.bo[bufnr].modifiable = false

  -- Scroll to bottom if window is valid
  local wins = vim.fn.win_findbuf(bufnr)
  for _, win in ipairs(wins) do
    if vim.api.nvim_win_is_valid(win) then
      local new_line_count = vim.api.nvim_buf_line_count(bufnr)
      vim.api.nvim_win_set_cursor(win, { new_line_count, 0 })
    end
  end
end

--- Close floating window
--- @param winid number Window ID
function M.close(winid)
  if winid and vim.api.nvim_win_is_valid(winid) then
    vim.api.nvim_win_close(winid, true)
  end
end

--- Check if floating window is open
--- @param winid number Window ID
--- @return boolean True if window is open
function M.is_open(winid)
  return winid and vim.api.nvim_win_is_valid(winid) or false
end

return M
