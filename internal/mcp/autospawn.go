package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gofrs/flock"
	"github.com/wayne930242/cc-dispatch/internal/client"
	"github.com/wayne930242/cc-dispatch/internal/config"
)

func probeHealthy() bool {
	if _, err := os.Stat(config.ConfigPath()); err != nil {
		return false
	}
	c, err := client.FromConfigFile()
	if err != nil {
		return false
	}
	h, err := c.Health()
	return err == nil && h != nil && h.OK
}

// EnsureDaemon starts the daemon if it is not already healthy. It uses a
// filesystem lock to serialize concurrent callers.
func EnsureDaemon(exePath string) error {
	if probeHealthy() {
		return nil
	}
	if err := os.MkdirAll(config.RuntimeDir(), 0o700); err != nil {
		return err
	}
	lockPath := config.SpawnLockPath()
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		if err := os.WriteFile(lockPath, nil, 0o600); err != nil {
			return err
		}
	}
	l := flock.New(lockPath)
	if err := l.Lock(); err != nil {
		return err
	}
	defer func() { _ = l.Unlock() }()

	if probeHealthy() {
		return nil
	}

	cmd := exec.Command(exePath, "serve")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	configureDetach(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("spawn daemon: %w", err)
	}
	_ = cmd.Process.Release()

	deadline := time.Now().Add(time.Duration(config.SpawnPollTimeout) * time.Second)
	for time.Now().Before(deadline) {
		if probeHealthy() {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("daemon failed to become healthy within %ds", config.SpawnPollTimeout)
}
