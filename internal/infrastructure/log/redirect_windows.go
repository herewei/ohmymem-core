//go:build windows

package log

import (
	"os"

	"golang.org/x/sys/windows"
)

// redirectStderr redirects stderr to the given file.
// On Windows, requires two steps:
// 1. SetStdHandle modifies the Windows CRT handle
// 2. os.Stderr sync ensures Go runtime uses the new handle
func redirectStderr(file *os.File) error {
	// Step 1: Modify Windows CRT standard handle
	// This affects C runtime libraries and spawned child processes
	if err := windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(file.Fd())); err != nil {
		return err
	}

	// Step 2: Sync Go runtime's os.Stderr
	// Critical: without this, fmt/log/slog still write to original stderr
	os.Stderr = file

	return nil
}
