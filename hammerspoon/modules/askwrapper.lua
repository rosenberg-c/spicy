local M = {}

local historyDir = os.getenv("HOME") .. "/.askwrapper"
local historyPath = historyDir .. "/history.json"

local function shellQuote(s)
	return "'" .. tostring(s):gsub("'", "'\\''") .. "'"
end

local function ensureHistoryDir()
	if not hs.fs.attributes(historyDir) then
		hs.fs.mkdir(historyDir)
	end
end

local function loadHistory()
	ensureHistoryDir()

	local f = io.open(historyPath, "r")
	if not f then
		return {}
	end

	local content = f:read("*a")
	f:close()
	if not content or content == "" then
		return {}
	end

	local data = hs.json.decode(content)
	if type(data) ~= "table" then
		return {}
	end

	return data
end

local function saveHistory(history)
	ensureHistoryDir()
	local content = hs.json.encode(history, true) or "[]"
	local f = io.open(historyPath, "w")
	if not f then
		return
	end
	if content:sub(-1) ~= "\n" then
		content = content .. "\n"
	end
	f:write(content)
	f:close()
end

local function appendHistoryEntry(question, answer)
	if not question or question == "" then
		return
	end

	local history = loadHistory()
	local entry = {
		question = tostring(question),
		answer = tostring(answer or ""),
		at = os.time(),
	}

	table.insert(history, 1, entry)
	saveHistory(history)
end

-- Terminal version (iTerm)
local askChooser = nil

local function runAskInITerm(query)
	if not query or query == "" then
		return
	end

	local escaped = shellQuote(query)

	local applescript = string.format(
		[[
    tell application "iTerm"
      activate
      create window with default profile
      tell current session of current window
        write text "clear && ask %s"
      end tell
    end tell
  ]],
		escaped
	)

	hs.osascript.applescript(applescript)
end

local function createAskChooser(onSubmit, subText)
	local chooser = hs.chooser.new(function(choice)
		if choice and choice.text and choice.text ~= "" then
			onSubmit(choice.text)
		end
	end)

	chooser:placeholderText("Ask…")
	chooser:searchSubText(false)
	chooser:choices({})
	chooser:queryChangedCallback(function(q)
		chooser:choices({
			{
				text = q or "",
				subText = subText,
			},
		})
	end)

	return chooser
end

local function createAskChooserWithHistory(onSubmitNew, onSubmitHistory, subText)
	local historyState = {
		count = 0,
		lastQuery = "",
	}

	local function buildChoices(q)
		local choices = {
			{
				text = q or "",
				subText = subText,
				kind = "input",
			},
		}

		local history = loadHistory()
		for _, item in ipairs(history) do
			local preview = tostring(item.answer or "")
			preview = preview:gsub("%s+", " ")
			if #preview > 160 then
				preview = preview:sub(1, 160) .. "..."
			end
			table.insert(choices, {
				text = tostring(item.question or ""),
				subText = preview,
				kind = "history",
				question = tostring(item.question or ""),
				answer = tostring(item.answer or ""),
			})
		end

		historyState.count = #history
		return choices
	end

	local chooser = hs.chooser.new(function(choice)
		if not choice then
			return
		end

		local query = historyState.lastQuery or ""
		if query ~= "" then
			onSubmitNew(query)
			return
		end

		if choice.kind == "history" then
			onSubmitHistory(choice)
		end
	end)
	local height = 16

	chooser:placeholderText("Ask…")
	chooser:searchSubText(false)
	local initialChoices = buildChoices("")
	chooser:choices(initialChoices)
	chooser:rows(height)
	chooser:queryChangedCallback(function(q)
		historyState.lastQuery = q or ""
		local choices = buildChoices(q)
		chooser:choices(choices)
		chooser:rows(height)
	end)

	local tap = hs.eventtap.new({ hs.eventtap.event.types.keyDown }, function(e)
		if not chooser:isVisible() then
			return false
		end

		local keyCode = e:getKeyCode()
		if keyCode ~= 125 and keyCode ~= 126 and keyCode ~= 51 and keyCode ~= 117 then
			return false
		end

		if historyState.count == 0 then
			return false
		end

		if keyCode == 51 or keyCode == 117 then
			if (historyState.lastQuery or "") ~= "" then
				return false
			end
			local selected = chooser:selectedRow()
			if not selected or selected < 2 then
				return false
			end
			local rowChoice = chooser:selectedRowContents(selected)
			if not rowChoice or rowChoice.kind ~= "history" then
				return false
			end
			local history = loadHistory()
			local index = selected - 1
			if history[index] then
				table.remove(history, index)
				saveHistory(history)
			end
			local choices = buildChoices(chooser:query() or "")
			chooser:choices(choices)
			local nextRow = math.min(selected, historyState.count + 1)
			chooser:selectedRow(nextRow)
			return true
		end

		local current = chooser:selectedRow()
		local target = current
		if keyCode == 125 then
			if current < 2 then
				target = 2
			else
				target = math.min(current + 1, historyState.count + 1)
			end
		else
			if current <= 2 then
				target = 2
			else
				target = current - 1
			end
		end

		historyState.lastQuery = ""
		chooser:query("")
		hs.timer.doAfter(0, function()
			chooser:selectedRow(target)
		end)
		return true
	end)

	chooser:showCallback(function()
		tap:start()
		local choices = buildChoices(chooser:query() or "")
		chooser:choices(choices)
		chooser:rows(height)
	end)
	chooser:hideCallback(function()
		tap:stop()
	end)

	return chooser
