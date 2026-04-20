# cc-dispatch Go Migration & Standalone Repo — v0.1.0 Design

**Date**: 2026-04-20
**Goal**: Extract cc-dispatch from the mp-dev plugin, port TypeScript → Go, and release it as a standalone open-source project at `github.com/wayne930242/cc-dispatch` with feature parity to Phase 1 (v0.1.0).
**Status**: draft, awaiting review.

---

## 1. Background

Phase 1 of cc-dispatch shipped inside the WayDoSoft-internal `mp-dev` plugin.
During the Phase 2 planning discussion, two observations reframed the project:

1. **Generality.** The git-workflow-for-dispatch pattern is not moldplan-specific.
   Anyone who runs multi-project dev work with Claude Code can benefit.
2. **Ecosystem alignment.** The author's earlier project `agent-workspace-engine` (`awe`) is Go-native.
   Language consistency between tools enables future integration and reduces cognitive overhead.

The decision is to:
- Extract cc-dispatch into a public MIT-licensed repo `wayne930242/cc-dispatch`
- Port the implementation from TypeScript (Node) to Go
- Distribute via GitHub Releases with an install script (no npm)
- Refactor the `mp-dev` plugin to consume the `ccd` binary instead of bundling source

The TypeScript version is retired; no parallel maintenance.

---

## 2. Scope

### In scope (v0.1.0)

- Feature parity with the current TypeScript Phase 1:
  - Daemon: HTTP server on `127.0.0.1:47821` with bearer auth, SQLite-backed session store
  - 5 RPC endpoints: `dispatch_start`, `dispatch_list`, `dispatch_status`, `dispatch_tail`, `dispatch_cancel`
  - Subprocess orchestration: spawn `claude -p` detached, capture exit code via in-memory handler, tracker loop for daemon-restart recovery
  - MCP stdio server exposing the same 5 tools
  - CLI subcommands: `serve`, `mcp`, `list`, `show`, `tail`, `cancel`, `resume-cmd`, `stop`, `version`
  - Auto-spawn daemon on first MCP invocation (file-locked double-check)
  - Bearer token in `~/.cc-dispatch/config.json`
- Multi-platform builds via GoReleaser: linux/darwin (amd64 + arm64), windows (amd64)
- Install script (`curl | sh`)
- GitHub Actions CI (test + lint) + Release pipeline (tag → build → publish)
- Public README with install + usage
- Retirement of the TypeScript implementation inside `mp-dev`
- `mp-dev` plugin refactor to consume the `ccd` binary

### Out of scope (later phases)

- **Phase 2**: git workflow (pull base, worktree, create MR/PR) — separate spec on the new repo
- **Phase 3**: webview
- Homebrew tap (deferred; revisit after adoption signals)
- Active Windows interactive testing (Windows builds must compile and smoke-check, but integration tests stay POSIX-only as today)
- Package-manager distribution (npm, pip, deb, rpm): not for v0.1.0

---

## 3. Project Layout

```
cc-dispatch/
├── cmd/
│   └── ccd/
│       └── main.go             # entry, routes subcommands
├── internal/
│   ├── config/                 # paths (env-override aware), constants
│   │   └── config.go
│   ├── db/
│   │   ├── db.go               # open, migrate
│   │   ├── migrations/
│   │   │   └── 001_init.sql
│   │   └── queries.go
│   ├── auth/
│   │   └── token.go            # load/create config.json, verify bearer
│   ├── jobs/
│   │   ├── spawner.go
│   │   ├── tracker.go
│   │   └── manager.go
│   ├── daemon/
│   │   ├── server.go           # net/http + routes wiring
│   │   ├── routes_health.go
│   │   ├── routes_rpc.go       # 5 endpoints
│   │   └── serve.go            # top-level Serve() composing server+tracker+lifecycle
│   ├── mcp/
│   │   ├── server.go           # stdio MCP via mark3labs/mcp-go
│   │   └── autospawn.go        # health probe + flock + detached spawn
│   ├── client/
│   │   └── client.go           # HTTP client shared by MCP + CLI
│   └── cli/
│       ├── root.go             # cobra root
│       ├── cmd_list.go
│       ├── cmd_show.go
│       ├── cmd_tail.go
│       ├── cmd_cancel.go
│       ├── cmd_resume_cmd.go
│       ├── cmd_stop.go
│       └── cmd_version.go
├── test/
│   ├── fixtures/
│   │   └── fake-claude.go      # builds into small Go binary for tests
│   ├── integration/
│   │   └── dispatch_flow_test.go
│   └── testutil/
│       └── harness.go          # tmp runtime dir, shim setup
├── docs/
│   └── architecture.md
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── .goreleaser.yaml
├── install.sh
├── Makefile
├── go.mod
├── go.sum
├── LICENSE                     # MIT
└── README.md
```

