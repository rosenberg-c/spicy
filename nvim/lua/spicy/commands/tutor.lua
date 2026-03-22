-- SpicyTutor command implementation
-- Rule #5: Assume operations are slow and fragile
-- Rule #20: Return results and errors consistently

local M = {}

local job = require("spicy.utils.job")
local config = require("spicy.config")
local float = require("spicy.ui.float")
local input_ui = require("spicy.ui.input")
local spinner = require("spicy.ui.spinner")
local helpers = require("spicy.utils.helpers")
local fs = require("spicy.utils.fs")
local cli = require("spicy.utils.cli")

--- Build the spicy tutor command
--- @param topic string The tutorial topic
--- @param opts table|nil Options (model, verbose)
--- @return string, table Command and args
local function build_command(topic, opts)
  opts = opts or {}

  local bin = config.get("bin.tutor") or "tutor"
  local args = cli.base_args("tutor", opts, true)

  -- Add topic
  table.insert(args, topic)

  return bin, args
end

--- Generate tutorial
--- @param topic string The topic
--- @param opts table|nil Options
--- @param callback function|nil Callback(content, err)
function M.execute(topic, opts, callback)
  opts = opts or {}

  -- Build command
  local cmd, args = build_command(topic, opts)

  -- Check if command exists
  if not job.command_exists(cmd) then
    local err = ("Command not found: %s"):format(cmd)
    helpers.error(err)
    if callback then
      callback(nil, err)
    end
    return
  end

  -- Debug output
  if config.get("verbose") then
    helpers.info(("Running: %s %s"):format(cmd, table.concat(args, " ")))
  end

  -- Start spinner
  local spinner_id = spinner.start("Generating tutorial...")

  -- Run command
  job.run(cmd, args, {
    timeout = opts.timeout or config.get("timeout"),
    on_exit = function(stdout, stderr, code)
      -- Stop spinner
      if spinner_id then
        spinner.stop(spinner_id)
      end

      -- Check for errors
      if code ~= 0 then
        local err_msg = table.concat(stderr, "\n")
        helpers.error(
          ("Tutor command failed (exit code %d): %s"):format(code, err_msg)
        )
        if callback then
          callback(nil, err_msg)
        end
        return
      end

      -- Get content
      local content = table.concat(stdout, "\n")
      content = helpers.trim(content)

      if content == "" then
        helpers.warn("Received empty tutorial")
        if callback then
          callback(nil, "empty content")
        end
        return
      end

      -- Handle output
      vim.schedule(function()
        -- Save to file if auto_save enabled
        local auto_save = opts.auto_save
        if auto_save == nil then
          auto_save = config.get("ui.tutor.auto_save")
        end

        local filepath = nil
        if auto_save then
          local save_dir = config.get("ui.tutor.save_dir")
          save_dir = fs.expand(save_dir)

          -- Create directory if needed
          if not fs.exists(save_dir) then
            fs.mkdir(save_dir)
          end

          -- Generate filename from topic
          local filename = topic:gsub("%s+", "-"):lower() .. ".md"
          filepath = fs.join(save_dir, filename)

          local ok, err = fs.write_file(filepath, content)
          if ok then
            helpers.info(("Tutorial saved to: %s"):format(filepath))
          else
            helpers.error(("Failed to save: %s"):format(err))
          end
        end

        -- Display in buffer
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

        if callback then
          callback(content, nil)
        end
      end)
    end,
  })
end

--- Generate tutorial (main entry point)
--- @param topic string|nil The topic (nil to prompt)
--- @param opts table|nil Options
function M.tutor(topic, opts)
  opts = opts or {}

  -- If topic not provided, prompt for it
  if not topic or topic == "" then
    input_ui.prompt({
      prompt = "Tutorial topic: ",
      default = "",
    }, function(input)
      if not input or input == "" then
        helpers.info("Tutor cancelled")
        return
      end

      M.execute(input, opts, opts.on_complete)
    end)
    return
  end

  M.execute(topic, opts, opts.on_complete)
end

return M
