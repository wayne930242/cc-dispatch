---
paths:
  - "internal/daemon/**"
  - "internal/mcp/**"
  - "internal/db/**"
  - "internal/client/**"
---

# Daemon & Protocol Stability

These packages form the public contract with installed clients (MCP consumers, `~/.cc-dispatch/` state on user machines). Breaking changes silently orphan existing installations.

## HTTP RPC surface (`internal/daemon/`, `internal/client/`)

- NEVER rename, remove, or change argument shape of an existing daemon HTTP endpoint. Add a new endpoint instead.
- Bearer-token auth header format is part of the contract: `Authorization: Bearer <token>` reading `~/.cc-dispatch/config.json`. NEVER alter header name or token location.
- Response JSON fields may be added. Existing fields MUST NOT be renamed, removed, or have their type changed.

## MCP tool surface (`internal/mcp/`)

- The five public MCP tools (`dispatch_start`, `dispatch_list`, `dispatch_status`, `dispatch_tail`, `dispatch_cancel`) are a stable API. NEVER rename them.
- Tool input schema fields MUST be backward compatible: new fields are optional; existing fields MUST NOT change type or become required.
- Auto-spawn behavior on first MCP call is part of the contract — MUST NOT be removed without a new flag to opt out.

## SQLite schema (`internal/db/`)

- Schema changes MUST add a new numbered migration file under `internal/db/migrations/NNN_*.sql`. NEVER edit an existing migration after it has been released.
- Migrations MUST bump `schema_version` and be idempotent (`IF NOT EXISTS`, `INSERT OR IGNORE`).
- NEVER drop a column or table that exists in a released version without a multi-step migration plan (new column → dual write → backfill → deprecate).

## When a break is unavoidable

Document the break in `CHANGELOG.md` under a `### Breaking` heading, bump the major version, and surface it in `README.md` install instructions. Coordinate with the user before proceeding.
