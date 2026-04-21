# cc-dispatch

Go CLI + daemon + MCP server for dispatching headless Claude Code sessions, tracking their state, and resuming them interactively later.

## Project layout

- `cmd/ccd/` — CLI entry point
- `internal/daemon/` — HTTP RPC server, subprocess tracker
- `internal/mcp/` — MCP server (`dispatch_start`, `dispatch_list`, `dispatch_status`, `dispatch_tail`, `dispatch_cancel`)
- `internal/db/` — sqlite session store
- `internal/client/` — daemon client
- `internal/cli/`, `internal/auth/`, `internal/config/`, `internal/jobs/` — supporting packages
- `test/` — integration tests (POSIX only)
- `docs/spec.md` — architecture spec
- `.goreleaser.yaml` + `.github/workflows/release.yml` — release pipeline

## Verification commands

Run before claiming a task done:

- `go test ./...` — unit + integration (integration tests are POSIX only; they skip on Windows)
- `go build ./cmd/ccd` — ensure the CLI still compiles

`go vet` runs automatically via PostToolUse hook on every `.go` edit. `golangci-lint run` is CI-only — do not invoke locally unless diagnosing a CI failure.

## Constraints

- **Zero runtime dependencies.** The binary must stay a single static Go build — no Node, no Python, no shared libraries. Do not add dependencies that require cgo or external binaries without explicit user approval.

## Gotchas

- `~/.cc-dispatch/config.json` contains a local auth bearer token. Never echo its contents in logs, tests, diagnostics, or commit examples.
- Claude session transcripts live under `~/.claude/projects/<encoded-cwd>/<session-id>.jsonl` and are owned by Claude Code. cc-dispatch reads them; do not write.
- Integration tests in `test/` assume a working `claude` binary on `PATH`. If absent, they skip rather than fail.
- `install.sh` symlinks `ccd` into an existing `PATH` dir. Never rename or remove the target symlink in `install.sh` without bumping `CHANGELOG.md` under a Breaking section.
