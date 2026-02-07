package cmd

import (
	"fmt"
	"os"

	"github.com/herewei/ohmymem-core/internal/infrastructure/log"
	"github.com/herewei/ohmymem-core/internal/version"
	"github.com/spf13/cobra"
)

var logCleanup func()

var RootCmd = &cobra.Command{
	Use:     "ohmymem",
	Version: version.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Enable debug mode when environment variable is set
		debug := os.Getenv("OHMYMEM_DEBUG") == "true"

		// Initialize logging system
		cleanup, err := log.Init(".ohmymem", debug)
		if err != nil {
			// Log init failed; ensure user sees it on stderr.
			fmt.Fprintln(os.Stderr, "log initialization failed:", err)
			return fmt.Errorf("log initialization failed: %w", err)
		}
		logCleanup = cleanup
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Ensure all logs are flushed
		if logCleanup != nil {
			logCleanup()
		}
	},
	Short: "OhMyMem MCP Server - Memory storage for AI agents",
	Long: `OhMyMem is a lightweight MCP (Model Context Protocol) server 
providing file-based memory storage for AI agents.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
