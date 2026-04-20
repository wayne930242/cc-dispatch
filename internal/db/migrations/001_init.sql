CREATE TABLE IF NOT EXISTS sessions (
  id             TEXT PRIMARY KEY,
  workspace      TEXT NOT NULL,
  app            TEXT NOT NULL,
  task           TEXT NOT NULL,
  cwd            TEXT NOT NULL,
  pid            INTEGER,
  status         TEXT NOT NULL,
  created_at     INTEGER NOT NULL,
  started_at     INTEGER,
  ended_at       INTEGER,
  exit_code      INTEGER,
  jsonl_path     TEXT,
  stderr_path    TEXT,
  stdout_path    TEXT,
  error_message  TEXT,
  metadata_json  TEXT
);
CREATE INDEX IF NOT EXISTS idx_sessions_workspace ON sessions(workspace);
CREATE INDEX IF NOT EXISTS idx_sessions_status    ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_created   ON sessions(created_at DESC);

CREATE TABLE IF NOT EXISTS events (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id   TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  ts           INTEGER NOT NULL,
  kind         TEXT NOT NULL,
  payload_json TEXT
);
CREATE INDEX IF NOT EXISTS idx_events_session ON events(session_id, ts DESC);

CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY);
INSERT OR IGNORE INTO schema_version (version) VALUES (1);
