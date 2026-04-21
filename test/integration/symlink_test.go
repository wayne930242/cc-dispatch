//go:build !windows

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestDispatchResolvesSymlinks verifies that when a symlinked cwd is passed to
// dispatch_start, the stored cwd and jsonl_path reflect the resolved real path.
func TestDispatchResolvesSymlinks(t *testing.T) {
	root := repoRoot(t)
	fake := buildFake(t, root)
	ccd := buildCcd(t, root)

	tmp := t.TempDir()
	// Create a real directory + a symlink pointing to it
	realDir := filepath.Join(tmp, "real")
	require.NoError(t, os.Mkdir(realDir, 0o755))
	linkDir := filepath.Join(tmp, "link")
	require.NoError(t, os.Symlink(realDir, linkDir))

	t.Setenv("CC_DISPATCH_HOME", tmp)
	t.Setenv("CC_DISPATCH_CLAUDE_BIN", fake)
	t.Setenv("FAKE_CLAUDE_SLEEP_MS", "50")
	t.Setenv("FAKE_CLAUDE_EXIT_CODE", "0")

	startDaemon(t, ccd)
	port, token := readCfg(t, tmp)

	body, _ := json.Marshal(map[string]string{
		"task": "ping", "app": "demo", "cwd": linkDir,
	})
	req, _ := http.NewRequest("POST", httpURL(port, "/rpc/dispatch_start"), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var start struct {
		SessionID string `json:"session_id"`
		Cwd       string `json:"cwd"`
		JsonlPath string `json:"jsonl_path"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&start))
	resp.Body.Close()

	// The response cwd should be the real path, not the symlink
	realResolved, err := filepath.EvalSymlinks(realDir)
	require.NoError(t, err)
	require.Equal(t, realResolved, start.Cwd,
		"dispatch_start should store the symlink-resolved cwd")
	// jsonl_path should encode the resolved cwd, not the symlink
	require.Contains(t, start.JsonlPath, filepath.Base(realResolved),
		"jsonl_path should be derived from the resolved cwd")

	// And status should echo the same
	time.Sleep(800 * time.Millisecond)
	statusBody, _ := json.Marshal(map[string]string{"session_id": start.SessionID})
	req2, _ := http.NewRequest("POST", httpURL(port, "/rpc/dispatch_status"), bytes.NewReader(statusBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, 200, resp2.StatusCode)

	var status struct {
		Cwd       string `json:"cwd"`
		ResumeCmd string `json:"resume_cmd"`
	}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&status))
	require.Equal(t, realResolved, status.Cwd)
	require.Contains(t, status.ResumeCmd, realResolved)

	_ = strconv.Itoa // keep import used across tests
}
