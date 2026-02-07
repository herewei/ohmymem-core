package template

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/herewei/ohmymem-core/internal/domain"
)

// Default template repository URLs
const (
	DefaultGitHubRepo = "https://github.com/herewei/ohmymem-templates.git"
	DefaultGiteeRepo  = "https://gitee.com/herewei/ohmymem-templates.git"
)

// DefaultFetchTimeout is the timeout for git clone operations
const DefaultFetchTimeout = 15 * time.Second

// GitRepoTemplateRepository implements domain.TemplateRepository using git clone
type GitRepoTemplateRepository struct {
	timeout time.Duration
}

// NewGitRepoTemplateRepository creates a new GitRepoTemplateRepository with the specified timeout
func NewGitRepoTemplateRepository(timeout time.Duration) *GitRepoTemplateRepository {
	return &GitRepoTemplateRepository{
		timeout: timeout,
	}
}

// IsGitAvailable checks if git command is available in the system
func (g *GitRepoTemplateRepository) IsGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// Fetch clones a template repository to a local temporary directory
// The caller is responsible for cleaning up the returned directory
func (g *GitRepoTemplateRepository) Fetch(ctx context.Context, repoURL string) (string, error) {
	if repoPath, ok := localRepoPath(repoURL); ok {
		return copyLocalRepo(repoPath)
	}

	if !g.IsGitAvailable() {
		return "", fmt.Errorf("git command not found, please install git")
	}

	// Create temporary directory for cloning
	tempDir, err := os.MkdirTemp("", "ohmymem-templates-*")
	if err != nil {
		return "", fmt.Errorf("create temp directory: %w", err)
	}

	// Create a context with timeout
	cloneCtx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// Build git clone command with --depth 1 for shallow clone
	cmd := exec.CommandContext(cloneCtx, "git", "clone", "--depth", "1", repoURL, tempDir)

	// Capture stderr for error reporting
	var stderr []byte
	cmd.Stderr = &stderrWriter{data: &stderr}

	if err := cmd.Run(); err != nil {
		// Clean up temp directory on failure
		os.RemoveAll(tempDir)

		if cloneCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git clone timeout after %v: %w", g.timeout, err)
		}

		stderrMsg := string(stderr)
		if stderrMsg != "" {
			return "", fmt.Errorf("git clone failed: %s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	return tempDir, nil
}

// FetchWithFallback tries multiple repository URLs in order
// Returns the local path from the first successful clone
func (g *GitRepoTemplateRepository) FetchWithFallback(ctx context.Context, repoURLs []string) (string, error) {
	if len(repoURLs) == 0 {
		return "", fmt.Errorf("no repository URLs provided")
	}

	var allErrors []error

	for _, repoURL := range repoURLs {
		path, err := g.Fetch(ctx, repoURL)
		if err == nil {
			return path, nil
		}
		allErrors = append(allErrors, fmt.Errorf("%s: %w", repoURL, err))
	}

	return "", &domain.TemplateFetchError{
		RepoURLs: repoURLs,
		Errors:   allErrors,
	}
}

// stderrWriter captures stderr output
type stderrWriter struct {
	data *[]byte
}

func (w *stderrWriter) Write(p []byte) (n int, err error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}

// Cleanup removes the temporary directory created by Fetch
func (g *GitRepoTemplateRepository) Cleanup(tempPath string) error {
	if tempPath == "" {
		return nil
	}
	return os.RemoveAll(tempPath)
}

// GetDefaultRepoURLs returns the default repository URLs in priority order
func GetDefaultRepoURLs() []string {
	return []string{
		DefaultGitHubRepo,
		DefaultGiteeRepo,
	}
}

func localRepoPath(repoURL string) (string, bool) {
	if repoURL == "" {
		return "", false
	}
	info, err := os.Stat(repoURL)
	if err != nil || !info.IsDir() {
		return "", false
	}
	return repoURL, true
}

func copyLocalRepo(src string) (string, error) {
	tempDir, err := os.MkdirTemp("", "ohmymem-templates-*")
	if err != nil {
		return "", fmt.Errorf("create temp directory: %w", err)
	}
	if err := copyDir(src, tempDir); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", err
	}
	return tempDir, nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := fmt.Sprintf("%s%c%s", src, os.PathSeparator, entry.Name())
		dstPath := fmt.Sprintf("%s%c%s", dst, os.PathSeparator, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", dstPath, err)
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

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
