package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/config"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run:   func(_ *cobra.Command, _ []string) { fmt.Println(config.DaemonVersion) },
}
