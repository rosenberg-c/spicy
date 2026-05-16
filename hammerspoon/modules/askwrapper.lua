local M = {}

local uiTasks = {}

local function shellQuote(s)
	return "'" .. tostring(s):gsub("'", "'\\''") .. "'"
end

local function trim(s)
	local out = tostring(s or "")
	out = out:gsub("^%s+", "")
	out = out:gsub("%s+$", "")
	return out
end

local function resolveAskwrapperPath()
	local candidates = {
		os.getenv("HOME") .. "/.local/bin/askwrapper",
		"/opt/homebrew/bin/askwrapper",
		"/usr/local/bin/askwrapper",
	}

	for _, path in ipairs(candidates) do
		if hs.fs.attributes(path) then
			return path
		end
	end

	local out = hs.execute("PATH=$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH; command -v askwrapper")
	out = trim(out)
	if out ~= "" and hs.fs.attributes(out) then
		return out
	end

	return nil
end

local function launchAskwrapper(args)
	local exe = resolveAskwrapperPath()
	if not exe then
		hs.alert.show("askwrapper not found in PATH")
		return
	end

	local cmdParts = { shellQuote(exe) }
	for _, arg in ipairs(args or {}) do
		table.insert(cmdParts, shellQuote(arg))
	end
	local shellCmd = "export PATH=$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH; " .. table.concat(cmdParts, " ")

	local task = nil
	task = hs.task.new("/bin/zsh", function(exitCode, stdOut, stdErr)
		uiTasks[task] = nil
		if exitCode ~= 0 then
			hs.alert.show("askwrapper exited (" .. tostring(exitCode) .. ")")
			local errText = trim(stdErr or "")
			local outText = trim(stdOut or "")
			if errText ~= "" then
				print(errText)
			elseif outText ~= "" then
				print(outText)
			end
		end
	end, { "-lc", shellCmd })

	if not task then
		hs.alert.show("Could not launch askwrapper")
		return
	end

	uiTasks[task] = true
	if not task:start() then
		uiTasks[task] = nil
		hs.alert.show("Failed to start askwrapper")
	end
end

function M.setup()
	hs.hotkey.bind({ "alt", "shift" }, "A", function()
		launchAskwrapper({ "ui", "ask" })
	end)

	hs.hotkey.bind({ "alt", "shift" }, "S", function()
		launchAskwrapper({ "ui", "followup" })
	end)
end

return M
