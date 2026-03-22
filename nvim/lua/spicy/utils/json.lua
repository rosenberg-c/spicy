-- JSON decoding helpers
local M = {}

local helpers = require("spicy.utils.helpers")

--- Decode JSON with fallback strategies for noisy outputs
--- @param output string
--- @return table|nil, string|nil decoded, err
function M.decode_loose(output)
  output = helpers.trim(output or "")
  if output == "" then
    return nil, "empty response"
  end

  local ok, decoded = pcall(vim.fn.json_decode, output)
  if not ok or type(decoded) ~= "table" then
    if type(vim.json) == "table" and type(vim.json.decode) == "function" then
      local json_ok, json_decoded = pcall(vim.json.decode, output)
      if json_ok and type(json_decoded) == "table" then
        decoded = json_decoded
        ok = true
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    local lines = vim.split(output, "\n", { plain = true })
    for i = #lines, 1, -1 do
      local line = helpers.trim(lines[i])
      if line ~= "" then
        local line_ok, line_decoded = pcall(vim.fn.json_decode, line)
        if line_ok then
          if type(line_decoded) == "table" then
            decoded = line_decoded
            ok = true
            break
          end
          if type(line_decoded) == "string" then
            local nested_ok, nested_decoded = pcall(vim.fn.json_decode, line_decoded)
            if nested_ok and type(nested_decoded) == "table" then
              decoded = nested_decoded
              ok = true
              break
            end
          end
        end
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    local first = output:find("{", 1, true)
    local last = nil
    for i = #output, 1, -1 do
      if output:sub(i, i) == "}" then
        last = i
        break
      end
    end
    if first and last and last > first then
      local candidate = output:sub(first, last)
      local cand_ok, cand_decoded = pcall(vim.fn.json_decode, candidate)
      if cand_ok and type(cand_decoded) == "table" then
        decoded = cand_decoded
        ok = true
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    local stripped = helpers.trim(output)
    local quote = stripped:sub(1, 1)
    if (quote == '"' or quote == "'") and stripped:sub(-1) == quote then
      local unquoted = stripped:sub(2, -2)
      local uq_ok, uq_decoded = pcall(vim.fn.json_decode, unquoted)
      if uq_ok and type(uq_decoded) == "table" then
        decoded = uq_decoded
        ok = true
      end
    end
  end

  if not ok or type(decoded) ~= "table" then
    return nil, "invalid json response"
  end

  return decoded, nil
end

return M
