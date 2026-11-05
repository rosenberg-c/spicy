# v-edit `$HOME` Hang — Findings & Fix Report

**Date:** 2026-08-25
**Component:** `spicy v-edit` (Go CLI `cli/cmd/ctx-edit` + nvim plugin `nvim/lua/spicy/`)
**Status:** Root cause identified; fixes implemented, tested, and verified; environment cleaned up.

## 1. Executive Summary

`spicy v-edit` works reliably inside a project directory but appears to "hang" (spinner never exits) when editing a file in the home directory (e.g. `~/.zsh_session`), even though the local LLM (lm-studio / `lms/qwen/qwen3.8-27b`) is visibly active in `nvtop`.

It is **not** a hang. The model call is being made, but the whole operation is **silently killed after a hard 5-minute budget** before the result can be returned and applied to nvim. The budget is consumed by opencode's per-"project" bootstrap, which in `$HOME` is pathologically expensive (a `git add --all` over the entire home tree + a file-picker that refuses to run in `$HOME`). Compounding this, the timeout kill does not kill opencode's **grandchild** processes, leaving orphaned `git add` processes that burn CPU for many hours and hold a snapshot `index.lock`, making each subsequent `$HOME` retry slower.

Two fixes were implemented: (1) kill the entire process group on timeout so no orphaned grandchildren remain, and (2) pin the `opencode run` working directory to a small scratch directory so bootstrap cost is near-zero regardless of where the target file lives.

## 2. Symptom

- `spicy v-edit ask and explain` on a file in a project dir (e.g. `~/personal/dev/spicy/`) → works.
- Same command in `$HOME` (e.g. `~/.zsh_session`) → nvim spinner keeps spinning; the model is active in `nvtop`; after a while it times out/fails with no edit applied.

## 3. Reproduction

1. Open nvim in `$HOME`.
2. Run `spicy v-edit ask and explain` against `~/.zsh_session`.
3. Observe: GPU/model activity in `nvtop` for minutes; nvim never applies the edit; eventually a timeout/failure.

## 4. Root-Cause Analysis

### 4.1 A hard 5-minute budget on both sides
- Go CLI: `cli/cmd/ctx-edit/main.go:82` — `context.WithTimeout(ctx, 5*time.Minute)`; `cli/internal/agent/agent.go` spawns `exec.CommandContext(ctx, "opencode", "run", ...)`. At t=5:00 the child is SIGKILLed.
- nvim: `nvim/lua/spicy/config.lua:89` — `timeout = 300000` (ms); `ctx_edit.lua:99` passes it into `job.lua` (plenary.job, `cwd = vim.fn.getcwd()`, `vim.defer_fn` timeout → `job:shutdown()`).

So the entire pipeline — bootstrap + LLM call + result — must complete within 5 minutes.

### 4.2 Bootstrap is pathologically expensive in `$HOME`
opencode treats the cwd as the "project" and, on start, (a) snapshots it and (b) initializes its FFF file-picker. In `$HOME`:
- Snapshot = `git add --all --sparse` over the **entire home directory** — large and slow.
- FFF **refuses** to run in `$HOME`: log shows `failed to initialize fff — "Can not run certain FFF features in a file system root or home directories"` (13× in the failing run).
- Concurrent runs collide on the snapshot git index: `failed to add snapshot files exitCode=128 — fatal: Unable to create '.../snapshot/b5f0b832.../index.lock': File exists` (6×).

