//go:build !windows

package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/config"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	RunE: func(_ *cobra.Command, _ []string) error {
		data, err := os.ReadFile(config.DaemonPIDPath())
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("daemon not running (no pid file)")
				return nil
			}
			return err
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			return err
		}
		_ = syscall.Kill(pid, syscall.SIGTERM)
		for i := 0; i < 25; i++ {
			if err := syscall.Kill(pid, 0); err != nil {
				_ = os.Remove(config.DaemonPIDPath())
				fmt.Println("stopped")
				return nil
			}
			time.Sleep(200 * time.Millisecond)
		}
		_ = syscall.Kill(pid, syscall.SIGKILL)
		_ = os.Remove(config.DaemonPIDPath())
		fmt.Println("stopped (sigkill)")
		return nil
	},
}