end

local function showAskBox()
	if not askChooser then
		askChooser = createAskChooser(runAskInITerm, "Press Enter to run ask in a new iTerm window. Esc to cancel.")
	end

	askChooser:query("")
	askChooser:show()
end

local function setupAskInITerm()
	hs.hotkey.bind({ "alt", "shift" }, "A", showAskBox)
end

-- Sublime version
local spinnerCanvas = nil
local spinnerTimer = nil
local spinnerFrames = { "⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏" }

local function createSpinnerCanvas()
	local screen = hs.screen.mainScreen()
	if not screen then
		return nil
	end

	local frame = screen:frame()
	local w, h = 160, 60
	local x = frame.x + (frame.w - w) / 2
	local y = frame.y + (frame.h - h) / 2

	local canvas = hs.canvas.new({ x = x, y = y, w = w, h = h })
	canvas:level("modalPanel")
	canvas:behavior("canJoinAllSpaces")

	canvas[1] = {
		type = "rectangle",
		action = "fill",
		fillColor = { red = 0.35, green = 0.05, blue = 0.05, alpha = 0.85 },
		roundedRectRadii = { xRadius = 10, yRadius = 10 },
	}

	canvas[2] = {
		type = "text",
		text = "ask ⠋",
		textSize = 22,
		textColor = { white = 1, alpha = 1 },
		textAlignment = "center",
	}

	canvas:show()
	return canvas
end

local function startSpinner()
	if spinnerTimer then
		return
	end
	local i = 1

	if spinnerCanvas then
		spinnerCanvas:delete()
		spinnerCanvas = nil
	end

	spinnerCanvas = createSpinnerCanvas()
	if not spinnerCanvas then
		return
	end

	local function updateSpinnerText()
		spinnerCanvas[2].text = spinnerFrames[i] .. " ask " .. spinnerFrames[i]
	end

	updateSpinnerText()

	spinnerTimer = hs.timer.doEvery(0.1, function()
		if spinnerCanvas then
			updateSpinnerText()
		end
		i = i % #spinnerFrames + 1
	end)
end

local function stopSpinner()
	if spinnerTimer then
		spinnerTimer:stop()
		spinnerTimer = nil
	end

	if spinnerCanvas then
		spinnerCanvas:delete()
		spinnerCanvas = nil
	end
end

local function findSubl()
	local candidates = {
		"/opt/homebrew/bin/subl",
		"/usr/local/bin/subl",
		"/Applications/Sublime Text.app/Contents/SharedSupport/bin/subl",
	}

	for _, path in ipairs(candidates) do
		if hs.fs.attributes(path) then
			return path
		end
	end

	return nil
end

local function openInSublime(content)
	local subl = findSubl()
	if not subl then
		hs.alert.show("Could not find subl")
		stopSpinner()
		return
	end

	local tmp = os.tmpname() .. ".txt"
	local f = io.open(tmp, "w")
	if not f then
		hs.alert.show("Could not create temp file")
		stopSpinner()
		return
	end

	f:write(content)
	f:close()

	hs.execute(shellQuote(subl) .. " " .. shellQuote(tmp))
end

local function runAskToSublime(query)
	if not query or query == "" then
		return
	end

	startSpinner()

	local marker = "__ASK_OUTPUT_START__"
	local shellCmd = "PATH=$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH; printf '%s\\n' "
		.. shellQuote(marker)
		.. "; ask "
		.. shellQuote(query)

	local task = hs.task.new("/bin/zsh", function(exitCode, stdOut, stdErr)
		stopSpinner()

		if exitCode ~= 0 then
			hs.alert.show("ask failed (" .. tostring(exitCode) .. ")")
			if stdErr and stdErr ~= "" then
				print(stdErr)
			elseif stdOut and stdOut ~= "" then
				print(stdOut)
			end
			return
		end

		local output = stdOut
		if output and output ~= "" then
			output = output:gsub("\r\n", "\n")

			local markerPos = output:find(marker, 1, true)
			if markerPos then
				local afterMarker = output:find("\n", markerPos, true)
				if afterMarker then
					output = output:sub(afterMarker + 1)
				else
					output = ""
				end
			end

			output = output:gsub("^%s+", "")
		end

		if not output or output == "" then
			output = "(ask returned no output)\n"
		end

		appendHistoryEntry(query, output)
		openInSublime(output)
	end, { "-lc", shellCmd })

	if task then
		hs.timer.doAfter(0, function()
			task:start()
		end)
	else
		stopSpinner()
		hs.alert.show("Could not start ask task")
	end
end

local function setupAskToSublime()
	local askToSublimeChooser = createAskChooserWithHistory(
		runAskToSublime,
		function(choice)
			if choice and choice.answer and choice.answer ~= "" then
				openInSublime(choice.answer)
			end
		end,
		"Press Enter to run ask and open the response in Sublime. Esc to cancel."
	)

	hs.hotkey.bind({ "alt", "shift" }, "S", function()
		askToSublimeChooser:query("")
		askToSublimeChooser:show()
	end)
end

function M.setup()
	setupAskInITerm()
	setupAskToSublime()
end

return M
