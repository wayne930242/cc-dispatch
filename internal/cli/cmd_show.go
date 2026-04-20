package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/client"
	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show session detail",
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
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}
