package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimeDirDefault(t *testing.T) {
	t.Setenv("CC_DISPATCH_HOME", "")
	home := HomeDir()
	require.Equal(t, filepath.Join(home, ".cc-dispatch"), RuntimeDir())
}

func TestRuntimeDirEnvOverride(t *testing.T) {
	t.Setenv("CC_DISPATCH_HOME", "/tmp/custom")
	require.Equal(t, "/tmp/custom", RuntimeDir())
}

func TestDerivedPaths(t *testing.T) {
	t.Setenv("CC_DISPATCH_HOME", "/tmp/ccd")
	require.Equal(t, "/tmp/ccd/config.json", ConfigPath())
	require.Equal(t, "/tmp/ccd/db.sqlite", DBPath())
	require.Equal(t, "/tmp/ccd/logs", LogsDir())
	require.Equal(t, "/tmp/ccd/daemon.pid", DaemonPIDPath())
	require.Equal(t, "/tmp/ccd/spawn.lock", SpawnLockPath())
}

func TestEncodeClaudeCwd(t *testing.T) {
	cases := []struct {
		in, out string
	}{
		{"/Users/me/proj", "-Users-me-proj"},
		{"/tmp/foo.bar_baz", "-tmp-foo-bar-baz"},
		{`C:\Users\me\proj`, "C--Users-me-proj"},
	}
	for _, c := range cases {
		require.Equal(t, c.out, EncodeClaudeCwd(c.in))
	}
}

func TestJsonlPathFor(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	got := JsonlPathFor("/Users/me/proj", "abc-123")
	require.Equal(t, "/home/u/.claude/projects/-Users-me-proj/abc-123.jsonl", got)
}
