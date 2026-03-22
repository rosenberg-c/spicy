-- Async job management using plenary.job
-- Rule #5: Assume operations are slow and fragile
-- Rule #7: Dependency injection for testability

local M = {}

-- Active jobs registry for cancellation
-- Rule #26: Avoid hidden global state - but this is module-scoped
M.active_jobs = {}

--- Run a command asynchronously
--- Rule #20: Return results and errors consistently
--- @param cmd string Command to run
--- @param args table Command arguments
--- @param opts table|nil Options:
---   - on_stdout: function(err, data) Called with stdout data
---   - on_stderr: function(err, data) Called with stderr data
---   - on_exit: function(stdout, stderr, return_code) Called on completion
---   - on_timeout: function() Called if timeout occurs
---   - timeout: number Timeout in milliseconds
---   - cwd: string Working directory
--- @return table|nil, string|nil Job object or nil with error
function M.run(cmd, args, opts)
  opts = opts or {}

  -- Check if plenary is available
  local has_plenary, Job = pcall(require, "plenary.job")
  if not has_plenary then
    return nil, "plenary.nvim is required but not installed"
  end

  -- Build job configuration
  local job_opts = {
    command = cmd,
    args = args,
    cwd = opts.cwd or vim.fn.getcwd(),
    on_stdout = opts.on_stdout,
    on_stderr = opts.on_stderr,
    on_exit = function(j, return_code)
      -- Remove from active jobs
      M.active_jobs[j] = nil

      if opts.on_exit then
        -- Rule #5: Propagate errors with context
        local stdout = j:result()
        local stderr = j:stderr_result()
        opts.on_exit(stdout, stderr, return_code)
      end
    end,
  }

  -- Create and start job
  local job = Job:new(job_opts)

  -- Track active job
  M.active_jobs[job] = true

  -- Start the job
  job:start()

  -- Handle timeout if specified
  if opts.timeout and opts.timeout > 0 then
    vim.defer_fn(function()
      -- Rule #18: Distinguish nil and false
      if M.active_jobs[job] then
        job:shutdown()
        M.active_jobs[job] = nil

        if opts.on_timeout then
          opts.on_timeout()
        end
      end
    end, opts.timeout)
  end

  return job, nil
end

--- Run a command and wait for result (blocking)
--- Use sparingly - prefer async version
--- @param cmd string Command to run
--- @param args table Command arguments
--- @param opts table|nil Options (same as run)
--- @return table|nil, table|nil, number|nil stdout, stderr, return_code
function M.run_sync(cmd, args, opts)
  opts = opts or {}

  local has_plenary, Job = pcall(require, "plenary.job")
  if not has_plenary then
    return nil, { "plenary.nvim is required but not installed" }, 1
  end

  local job = Job:new({
    command = cmd,
    args = args,
    cwd = opts.cwd or vim.fn.getcwd(),
  })

  -- Run synchronously with timeout
  local timeout = opts.timeout or 60000
  job:sync(timeout)

  local stdout = job:result()
  local stderr = job:stderr_result()
  local code = job.code

  return stdout, stderr, code
end

--- Cancel a specific job
--- @param job table The job to cancel
function M.cancel(job)
  if job and M.active_jobs[job] then
    job:shutdown()
    M.active_jobs[job] = nil
  end
end

--- Cancel all active jobs
function M.cancel_all()
  for job, _ in pairs(M.active_jobs) do
    if not job.is_shutdown then
      job:shutdown()
    end
  end
  M.active_jobs = {}
end

--- Get count of active jobs
--- @return number Number of active jobs
function M.count_active()
  local count = 0
  for _, _ in pairs(M.active_jobs) do
    count = count + 1
  end
  return count
end

--- Check if a command exists in PATH
--- @param cmd string Command name
--- @return boolean True if command exists
function M.command_exists(cmd)
  return vim.fn.executable(cmd) == 1
end

--- Get the full path of a command
--- @param cmd string Command name
--- @return string|nil Full path or nil if not found
function M.command_path(cmd)
  local path = vim.fn.exepath(cmd)
  if path == "" then
    return nil
  end
  return path
end

return M
