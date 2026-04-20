package cli

import (
	"github.com/spf13/cobra"
	"github.com/wayne930242/cc-dispatch/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run as an MCP stdio server (auto-spawns daemon if missing)",
	RunE: func(_ *cobra.Command, _ []string) error {
		return mcp.Run()
	},
}
