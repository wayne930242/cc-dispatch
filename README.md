# cc-dispatch

Dispatch headless Claude Code sessions, track their state, resume them interactively later.

The idea: running `claude -p` inside a project directory picks up that project's agent system (`CLAUDE.md`, local skills, hooks). But you may want to fire off that work from somewhere else — a different session, a chat on your phone, a scheduled job. `cc-dispatch` is a small daemon + MCP server + CLI that does the dispatch for you, keeps the `session_id` around, and lets you later `claude --resume` into it from any terminal.

- **MCP server** exposing `dispatch_start`, `dispatch_list`, `dispatch_status`, `dispatch_tail`, `dispatch_cancel`. Drop it into any MCP-speaking client.
- **Daemon** keeps a SQLite log of sessions and tracks subprocess lifecycle.
- **CLI** (`ccd list`, `ccd show <id>`, `ccd tail <id>`, …) for quick shell queries.
- **Zero runtime dependencies** — single static binary. No Node, no Python.

## Install

Once v0.1.0 ships:

```sh
curl -fsSL https://raw.githubusercontent.com/wayne930242/cc-dispatch/main/install.sh | sh
```

Installs to `~/.cc-dispatch/bin/ccd`. Add that to your `PATH`.

Alternative (requires Go toolchain, 1.24+):

```sh
go install github.com/wayne930242/cc-dispatch/cmd/ccd@latest
```

Windows: download the `.zip` from [Releases](https://github.com/wayne930242/cc-dispatch/releases) and put `ccd.exe` on your `PATH`.

## Usage

With `claude` on your `PATH`:

```sh
ccd serve &          # start daemon (usually auto-spawned by `ccd mcp`)
ccd list             # list sessions
ccd show <id>        # detail + resume command
```

To use as an MCP server (Claude Code plugin, other MCP client):

```json
{
  "mcpServers": {
    "cc-dispatch": { "command": "ccd", "args": ["mcp"] }
  }
}
```

The first MCP call auto-spawns the daemon.

## Data paths

- Runtime state: `~/.cc-dispatch/` (config, sqlite, logs, pid)
- Claude session transcripts: `~/.claude/projects/<encoded-cwd>/<session-id>.jsonl` (owned by Claude Code; cc-dispatch only reads)

`~/.cc-dispatch/config.json` contains your local auth token. Don't commit it.

## Subsystem map

```
┌─────────────┐   bearer   ┌────────────────────────┐
│ MCP client  │──────────▶│  ccd daemon (HTTP)     │
│ (Claude,    │            │  - 5 RPC endpoints     │
│  Cursor,    │            │  - sqlite sessions     │
│  your bot)  │            │  - subprocess tracker  │
└─────────────┘            └────────┬───────────────┘
                                    │ spawn
                                    ▼
                           ┌─────────────────────┐
                           │ claude -p           │
                           │ (headless session)  │
                           └─────────────────────┘
```

## Development

```sh
go test ./...          # unit + integration (POSIX only for integration)
go vet ./...
go build ./cmd/ccd
```

## License

MIT
