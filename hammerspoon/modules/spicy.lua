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

local function resolveBinaryPath(binaryName)
	local candidates = {
		os.getenv("HOME") .. "/.local/bin/" .. binaryName,
		"/opt/homebrew/bin/" .. binaryName,
		"/usr/local/bin/" .. binaryName,
	}

	for _, path in ipairs(candidates) do
		if hs.fs.attributes(path) then
			return path
		end
	end

	local out = hs.execute("PATH=$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH; command -v " .. binaryName)
	out = trim(out)
	if out ~= "" and hs.fs.attributes(out) then
		return out
	end

	return nil
end

local function launchBinary(binaryName, args)
	local exe = resolveBinaryPath(binaryName)
	if not exe then
		hs.alert.show(binaryName .. " not found in PATH")
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
			hs.alert.show(binaryName .. " exited (" .. tostring(exitCode) .. ")")
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
		hs.alert.show("Could not launch " .. binaryName)
		return
	end

	uiTasks[task] = true
	if not task:start() then
		uiTasks[task] = nil
		hs.alert.show("Failed to start " .. binaryName)
	end
end

local function launchAskwrapper(args)
	launchBinary("askwrapper", args)
end

local function launchImgwalker()
	launchBinary("imgwalker")
end

function M.setup()
	hs.hotkey.bind({ "alt", "shift" }, "A", function()
		launchAskwrapper({ "ui", "ask" })
	end)

	hs.hotkey.bind({ "alt", "shift" }, "S", function()
		launchAskwrapper({ "ui", "followup" })
	end)

	hs.hotkey.bind({ "alt", "shift" }, "D", function()
		launchImgwalker()
	end)
end

return M
