-- Input prompt utilities

local M = {}

--- Prompt user for input
--- @param opts table Options:
---   - prompt: string Prompt message
---   - default: string Default value
---   - completion: string Completion type (file, dir, etc.)
--- @param callback function(input) Called with user input or nil if cancelled
function M.prompt(opts, callback)
  opts = opts or {}

  local prompt_text = opts.prompt or "Input: "
  local default_value = opts.default or ""
  local completion = opts.completion

  vim.ui.input({
    prompt = prompt_text,
    default = default_value,
    completion = completion,
  }, function(input)
    if input == nil then
      -- User cancelled
      callback(nil)
    else
      callback(input)
    end
  end)
end

--- Prompt for yes/no confirmation
--- @param message string Confirmation message
--- @param callback function(confirmed) Called with boolean result
function M.confirm(message, callback)
  vim.ui.select(
    { "Yes", "No" },
    {
      prompt = message,
    },
    function(choice)
      callback(choice == "Yes")
    end
  )
end

--- Prompt for selection from list
--- @param items table List of items
--- @param opts table Options:
---   - prompt: string Prompt message
---   - format_item: function(item) Format function for display
--- @param callback function(item, idx) Called with selected item and index
function M.select(items, opts, callback)
  opts = opts or {}

  vim.ui.select(items, {
    prompt = opts.prompt or "Select: ",
    format_item = opts.format_item,
  }, function(item, idx)
    callback(item, idx)
  end)
end

return M