### Layout rationale

- `cmd/ccd/main.go` kept minimal; all logic under `internal/` so it is non-importable by external packages (we are not publishing a Go library yet, only a binary).
- Subpackages mirror the TS module structure so migration is one-to-one and reviewers can trace parity.
- `test/` at the top level (not `_test.go` only) to host fixtures and integration tests that drive the binary end-to-end.

---

## 4. Tech Stack

| Concern | Choice | Notes |
|---|---|---|
| HTTP server | `net/http` stdlib | API surface is small; no framework needed. |
| Routing | stdlib `http.ServeMux` | Go 1.22+ supports method+path patterns. |
| SQLite | `modernc.org/sqlite` | Pure-Go driver; no CGO; cross-compile clean. Trade some perf for distribution simplicity. |
| CLI | `github.com/spf13/cobra` | De-facto Go CLI framework. |
| MCP SDK | `github.com/mark3labs/mcp-go` | Only mature Go MCP library today. API-stability risk noted. |
| Config persistence | JSON via stdlib `encoding/json` | `~/.cc-dispatch/config.json`. |
| File locking | `github.com/gofrs/flock` | Cross-platform (POSIX fcntl + Windows LockFileEx). |
| Logging | `log/slog` stdlib | Structured logs; default text handler. |
| Testing | stdlib `testing` + `github.com/stretchr/testify/require` | `require` for clean assertions without dependency chains. |
| Release pipeline | `github.com/goreleaser/goreleaser` | Cross-platform binary builds + archives + checksums + GitHub Releases. |
| Lint | `golangci-lint` | Default ruleset plus `gofumpt`, `revive`. |

Go toolchain minimum: **Go 1.22** (for the `net/http` routing patterns; widely available via install scripts by April 2026).

---

## 5. Data & Runtime Layout (unchanged from Phase 1)

```
~/.cc-dispatch/
├── config.json              # port, token, version, created_at; 0600
├── db.sqlite                # sessions + events + schema_version
├── logs/
│   ├── <sid>.stdout         # claude stream-json output
│   └── <sid>.stderr
├── daemon.pid               # current daemon's pid
├── daemon.log               # daemon's own log output
└── spawn.lock               # ensureDaemon mutex (gofrs/flock)
```

Environment override for tests: `CC_DISPATCH_HOME=<path>` (same as TS version).

Session transcripts continue to live at `~/.claude/projects/<encoded-cwd>/<session-id>.jsonl` (owned by Claude Code, we only read).

---

## 6. RPC + MCP Surface

Unchanged from Phase 1 — reuse the types and endpoint shapes.
The Go types in `internal/daemon/types.go` mirror the TS `shared/types.ts`:

```go
type SessionStatus string

const (
    StatusQueued    SessionStatus = "queued"
    StatusRunning   SessionStatus = "running"
    StatusCompleted SessionStatus = "completed"
    StatusFailed    SessionStatus = "failed"
    StatusCancelled SessionStatus = "cancelled"
)

type SessionRow struct {
    ID           string        `json:"id"`
    Workspace    string        `json:"workspace"`
    App          string        `json:"app"`
    Task         string        `json:"task"`
    Cwd          string        `json:"cwd"`
    PID          *int          `json:"pid"`
    Status       SessionStatus `json:"status"`
    CreatedAt    int64         `json:"created_at"`
    StartedAt    *int64        `json:"started_at"`
    EndedAt      *int64        `json:"ended_at"`
    ExitCode     *int          `json:"exit_code"`
    JsonlPath    *string       `json:"jsonl_path"`
    StderrPath   *string       `json:"stderr_path"`
    StdoutPath   *string       `json:"stdout_path"`
    ErrorMessage *string       `json:"error_message"`
    MetadataJSON *string       `json:"metadata_json"`
}
```

RPC paths, request shapes, and response shapes stay identical.
MCP tool names (`dispatch_start`, etc.) stay identical so `mp-dev`'s skill calls them unchanged after re-pointing the MCP server command at `ccd mcp`.

---

## 7. Subprocess Model

