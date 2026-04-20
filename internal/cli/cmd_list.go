package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/client"
	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

var (
	listWorkspace string
	listStatus    string
	listLimit     int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List dispatched sessions",
	RunE: func(_ *cobra.Command, _ []string) error {
		c, err := client.FromConfigFile()
		if err != nil {
			return err
		}
		var out daemon.DispatchListResponse
		if err := c.RPC("dispatch_list", daemon.DispatchListRequest{
			Workspace: listWorkspace, Status: listStatus, Limit: listLimit,
		}, &out); err != nil {
			return err
		}
		if len(out.Sessions) == 0 {
			fmt.Println("(no sessions)")
			return nil
		}
		fmt.Println("session_id\tworkspace\tapp\tstatus\tcreated")
		for _, s := range out.Sessions {
			short := s.ID
			if len(short) > 8 {
				short = short[:8]
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", short, s.Workspace, s.App, s.Status,
				time.UnixMilli(s.CreatedAt).UTC().Format(time.RFC3339))
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "")
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 50, "default 50")
}
