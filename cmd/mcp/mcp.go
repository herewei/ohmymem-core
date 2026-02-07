package mcp

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"github.com/herewei/ohmymem-core/cmd"
	mcpapp "github.com/herewei/ohmymem-core/internal/application/usecase"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func init() {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP server",
		Long: `Start the OhMyMem MCP server.
This server provides tools for reading and appending to the working memory file.`,
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}
	cmd.RootCmd.AddCommand(mcpCmd)
}

func runServer() {
	basePath := "."

	// Create MCP server and file store
	s, _, err := mcpapp.NewServer(basePath)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	// Start stdio server with graceful shutdown support
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)

	go func() {
		if err := server.ServeStdio(s); err != nil {
			if !errors.Is(err, context.Canceled) {
				errChan <- err
			}
		}
		cancel()
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-sigChan:
		// Interrupt signal received
		slog.Info("shutting down...")
		os.Exit(0)
	case <-ctx.Done():
		// Server stopped normally
		os.Exit(0)
	}
}
