//go:build windows

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
		if out, err := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F").CombinedOutput(); err != nil {
			fmt.Printf("taskkill failed: %v\n%s\n", err, out)
		}
		_ = os.Remove(config.DaemonPIDPath())
		fmt.Println("stopped")
		return nil
	},
}
