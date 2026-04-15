-- Example configuration for spicy.nvim
-- Copy this to your init.lua or init.vim and customize to your needs

-- Setup spicy.nvim
require("spicy").setup({
	-- CLI binary configuration
	bin = {
		ask = "ask",
		tutor = "tutor",
		explain = "explain",
		gitmessage = "gitmessage",
		ctx_edit = "v-edit",
	},

	-- Default models for each command
	models = {
		ask = "openai/gpt-5.3-codex",
		tutor = "openai/gpt-5.3-codex",
		explain = "openai/gpt-5.3-codex",
		gitmessage = "openai/gpt-5.3-codex",
		ctx_edit = "openai/gpt-5.3-codex",
	},

	-- UI configuration per command
	ui = {
		ask = {
			output = "float",
			float_opts = {
				relative = "editor",
				width = 0.8,
				height = 0.6,
				border = "rounded",
				title = " Spicy Ask ",
				title_pos = "center",
			},
			syntax = "markdown",
			show_spinner = true,
			auto_close = false,
		},

		tutor = {
			output = "buffer",
			auto_save = true,
			-- save_dir = vim.fn.expand("~/tutorials"),
			-- Example: pull from vimwiki (2nd configured wiki)
			save_dir = function()
				local list = vim.g.vimwiki_list
				if type(list) == "table" and list[2] and list[2].path then
					return vim.fn.expand(list[2].path)
				end
				return vim.fn.expand("~/tutorials")
			end,
			auto_open = true,
			markdown_preview = false,
		},

		explain = {
			output = "buffer",
			auto_save = true,
			save_dir = vim.fn.expand("~/explanations"),
			side_by_side = false,
			show_line_numbers = true,
			context_max_chars = 3000,
			context_surround_lines = 80,
		},

		gitmessage = {
			output = "float",
			auto_copy = true,
			float_opts = {
				relative = "editor",
				width = 0.6,
				height = 0.4,
				border = "rounded",
				title = " Git Commit Message ",
			},
			conventional_commits = true,
			show_diff = true,
			auto_insert = false,
		},

		ctx_edit = {
			show_spinner = true,
		},
	},

	-- Behavior
	verbose = false,
	timeout = 300000, -- 5 minutes in milliseconds

	-- History
	history = {
		enabled = true,
		max_entries = 100,
		save_to_file = true,
		file_path = vim.fn.stdpath("data") .. "/spicy_history.json",
	},

	-- Telescope integration
	telescope = {
		enabled = true,
		theme = "dropdown",
	},

	-- Statusline integration
	statusline = {
		enabled = false,
		show_running = true,
		show_last_result = false,
	},
})

-- ============================================================================
-- Keymaps - Set up your own keybindings here
-- ============================================================================

-- Normal mode keymaps
vim.keymap.set("n", "<leader>sa", "<cmd>SpicyAsk<CR>", {
	desc = "Spicy: Ask a question",
})

vim.keymap.set("n", "<leader>st", "<cmd>SpicyTutor<CR>", {
	desc = "Spicy: Generate tutorial",
})

vim.keymap.set("n", "<leader>se", "<cmd>SpicyExplain<CR>", {
	desc = "Spicy: Explain code",
})

vim.keymap.set("n", "<leader>sg", "<cmd>SpicyGitmessage<CR>", {
	desc = "Spicy: Generate git commit message",
})

vim.keymap.set("v", "<leader>sc", ":'<,'>SpicyCtxEdit<CR>", {
	desc = "Spicy: Edit selection",
})

-- Visual mode keymaps
vim.keymap.set("v", "<leader>sa", ":'<,'>SpicyAsk<CR>", {
	desc = "Spicy: Ask about selection",
})

vim.keymap.set("v", "<leader>se", ":'<,'>SpicyExplain<CR>", {
	desc = "Spicy: Explain selection",
})

-- Alternative keymap examples:
-- vim.keymap.set("n", "<leader>ai", "<cmd>SpicyAsk<CR>", { desc = "AI Ask" })
-- vim.keymap.set("n", "<leader>at", "<cmd>SpicyTutor<CR>", { desc = "AI Tutor" })
-- vim.keymap.set("n", "<leader>ae", "<cmd>SpicyExplain<CR>", { desc = "AI Explain" })
-- vim.keymap.set("n", "<leader>ac", "<cmd>SpicyGitmessage<CR>", { desc = "AI Commit" })
-- vim.keymap.set("v", "<leader>as", ":'<,'>SpicyCtxEdit<CR>", { desc = "AI Edit" })

-- Using different leader keys:
-- vim.keymap.set("n", "<C-a>", "<cmd>SpicyAsk<CR>", { desc = "Spicy Ask" })
-- vim.keymap.set("v", "<C-e>", ":'<,'>SpicyExplain<CR>", { desc = "Spicy Explain" })
