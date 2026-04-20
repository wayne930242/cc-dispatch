package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/client"
	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

var (
	tailSource string
	tailLines  int
)

var tailCmd = &cobra.Command{
	Use:   "tail <id>",
	Short: "Tail session log",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		c, err := client.FromConfigFile()
		if err != nil {
			return err
		}
		var out daemon.DispatchTailResponse
		if err := c.RPC("dispatch_tail", daemon.DispatchTailRequest{
			SessionID: args[0], Source: tailSource, Lines: tailLines,
		}, &out); err != nil {
			return err
		}
		for _, l := range out.Lines {
			fmt.Println(l)
		}
		if out.Truncated {
			fmt.Fprintln(os.Stderr, "(older lines omitted)")
		}
		return nil
	},
}

func init() {
	tailCmd.Flags().StringVarP(&tailSource, "source", "S", "jsonl", "jsonl|stdout|stderr")
	tailCmd.Flags().IntVarP(&tailLines, "lines", "n", 50, "default 50")
}
