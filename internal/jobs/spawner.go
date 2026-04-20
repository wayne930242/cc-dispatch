package jobs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wayne930242/cc-dispatch/internal/config"
)

type SpawnInput struct {
	SessionID      string
	Task           string
	Cwd            string
	ClaudeBin      string   // defaults to ResolveClaudeBin()
	ExecArgsPrefix []string // test-only: prepend args before the claude args
}

type SpawnResult struct {
	Cmd *exec.Cmd
	PID int
}

func ResolveClaudeBin() string {
	if v := os.Getenv("CC_DISPATCH_CLAUDE_BIN"); v != "" {
		return v
	}
	return "claude"
}

// Spawn launches the claude subprocess with stdout/stderr captured to
// per-session log files. The caller must call result.Cmd.Wait() to collect
// the exit code (typically in a goroutine).
func Spawn(in SpawnInput) (*SpawnResult, error) {
	bin := in.ClaudeBin
	if bin == "" {
		bin = ResolveClaudeBin()
	}
	args := append([]string{}, in.ExecArgsPrefix...)
	args = append(args,
		"-p", in.Task,
		"--session-id", in.SessionID,
		"--output-format", "stream-json",
		"--verbose",
		"--include-partial-messages",
	)

	if err := os.MkdirAll(filepath.Dir(config.StdoutLogPath(in.SessionID)), 0o700); err != nil {
		return nil, fmt.Errorf("mkdir logs: %w", err)
	}

	stdoutF, err := os.OpenFile(config.StdoutLogPath(in.SessionID),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	stderrF, err := os.OpenFile(config.StderrLogPath(in.SessionID),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		_ = stdoutF.Close()
		return nil, err
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = in.Cwd
	cmd.Stdout = stdoutF
	cmd.Stderr = stderrF
	cmd.Env = os.Environ()
	configureSysProcAttr(cmd) // platform-specific

	if err := cmd.Start(); err != nil {
		_ = stdoutF.Close()
		_ = stderrF.Close()
		return nil, fmt.Errorf("start %s: %w", bin, err)
	}

	// Child owns its own open fds now.
	_ = stdoutF.Close()
	_ = stderrF.Close()

	return &SpawnResult{Cmd: cmd, PID: cmd.Process.Pid}, nil
}