Log evidence (run `63bbfab6`, timestamps UTC; local = UTC+2): instance created **08:54:58**, but the main edit LLM call did not start until **08:58:04** — **3m06s of bootstrap** — leaving only ~114s before the 5-minute kill. The edit call is thus SIGKILLed mid-generation. The model is visibly running (that's the `nvtop` GPU activity) but the result never returns → the spinner keeps going until the kill.

### 4.3 The timeout kill orphans grandchildren
Killing `v-edit` (SIGTERM from nvim timeout) or `opencode` (SIGKILL from the Go context) does **not** kill the grandchild tree (`opencode` → shell → `git add`). There is no process-group kill anywhere. Result: orphaned `git add` processes that keep running and hold the snapshot `index.lock`.

Live evidence captured on the machine (before the user killed it):
```
PID 66661  PPID 1  STARTED Mon Aug 24 16:54:27  ELAPSED 18:29:19  %CPU 99.3
git ... --git-dir ~/.local/share/opencode/snapshot/b5f0b832.../75e843c... --work-tree /Users/christian add --all --sparse ...
```
Parent long dead (PPID=1), ~18.5 hours at ~99% CPU (~1090 CPU-minutes), still adding the home tree to the snapshot, holding the lock that made every subsequent `$HOME` run fail/slow.

### 4.4 Retry-storm amplification
Aug 24: **27 `$HOME` runs in 70 minutes** (14:52–16:03 UTC) + 3 this morning. All share one snapshot repo (projectID `b5f0b832a3c78d123d8e710f6a126ce784d30481`), so concurrent `git add` + a stale `index.lock` → each retry slower → more likely to time out → more orphans. A vicious cycle.

### 4.5 Why a project dir works
Small file tree → snapshot `git add` is fast and succeeds; FFF initializes fine (not `$HOME`/root); no shared-lock contention; bootstrap ≈2s (observed `spicy-ctx-edit-cwd` runs: 16:06:04→16:06:06) → the full ~5 minutes is available for the LLM call → JSON out → applied.

### 4.6 Prior (abandoned) attempt
A previous commit pinned the job cwd to `~/.cache/nvim/spicy-ctx-edit-cwd` (8 runs used it, Aug 24 16:06–16:08). It was removed from the tree/config, and those runs still stalled after the 2nd LLM call — the orphan/lock/budget problems remained, so cwd-pinning alone was insufficient without also fixing the orphan kill.

## 5. Configuration Context (not the differentiator)
- `~/.config/spicy/config.json` → `default_model: lms/qwen/qwen3.8-27b`.
- `~/.config/opencode/opencode.json` → LM Studio providers. No `$HOME`-specific config; config is not the cause.

## 6. Fixes Implemented

### Fix 1 — Kill the whole `opencode run` process tree on timeout/failure
File: `cli/internal/agent/agent.go`
- `agent.go:62` — child runs in its own process group (`SysProcAttr{Setpgid: true}`).
- `agent.go:92-111` — `cmd.Run()` replaced with `Start()` + a watcher goroutine that `select`s on `ctx.Done()` → `killProcessGroup(cmd)`, then `Wait()` + `close(done)`. The watcher exits on either path (no goroutine leak).
- `agent.go:49` — `killProcessGroup` nil-checks `cmd.Process` and issues `syscall.Kill(-pid, SIGKILL)` (negative PID = entire group); errors ignored (best-effort).
- `agent.go:111` — best-effort group kill also on non-zero `Wait()` errors (catches grandchildren left behind when the direct child exits non-zero).
- Error reporting unchanged: still `agent command failed: %w`; the CLI still surfaces a clear timeout/kill error and non-zero exit.

### Fix 2 — Pin subprocess cwd to a scratch dir (HOME-aware bootstrap)
File: `cli/internal/agent/agent.go`
- `agent.go:39` — `scratchCwdDir(base)` → `<base>/spicy/v-edit-cwd`, created with `os.MkdirAll(..., 0o755)`.
- `agent.go:65-74` — `Run` resolves `os.UserCacheDir()` and sets `cmd.Dir` to the scratch dir (on macOS: `~/Library/Caches/spicy/v-edit-cwd`). opencode's snapshot/FFF now operate on a tiny tree instead of `$HOME` (or any large dir).

### Fix 3 — Absolute target-file path (required by Fix 2)
File: `cli/cmd/ctx-edit/main.go`
- `main.go:214-219` — `getContextInput` resolves `filepath.Abs(fileFlag)` before `os.ReadFile`.
- `main.go:247` — the returned `selection.filePath` is absolute, so the write path (`applyUpdate`) and JSON `file_path` are absolute regardless of the caller's cwd. nvim only consumes `updated_text`, so the contract is safe.

### Requirements & test coverage
- `docs/requirements/shared-cli.md` — `CORE-CLI-007` (process-group kill), `CORE-CLI-008` (scratch cwd), each with `Owner:`.
- `docs/requirements/ctx-edit.md` — `CLI-CTX-003` (absolute file path).
- `cli/internal/agent/agent_test.go` — `TestKillProcessGroup_KillsGrandchildren` (real `sh` grandchild that would write a marker if it survived the group kill), `TestKillProcessGroup_NilProcess`, `TestScratchCwdDir_CreatesPinnedDir`.
- `cli/cmd/ctx-edit/main_test.go` — `TestGetContextInput_ResolvesAbsolutePath` (note: on macOS `filepath.Abs` resolves the `/var → /private/var` symlink, so the assertion compares against `EvalSymlinks`).
- `docs/TEST_MATRIX.md` — regenerated via `make sync-test-matrix`; all three new IDs **Covered**.

## 7. Verification
- `go build ./...` ✅
- `go vet ./internal/agent/... ./cmd/ctx-edit/...` ✅
- `go test ./internal/agent/... ./cmd/ctx-edit/...` ✅ (agent suite ~1.3s, includes the live process-group kill test)
- Full `go test ./...` — green except one **pre-existing** failure: `internal/constants/constants_test.go` references `defaultModelEnvKey`/`originalWD`, removed when model resolution moved to `config.json` (commit `e47492c`). Unrelated to this change; that package's tests don't compile as of that commit.

## 8. Environment Cleanup
- The user killed the orphaned `git add` (PID 66661).
- Verified post-kill: no stray snapshot `git add`/`opencode` processes (only this session's own opencode + one unrelated ssh/tmux line).
- `find ~/.local/share/opencode/snapshot -name index.lock` → **0 lock files** (14 snapshot dirs). Store is clean; nothing to delete.

## 9. Remaining / Follow-ups
1. **Pre-existing test failure** — `internal/constants/constants_test.go` is stale (references removed env-key/WD symbols). Decide: restore env-var model resolution, or update the stale tests. (Not caused by this change.)
2. **Surface progress in nvim** — `cmd.Stderr` is currently discarded unless `--verbose`, so long waits show a bare spinner. Consider streaming stderr/progress into the nvim UI.
3. **Budget** — the scratch-cwd fix makes bootstrap near-zero, so the 5-minute budget is now effectively all model time. Consider making it configurable if very large edits are needed.
4. **Optional hardening** — surface/act on the FFF and snapshot warnings in `$HOME` contexts, and add a log line when the process-group kill fires, for easier future diagnosis.

## 10. Appendix — Key File/Line References
| Area | File:Line | What |
|---|---|---|
| Go 5-min timeout | `cli/cmd/ctx-edit/main.go:82` | `context.WithTimeout(ctx, 5*time.Minute)` |
| `opencode run` spawn | `cli/internal/agent/agent.go:37` (pre-fix) | `exec.CommandContext(ctx, "opencode", "run", ...)` |
| Process-group kill | `cli/internal/agent/agent.go:49,62,92-111` | Setpgid + group kill on timeout/failure |
| Scratch cwd | `cli/internal/agent/agent.go:39,65-74` | `scratchCwdDir`, `cmd.Dir` |
| Absolute file path | `cli/cmd/ctx-edit/main.go:214-219,247` | `filepath.Abs` |
| nvim timeout | `nvim/lua/spicy/config.lua:89` | `timeout = 300000` |
| nvim job | `nvim/lua/spicy/ctx_edit.lua:99`, `nvim/lua/spicy/job.lua` | plenary.job, `cwd=vim.fn.getcwd()`, `vim.defer_fn` → `job:shutdown()` |
