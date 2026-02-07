package domain

import (
	"context"
	"fmt"
)

// Template represents a loaded template from repository
type Template struct {
	BasePath    string // Local path where template is cloned
	MemoryFiles []MemoryTemplateFile
	AgentsFile  string // Path to agents.md
}

// MemoryTemplateFile represents a single memory template file
type MemoryTemplateFile struct {
	Path     string   // Relative path in template repo
	Content  string   // File content
	Tags     []string // e.g., ["go", "logging"]
	Source   string   // e.g., "languages/go"
	Category string   // e.g., "constraints", "patterns"
}

// TemplateRepository defines the port for fetching templates from remote repositories
type TemplateRepository interface {
	// Fetch clones a template repository to a local temporary directory
	// Returns the local path to the cloned repository
	// The caller is responsible for cleaning up the temporary directory
	Fetch(ctx context.Context, repoURL string) (string, error)

	// FetchWithFallback tries multiple repository URLs in order
	// Returns the local path from the first successful clone
	FetchWithFallback(ctx context.Context, repoURLs []string) (string, error)

	// IsGitAvailable checks if git command is available in the system
	IsGitAvailable() bool

	// Cleanup removes the temporary directory created by Fetch
	Cleanup(tempPath string) error
}

// TemplateFetchError represents an error when fetching templates
type TemplateFetchError struct {
	RepoURLs []string
	Errors   []error
}

func (e *TemplateFetchError) Error() string {
	return fmt.Sprintf("failed to fetch templates from all repositories: %v", e.Errors)
}

// TemplateLoader defines the port for loading templates from local path
type TemplateLoader interface {
	// LoadTemplate loads template from a local repository path
	LoadTemplate(basePath string) (*Template, error)

	// LoadAgents loads the agents.md content from repository
	LoadAgents(basePath string) (string, error)
}
