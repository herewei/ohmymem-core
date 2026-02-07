package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"log/slog"
)

// Init initializes the logging system.
//
// Args:
//   - baseDir: directory for log files
//   - debug: enable debug mode (creates debug.log)
//
// Returns:
//   - cleanup: function to flush and close log files
//   - error: initialization error
func Init(baseDir string, debug bool) (cleanup func(), err error) {
	// Ensure log directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log dir: %w", err)
	}

	// Open error.log (always enabled)
	errLogPath := filepath.Join(baseDir, "error.log")
	errFile, err := os.OpenFile(errLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open error.log: %w", err)
	}

	// Configure slog
	var handler slog.Handler
	var cleanupFunc func()

	errWriter := io.MultiWriter(os.Stderr, errFile)
	errHandler := slog.NewJSONHandler(errWriter, &slog.HandlerOptions{
		Level: slog.LevelError,
	})

	if debug {
		// Debug mode: also write to debug.log
		debugLogPath := filepath.Join(baseDir, "debug.log")
		debugFile, err := os.OpenFile(debugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			errFile.Close()
			return nil, fmt.Errorf("failed to open debug.log: %w", err)
		}

		debugHandler := slog.NewJSONHandler(debugFile, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		handler = multiHandler{handlers: []slog.Handler{errHandler, debugHandler}}

		cleanupFunc = func() {
			_ = debugFile.Sync()
			_ = debugFile.Close()
			_ = errFile.Sync()
			_ = errFile.Close()
		}
	} else {
		// Non-debug: only errors to stderr + error.log
		handler = errHandler

		cleanupFunc = func() {
			_ = errFile.Sync()
			_ = errFile.Close()
		}
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return cleanupFunc, nil
}

type multiHandler struct {
	handlers []slog.Handler
}

func (h multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			_ = handler.Handle(ctx, record)
		}
	}
	return nil
}

func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return multiHandler{handlers: handlers}
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return multiHandler{handlers: handlers}
}
