package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/client"
	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

var cancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a running or queued session",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		c, err := client.FromConfigFile()
		if err != nil {
			return err
		}
		var out daemon.DispatchCancelResponse
		if err := c.RPC("dispatch_cancel",
			daemon.DispatchCancelRequest{SessionID: args[0]}, &out); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		if !out.Killed {
			os.Exit(1)
		}
		return nil
	},
}
