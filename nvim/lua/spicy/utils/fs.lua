-- Filesystem utilities
-- Rule #2: Build file paths with a helper
-- Rule #24: Prefer atomic write patterns

local M = {}

--- Join path components safely
--- Rule #2: Centralized path logic
--- @param ... string Path components to join
--- @return string The joined path
function M.join(...)
  local parts = { ... }
  local path = table.concat(parts, "/")

  -- Normalize: remove duplicate slashes
  path = path:gsub("//+", "/")

  return path
end

--- Expand tilde and environment variables in path
--- @param path string The path to expand
--- @return string The expanded path
function M.expand(path)
  return vim.fn.expand(path)
end

--- Check if a file exists
--- @param path string Path to check
--- @return boolean True if file exists
function M.exists(path)
  local stat = vim.loop.fs_stat(path)
  return stat ~= nil
end

--- Check if path is a directory
--- @param path string Path to check
--- @return boolean True if path is a directory
function M.is_dir(path)
  local stat = vim.loop.fs_stat(path)
  return stat and stat.type == "directory" or false
end

--- Create directory and any parent directories
--- @param path string Directory path to create
--- @return boolean, string|nil True on success, or false with error
function M.mkdir(path)
  path = M.expand(path)

  -- Rule #5: Always propagate errors with context
  local ok, err = pcall(vim.fn.mkdir, path, "p")
  if not ok then
    return false, ("failed to create directory %s: %s"):format(path, err)
  end

  return true, nil
end

--- Read file contents
--- @param path string Path to file
--- @return string|nil, string|nil Content on success, or nil with error
function M.read_file(path)
  path = M.expand(path)

  local fd, err = vim.loop.fs_open(path, "r", 438) -- 0666
  if not fd then
    return nil, ("failed to open %s: %s"):format(path, err)
  end

  local stat, err2 = vim.loop.fs_fstat(fd)
  if not stat then
    vim.loop.fs_close(fd)
    return nil, ("failed to stat %s: %s"):format(path, err2)
  end

  local data, err3 = vim.loop.fs_read(fd, stat.size, 0)
  vim.loop.fs_close(fd)

  if not data then
    return nil, ("failed to read %s: %s"):format(path, err3)
  end

  return data, nil
end

--- Write content to file atomically
--- Rule #24: Atomic write pattern (temp file + rename)
--- @param path string Destination file path
--- @param content string Content to write
--- @return boolean, string|nil True on success, or false with error
function M.write_file(path, content)
  path = M.expand(path)

  -- Create parent directory if needed
  local parent = vim.fn.fnamemodify(path, ":h")
  if not M.exists(parent) then
    local ok, err = M.mkdir(parent)
    if not ok then
      return false, err
    end
  end

  -- Write to temporary file first
  local tmp_path = path .. ".tmp"
  local fd, err = vim.loop.fs_open(tmp_path, "w", 438) -- 0666
  if not fd then
    return false, ("failed to open %s: %s"):format(tmp_path, err)
  end

  local ok2, err2 = vim.loop.fs_write(fd, content, 0)
  if not ok2 then
    vim.loop.fs_close(fd)
    vim.loop.fs_unlink(tmp_path)
    return false, ("failed to write %s: %s"):format(tmp_path, err2)
  end

  vim.loop.fs_close(fd)

  -- Atomic rename
  local ok3, err3 = vim.loop.fs_rename(tmp_path, path)
  if not ok3 then
    vim.loop.fs_unlink(tmp_path)
    return false, ("failed to rename %s to %s: %s"):format(tmp_path, path, err3)
  end

  return true, nil
end

--- Get file basename
--- @param path string File path
--- @return string The basename
function M.basename(path)
  return vim.fn.fnamemodify(path, ":t")
end

--- Get file extension
--- @param path string File path
--- @return string The extension (including dot)
function M.extension(path)
  return vim.fn.fnamemodify(path, ":e")
end

--- Remove file extension
--- @param path string File path
--- @return string Path without extension
function M.remove_extension(path)
  return vim.fn.fnamemodify(path, ":r")
end

return M
