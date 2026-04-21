//go:build !windows

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "..", "..")
}

func buildFake(t *testing.T, root string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "fake-claude")
	cmd := exec.Command("go", "build", "-o", out, "./test/fixtures/fake-claude")
	cmd.Dir = root
	require.NoError(t, cmd.Run())
	return out
}

func buildCcd(t *testing.T, root string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "ccd")
	cmd := exec.Command("go", "build", "-o", out, "./cmd/ccd")
	cmd.Dir = root
	require.NoError(t, cmd.Run())
	return out
}

func startDaemon(t *testing.T, ccd string) {
	t.Helper()
	c := exec.Command(ccd, "serve")
	c.Env = os.Environ()
	require.NoError(t, c.Start())
	time.Sleep(1500 * time.Millisecond)
	t.Cleanup(func() {
		_ = c.Process.Signal(os.Interrupt)
		_ = c.Wait()
	})
}

func readCfg(t *testing.T, tmp string) (int, string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(tmp, "config.json"))
	require.NoError(t, err)
	var cfg struct {
		Port  int    `json:"port"`
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(data, &cfg))
	return cfg.Port, cfg.Token
}

func httpURL(port int, path string) string {
	return "http://127.0.0.1:" + strconv.Itoa(port) + path
}

func TestEndToEndDispatch(t *testing.T) {
	root := repoRoot(t)
	fake := buildFake(t, root)
	ccd := buildCcd(t, root)

	tmp := t.TempDir()
	t.Setenv("CC_DISPATCH_HOME", tmp)
	t.Setenv("CC_DISPATCH_CLAUDE_BIN", fake)
	t.Setenv("FAKE_CLAUDE_SLEEP_MS", "100")
	t.Setenv("FAKE_CLAUDE_EXIT_CODE", "0")

	startDaemon(t, ccd)
	port, token := readCfg(t, tmp)

	startBody, _ := json.Marshal(map[string]string{
		"task": "hi", "app": "demo", "cwd": tmp,
	})
	req, _ := http.NewRequest("POST", httpURL(port, "/rpc/dispatch_start"), bytes.NewReader(startBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var startResp struct {
		SessionID string `json:"session_id"`
		ResumeCmd string `json:"resume_cmd"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&startResp))
	resp.Body.Close()
	require.Regexp(t, `^[0-9a-f-]{36}$`, startResp.SessionID)
	require.Contains(t, startResp.ResumeCmd, "claude --resume")

	time.Sleep(1500 * time.Millisecond)

	statusBody, _ := json.Marshal(map[string]string{"session_id": startResp.SessionID})
	req2, _ := http.NewRequest("POST", httpURL(port, "/rpc/dispatch_status"), bytes.NewReader(statusBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	require.Equal(t, 200, resp2.StatusCode)

	var status struct {
		Status   string `json:"status"`
		ExitCode *int   `json:"exit_code"`
	}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&status))
	resp2.Body.Close()
	require.Equal(t, "completed", status.Status)
	require.NotNil(t, status.ExitCode)
	require.Equal(t, 0, *status.ExitCode)
}

func TestMissingBearerRejected(t *testing.T) {
	root := repoRoot(t)
	ccd := buildCcd(t, root)

	tmp := t.TempDir()
	t.Setenv("CC_DISPATCH_HOME", tmp)

	startDaemon(t, ccd)
	port, _ := readCfg(t, tmp)

	req, _ := http.NewRequest("POST", httpURL(port, "/rpc/dispatch_list"), bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 401, resp.StatusCode)
}
