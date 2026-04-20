package cli

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/daemon"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the cc-dispatch daemon (foreground)",
	RunE: func(_ *cobra.Command, _ []string) error {
		s, err := daemon.NewFromEnv()
		if err != nil {
			log.Fatal(err)
		}
		return s.Serve()
	},
}
