package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/client"
	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

var resumeCmd = &cobra.Command{
	Use:   "resume-cmd <id>",
	Short: "Print the shell command to resume a session interactively",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		c, err := client.FromConfigFile()
		if err != nil {
			return err
		}
		var out daemon.DispatchStatusResponse
		if err := c.RPC("dispatch_status",
			daemon.DispatchStatusRequest{SessionID: args[0]}, &out); err != nil {
			return err
		}
		q := strings.ReplaceAll(out.Cwd, "'", `'\''`)
		fmt.Printf("cd '%s' && claude --resume %s\n", q, out.ID)
		return nil
	},
}
