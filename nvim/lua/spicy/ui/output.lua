-- Output helpers for displaying content
local M = {}

local helpers = require("spicy.utils.helpers")

--- Show content in a scratch markdown buffer
--- @param content string
--- @return number bufnr
function M.show_markdown_buffer(content)
  local bufnr = helpers.create_scratch_buffer()
  vim.api.nvim_buf_set_lines(
    bufnr,
    0,
    -1,
    false,
    vim.split(content, "\n", { plain = true })
  )
  vim.bo[bufnr].filetype = "markdown"
  vim.cmd("buffer " .. bufnr)
  return bufnr
end

return M
