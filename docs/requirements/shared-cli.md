# Shared CLI Foundations Requirements

### `CORE-CLI-001`

Every CLI command must accept direct arguments for primary input without
forcing an interactive prompt.

### `CORE-CLI-002`

Commands that persist outputs must write files atomically.

### `CORE-CLI-003`

History persistence must never fail silently.

### `CORE-CLI-004`

User-facing validation errors must be actionable.

### `CORE-CLI-005`

Owner: `cli/internal/constants/constants.go`, `cli/internal/constants/constants_test.go`,
`cli/cmd/ask/main.go`, `cli/cmd/explain/main.go`, `cli/cmd/tutor/main.go`,
`cli/cmd/gitmessage/main.go`, `cli/cmd/ctx-edit/main.go`

CLI commands must load persisted model configuration from the global config
file `~/.config/spicy/config.json` (layered under any per-repo local config,
see `CORE-CLI-009`), using `models.<command>` before `default_model`, while
allowing `-m`/`--model` to override the configured value for the current
invocation.

### `CORE-CLI-006`

Owner: `cli/internal/agent/auth.go`, `cli/internal/agent/auth_test.go`

CLI auth preflight must keep friendly validation for known providers such as
OpenAI and Anthropic, but must not reject unknown or custom model provider
identifiers before `opencode run` executes. Expired OpenAI or Anthropic OAuth
access tokens with a refresh token must not block `opencode run`, so OpenCode
can refresh them; missing credentials and expired tokens without a refresh
token must still produce actionable errors.

### `CORE-CLI-007`

Owner: `cli/internal/agent/agent.go`, `cli/internal/agent/agent_test.go`

The `opencode run` subprocess must run in its own process group, and the whole
group must be killed when the invocation times out or fails so no orphaned
grandchildren are left running.

### `CORE-CLI-008`

Owner: `cli/internal/agent/agent.go`, `cli/internal/agent/agent_test.go`

When the CLI's own working directory is the user's home directory, the
`opencode run` subprocess must run with its working directory pinned to a
small scratch directory under the user cache dir
(`<os.UserCacheDir()>/spicy/v-edit-cwd`), so model-generated files land in the
scratch directory instead of the home directory. When the CLI is run from
anywhere else, the subprocess must keep the CLI's actual working directory so
model-generated files land next to the caller's project.

### `CORE-CLI-009`

Owner: `cli/internal/constants/constants.go`, `cli/internal/constants/constants_test.go`

CLI commands must support a per-repo local config file named
`.spicy.config.json` with the same JSON shape as the global config
(`default_model` string, `models` map of command to model). Discovery starts
at the current working directory and checks that directory and each parent
directory up to the filesystem root; the first existing regular file named
`.spicy.config.json` wins, and a missing file means no local config.
Precedence is local `.spicy.config.json` over global
`~/.config/spicy/config.json` over the built-in fallback
`openai/gpt-5.3-codex`, merged per key: a non-empty local `default_model`
overrides the global one, each non-empty local `models` entry overrides the
matching global entry, and global `models` keys absent from the local file
are preserved. A missing, unreadable, or invalid local config file must be
ignored silently (no crash, no stderr noise) so the global config and
fallback still apply.
