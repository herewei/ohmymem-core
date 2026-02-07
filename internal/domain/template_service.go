package domain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateService handles template loading and memory generation
type TemplateService struct {
	repo   TemplateRepository
	loader TemplateLoader
}

// NewTemplateService creates a new TemplateService
func NewTemplateService(repo TemplateRepository, loader TemplateLoader) *TemplateService {
	return &TemplateService{
		repo:   repo,
		loader: loader,
	}
}

// InitTemplate generates a complete memory.md content based on project info
func (s *TemplateService) InitTemplate(ctx context.Context, info *ProjectInfo, repoURLs []string) (string, string, error) {
	// Fetch templates from repository
	var (
		tempPath string
		err      error
	)
	switch len(repoURLs) {
	case 0:
		return "", "", fmt.Errorf("no template repository URL provided")
	case 1:
		// When the user specifies a single custom repo, do a direct fetch and surface the raw git error.
		tempPath, err = s.repo.Fetch(ctx, repoURLs[0])
		if err != nil {
			return "", "", err
		}
	default:
		tempPath, err = s.repo.FetchWithFallback(ctx, repoURLs)
		if err != nil {
			return "", "", fmt.Errorf("fetch templates: %w", err)
		}
	}
	defer s.repo.Cleanup(tempPath)

	// Load template
	template, err := s.loader.LoadTemplate(tempPath)
	if err != nil {
		return "", "", fmt.Errorf("load template: %w", err)
	}

	// Load agents content
	agentsContent, err := s.loader.LoadAgents(tempPath)
	if err != nil {
		return "", "", fmt.Errorf("load agents: %w", err)
	}

	// Generate memory content with project-specific sections
	memoryContent := s.generateMemoryContent(template, info)

	return memoryContent, agentsContent, nil
}

// generateMemoryContent generates the final memory.md content
func (s *TemplateService) generateMemoryContent(template *Template, info *ProjectInfo) string {
	var sb strings.Builder

	// Frontmatter
	sb.WriteString("---\n")
	sb.WriteString("schema_version: \"0.1\"\n")
	sb.WriteString("entry_format: \"anchored\"\n")
	sb.WriteString(fmt.Sprintf("generated_by: \"ohmymem init\"\n"))

	if info != nil && info.IsDetected() {
		sb.WriteString("detected_stack:\n")
		sb.WriteString(fmt.Sprintf("  language: \"%s\"\n", info.Language))
		if info.Framework != "" {
			sb.WriteString(fmt.Sprintf("  framework: \"%s\"\n", info.Framework))
		}
		if info.Database != "" {
			sb.WriteString(fmt.Sprintf("  database: \"%s\"\n", info.Database))
		}
	}

	sb.WriteString("---\n\n")

	// Merge all template content
	sections := map[string][]string{
		"Constraints":   {},
		"Decisions":     {},
		"Patterns":      {},
		"Anti-Patterns": {},
	}

	// Collect content from all template files
	for _, file := range template.MemoryFiles {
		section := s.mapCategoryToSection(file.Category)
		if section != "" {
			sections[section] = append(sections[section], file.Content)
		}
	}

	// Write sections in order
	for _, sectionName := range []string{"Constraints", "Decisions", "Patterns", "Anti-Patterns"} {
		sb.WriteString(fmt.Sprintf("## %s\n\n", sectionName))
		for _, content := range sections[sectionName] {
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// mapCategoryToSection maps template categories to memory sections
func (s *TemplateService) mapCategoryToSection(category string) string {
	// For now, all content goes into appropriate sections based on content type
	// This can be extended based on meta.yaml parsing
	switch category {
	case "constraints":
		return "Constraints"
	case "decisions":
		return "Decisions"
	case "patterns":
		return "Patterns"
	case "anti-patterns":
		return "Anti-Patterns"
	default:
		return "Constraints" // Default section
	}
}

// LocalTemplateLoader implements TemplateLoader for local filesystem
type LocalTemplateLoader struct{}

// NewLocalTemplateLoader creates a new LocalTemplateLoader
func NewLocalTemplateLoader() *LocalTemplateLoader {
	return &LocalTemplateLoader{}
}

// LoadTemplate loads template files from a local repository path
func (l *LocalTemplateLoader) LoadTemplate(basePath string) (*Template, error) {
	template := &Template{
		BasePath:    basePath,
		MemoryFiles: []MemoryTemplateFile{},
	}

	// Load base/common template
	baseDir := filepath.Join(basePath, "bases", "common")
	if err := l.loadMemoryFilesFromDir(baseDir, "common", &template.MemoryFiles); err != nil {
		// Base template is optional
	}

	return template, nil
}

// LoadAgents loads the agents.md content from repository
func (l *LocalTemplateLoader) LoadAgents(basePath string) (string, error) {
	agentsPath := filepath.Join(basePath, "agents.md")
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default agents content if file doesn't exist
			return l.getDefaultAgentsContent(), nil
		}
		return "", fmt.Errorf("read agents.md: %w", err)
	}
	return string(content), nil
}

// loadMemoryFilesFromDir loads all memory.md files from a directory
func (l *LocalTemplateLoader) loadMemoryFilesFromDir(dir, source string, files *[]MemoryTemplateFile) error {
	memoryPath := filepath.Join(dir, "memory.md")
	content, err := os.ReadFile(memoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	*files = append(*files, MemoryTemplateFile{
		Path:     memoryPath,
		Content:  string(content),
		Source:   source,
		Category: "constraints", // Default category
	})

	return nil
}

// getDefaultAgentsContent returns default agents content
func (l *LocalTemplateLoader) getDefaultAgentsContent() string {
	return `### Boot Protocol

At the **START** of every conversation:
1. Call ` + "`ohmymem_read`" + ` tool to load project constraints
2. Review all Constraints before writing any code
3. Apply Patterns to maintain consistency

### Memory Protocol

When you identify important information:

| Type | When to Record |
|------|----------------|
| **Constraint** | Technical requirements, must/must-not rules |
| **Decision** | Architecture choices with rationale |
| **Pattern** | Code style, naming conventions |
| **Anti-Pattern** | Failed approaches, things to avoid |

Use ` + "`ohmymem_capture`" + ` tool with appropriate category and tags.

### Enforcement

- **Constraints** are non-negotiable. STOP and clarify if request conflicts.
- **Patterns** should be followed for consistency.
- **Anti-Patterns** are warnings. Suggest alternatives if user requests them.
`
}
