package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ccd",
	Short: "Dispatch headless Claude Code sessions",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(serveCmd, mcpCmd, listCmd, showCmd, tailCmd,
		cancelCmd, resumeCmd, stopCmd, versionCmd)
}