Port the TS pattern verbatim:

1. `manager.Start(sessionID)` calls `spawner.Spawn(...)`.
2. `spawner.Spawn` creates log files, calls `exec.Command("claude", args...)` with:
   - `Stdout = logFile`, `Stderr = logFile`
   - `SysProcAttr.Setpgid = true` on POSIX (own process group so parent dying doesn't cascade)
   - On Windows, `CREATE_NEW_PROCESS_GROUP`
3. Start the command, then call `cmd.Process.Release()` (Go equivalent of `unref()` — detach but keep observing exit via a goroutine).
4. Manager goroutine: `cmd.Wait()` → resolve status from exit code + signal, update DB.
5. Tracker loop (`ticker := time.NewTicker(5 * time.Second)`) scans `status='running'` rows; for each, `syscall.Kill(pid, 0)` (POSIX) or `FindProcess` (Windows) to check liveness; dead pid + no in-memory handler → fall back to `completed` with `exit_code = NULL` and explanatory `error_message`.

`cmd.Process.Release()` keeps the subprocess alive after daemon exits (matches TS `unref` semantics). On daemon restart, the tracker handles recovery.

---

## 8. Distribution

### GitHub Releases

GoReleaser produces these artifacts per tag push:

```
cc-dispatch_<version>_linux_amd64.tar.gz
cc-dispatch_<version>_linux_arm64.tar.gz
cc-dispatch_<version>_darwin_amd64.tar.gz
cc-dispatch_<version>_darwin_arm64.tar.gz
cc-dispatch_<version>_windows_amd64.zip
checksums.txt
```

Each archive contains the `ccd` binary, `LICENSE`, and `README.md`.

### `install.sh`

Single-file POSIX sh installer:

1. Detect OS + arch (`uname -s -m`)
2. Query GitHub API for latest release tag (or accept `CC_DISPATCH_VERSION` env)
3. Download matching tarball + checksum
4. Verify SHA256
5. Extract `ccd` into `$HOME/.cc-dispatch/bin/ccd`, `chmod +x`
6. Print PATH hint: `export PATH="$HOME/.cc-dispatch/bin:$PATH"`
7. Exit 0

The installer also handles upgrade (overwrites the binary).
No Go toolchain required for consumers.

For plugin-driven install, `mp-dev`'s SessionStart hook calls the installer when `ccd` is not on PATH.

### Optional `go install`

Users who prefer `go install github.com/wayne930242/cc-dispatch/cmd/ccd@latest` can still do so (requires Go toolchain).
Both paths land the same binary.

---

## 9. CI & Release Pipeline

### `.github/workflows/ci.yml`

Triggers: PR + push to `main`.
Jobs:
- `test`: `go test ./...` on ubuntu-latest (POSIX integration tests)
- `test-windows`: `go test ./... -skip 'Integration'` on windows-latest (compile + unit tests only)
- `lint`: `golangci-lint run`

### `.github/workflows/release.yml`

Triggers: tag push matching `v*`.
Jobs:
- `release`: `goreleaser release` with `GITHUB_TOKEN` secret.
  GoReleaser handles: cross-compile, archive, checksum, upload to GitHub Release, create release notes from commits.

Branch protection: require CI green before merge to `main`. Tagged releases only from `main`.

---

## 10. Migration of mp-dev Plugin

Once `wayne930242/cc-dispatch` is at v0.1.0:

1. `mp-dev` `plugin.json`:
   ```json
   "mcpServers": {
     "cc-dispatch": { "command": "ccd", "args": ["mcp"] }
   }
   ```
   (Relies on `ccd` being on PATH.)

2. `mp-dev` SessionStart hook:
   - Check `command -v ccd`
   - If absent, run `install.sh` (fetched via curl from the new repo's `main`) targeting `~/.cc-dispatch/bin/`
   - Add `$HOME/.cc-dispatch/bin` to `PATH` for the session (via hook-controlled shell RC injection, or just instruct user on first run)

3. Delete:
   - `plugins/mp-dev/cc-dispatch/` entire TypeScript subdirectory
   - `scripts/mp-dev/src/mp_dev/commands/plugin_dev.py` and its registration
   - The "迭代 mp-dev plugin" section in `scripts/mp-dev/README.md`

4. `mp-dev` bumped to **0.2.0** in plugin.json, marketplace.json, and the plugin README table.

The plugin's `dispatching-work` skill stays unchanged — it calls the same MCP tool names.

---

## 11. Testing Strategy

### Unit tests (Go)

Per module: `config`, `db`, `auth/token`, `jobs/spawner`, `jobs/tracker`, `jobs/manager`.
Use `testing` + `testify/require`.
Follow the TS coverage: minimum 2-3 happy + edge case tests per module.

### Integration test

`test/integration/dispatch_flow_test.go`:
- Build a small Go `fake-claude` binary in `test/fixtures/` (simulates `claude -p` with env-driven exit code and sleep)
- Start daemon in-process with `CC_DISPATCH_HOME` pointed at a tmp dir
- POST `dispatch_start` with `cwd=tmp` and `CC_DISPATCH_CLAUDE_BIN=<fake-claude path>`
- Assert end state `completed` + `exit_code=0`
- Assert missing-bearer returns 401

Mirrors the TS integration test's two cases.
POSIX only; gated with `//go:build !windows`.

### Smoke

Manual checklist in `docs/smoke.md`:
- `install.sh` run from clean shell → binary on PATH
- `ccd serve &` then `curl /health`
- `ccd mcp` auto-spawn smoke (bash harness like today's)

---

## 12. Open Risks

- **mark3labs/mcp-go API stability.** If breaking changes land upstream, we may need to pin or fork.
  Mitigation: pin exact version in `go.mod`; review releases before bumping.
- **modernc.org/sqlite performance.** Pure-Go is slower than CGO sqlite.
  Mitigation: for cc-dispatch's workload (hundreds of sessions max, simple indexed queries) this is a non-issue. If it ever becomes one, swap in `mattn/go-sqlite3` behind a build tag.
- **Install-script trust.** Users run `curl | sh`.
  Mitigation: publish SHA256 checksums; README documents both `go install` and manual-download paths.
- **Windows distribution gap.** No active interactive testing.
  Mitigation: ship builds, document the gap in README, accept community PRs to close it.

---

## 13. Deferred to Phase 2 (next spec on the new repo)

- `dispatch_start` gains `base_branch?`, `create_worktree`, `create_pr`, `branch_name?`
- GitLab auth fallback chain (ssh-agent → glab CLI → env `GITLAB_TOKEN` → `~/.gitlab-token`) — aligns with awe's pattern
- Worktree storage at `~/.cc-dispatch/worktrees/<sid-short>/`
- Branch naming `cc-dispatch/<sid-short>`
- Explicit cleanup RPC `dispatch_cleanup(session_id)` — no auto-delete

These design notes carry forward into the Phase 2 spec; they are not implemented in v0.1.0.

---

## 14. Success Criteria

v0.1.0 is done when:

- [ ] `github.com/wayne930242/cc-dispatch` exists, public, MIT-licensed, has README
- [ ] `go test ./...` passes on POSIX; Windows CI passes unit tests
- [ ] Tagged `v0.1.0` release has 5 archives + checksums on GitHub Releases
- [ ] `curl -fsSL https://raw.githubusercontent.com/wayne930242/cc-dispatch/main/install.sh | sh` from a clean machine installs `ccd` and `ccd --version` works
- [ ] `ccd serve` + `ccd list` + `ccd stop` round-trip works
- [ ] `ccd mcp` auto-spawns daemon, responds to MCP `list_tools` with the 5 tool definitions
- [ ] `mp-dev` plugin bumped to 0.2.0; old cc-dispatch/ dir and plugin-dev Python command deleted; `/plugin install mp-dev@waydosoft-plugins` + reload works end-to-end with the binary on PATH
- [ ] Phase 1 smoke test from `docs/superpowers/plans/2026-04-20-mp-dev-cc-dispatch-phase1-smoke-results.md` re-runs green against the Go binary

---

## 15. References

- Phase 1 spec: `docs/superpowers/specs/2026-04-20-mp-dev-cc-dispatch-phase1-design.md`
- Phase 1 plan: `docs/superpowers/plans/2026-04-20-mp-dev-cc-dispatch-phase1.md`
- Related project (Go-native stack precedent): `github.com/wayne930242/agent-workspace-engine`
- MCP Go SDK: `github.com/mark3labs/mcp-go`
- GoReleaser: `github.com/goreleaser/goreleaser`
- Pure-Go SQLite: `modernc.org/sqlite`

---

**End of spec.** After approval, the plan (MG-01 through MG-18 from the chat preview) gets formalized via `superpowers:writing-plans`.
