package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// CmdResult represents the result of a command execution
type CmdResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// runCmd executes the ohmymem binary with given arguments in the specified directory
func runCmd(dir string, args ...string) CmdResult {
	// Find the binary path - look in project root
	_, testFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(testFile), "..", "..")
	binaryPath := filepath.Join(projectRoot, "ohmymem")
	templateRepo := filepath.Join(projectRoot, "tests", "testdata", "templates")

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "OHMYMEM_TEMPLATE_REPO="+templateRepo)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return CmdResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
}

// setupTestEnv creates a test environment by copying a fixture to a temp directory
// If fixture is empty string, returns an empty temp directory
func setupTestEnv(t *testing.T, fixture string) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "ohmymem-e2e-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	if fixture != "" {
		_, testFile, _, _ := runtime.Caller(0)
		fixturePath := filepath.Join(filepath.Dir(testFile), "testdata", fixture)
		if err := copyDir(fixturePath, tmpDir); err != nil {
			t.Fatalf("failed to copy fixture %s: %v", fixture, err)
		}
	}

	// Register cleanup
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

// copyDir recursively copies a directory tree
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", dstPath, err)
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source %s: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dest %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}

	return nil
}

// assertFileExists checks if a file exists
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if err != nil {
		t.Errorf("file should exist: %s", path)
	}
}

// assertFileNotExists checks if a file does not exist
func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		t.Errorf("file should not exist: %s", path)
	}
}

// assertFileContains checks if a file contains a substring
func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("failed to read file %s: %v", path, err)
		return
	}
	if !bytes.Contains(content, []byte(substr)) {
		t.Errorf("file %s should contain %q, got:\n%s", path, substr, string(content))
	}
}

// assertSymlink checks if a path is a symlink pointing to the expected target
func assertSymlink(t *testing.T, link, target string) {
	t.Helper()
	actual, err := os.Readlink(link)
	if err != nil {
		t.Errorf("expected %s to be a symlink: %v", link, err)
		return
	}
	if actual != target {
		t.Errorf("symlink %s should point to %q, got %q", link, target, actual)
	}
}
