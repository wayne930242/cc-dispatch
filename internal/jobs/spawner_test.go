package jobs

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func buildFakeClaude(t *testing.T) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Join(filepath.Dir(testFile), "..", "..")
	out := filepath.Join(t.TempDir(), "fake-claude")
	cmd := exec.Command("go", "build", "-o", out, "./test/fixtures/fake-claude")
	cmd.Dir = repoRoot
	require.NoError(t, cmd.Run())
	return out
}

func TestSpawnSuccess(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CC_DISPATCH_HOME", tmp)
	t.Setenv("FAKE_CLAUDE_SLEEP_MS", "50")
	fake := buildFakeClaude(t)

	res, err := Spawn(SpawnInput{
		SessionID: "sess-ok", Task: "hi", Cwd: tmp,
		ClaudeBin: fake,
	})
	require.NoError(t, err)
	require.Greater(t, res.PID, 0)
	require.NoError(t, res.Cmd.Wait())
}

func TestSpawnNonZeroExit(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CC_DISPATCH_HOME", tmp)
	t.Setenv("FAKE_CLAUDE_SLEEP_MS", "20")
	t.Setenv("FAKE_CLAUDE_EXIT_CODE", "3")
	fake := buildFakeClaude(t)

	res, err := Spawn(SpawnInput{
		SessionID: "sess-fail", Task: "x", Cwd: tmp,
		ClaudeBin: fake,
	})
	require.NoError(t, err)
	err = res.Cmd.Wait()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 3, exitErr.ExitCode())
}
