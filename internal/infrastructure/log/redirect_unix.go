//go:build !windows

package log

import (
	"os"

	"golang.org/x/sys/unix"
)

// redirectStderr redirects stderr to the given file.
// On Unix systems, uses Dup3 to replace FD 2 (stderr).
func redirectStderr(file *os.File) error {
	return unix.Dup2(int(file.Fd()), unix.Stderr)
}
