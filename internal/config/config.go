package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultPort      = 47821
	DaemonVersion    = "0.1.1"
	TrackerInterval  = 5
	HealthTimeout    = 800
	SpawnPollTimeout = 10
	MaxTaskBytes     = 50 * 1024
)

func HomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return h
}

func RuntimeDir() string {
	if v := os.Getenv("CC_DISPATCH_HOME"); v != "" {
		return v
	}
	return filepath.Join(HomeDir(), ".cc-dispatch")
}

func ConfigPath() string    { return filepath.Join(RuntimeDir(), "config.json") }
func DBPath() string        { return filepath.Join(RuntimeDir(), "db.sqlite") }
func LogsDir() string       { return filepath.Join(RuntimeDir(), "logs") }
func DaemonLogPath() string { return filepath.Join(RuntimeDir(), "daemon.log") }
func DaemonPIDPath() string { return filepath.Join(RuntimeDir(), "daemon.pid") }
func SpawnLockPath() string { return filepath.Join(RuntimeDir(), "spawn.lock") }
func ClaudeHome() string    { return filepath.Join(HomeDir(), ".claude") }

// EncodeClaudeCwd converts an absolute cwd to the path segment Claude Code uses
// under ~/.claude/projects/. Every non-alphanumeric character becomes '-'.
func EncodeClaudeCwd(cwd string) string {
	out := make([]byte, 0, len(cwd))
	for i := 0; i < len(cwd); i++ {
		c := cwd[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			out = append(out, c)
		} else {
			out = append(out, '-')
		}
	}
	return string(out)
}

func JsonlPathFor(cwd, sessionID string) string {
	return filepath.Join(ClaudeHome(), "projects", EncodeClaudeCwd(cwd), sessionID+".jsonl")
}

func StdoutLogPath(sessionID string) string {
	return filepath.Join(LogsDir(), sessionID+".stdout")
}

func StderrLogPath(sessionID string) string {
	return filepath.Join(LogsDir(), sessionID+".stderr")
}
