package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/herewei/ohmymem-core/internal/domain"
	"github.com/herewei/ohmymem-core/internal/infrastructure/template"
)

type InitUseCase struct {
	detector domain.ProjectDetector
	template *domain.TemplateService
}

func NewInitUseCase(detector domain.ProjectDetector) *InitUseCase {
	return &InitUseCase{
		detector: detector,
	}
}

func NewInitUseCaseWithTemplate(detector domain.ProjectDetector, template *domain.TemplateService) *InitUseCase {
	return &InitUseCase{
		detector: detector,
		template: template,
	}
}

// InitOptions init command options
type InitOptions struct {
	RootPath    string              // Project root directory
	ProjectInfo *domain.ProjectInfo // Optional: reuse detected project info
	Force       bool                // Force overwrite
	Yes         bool                // Skip confirmation
	RepoURLs    []string            // Custom template repository URLs
}

// InitResult init result
type InitResult struct {
	ProjectInfo  *domain.ProjectInfo
	CreatedFiles []string
	Warnings     []string
}

// Preview prepares init by detecting project.
// It does not create or modify any files.
func (uc *InitUseCase) Preview(opts InitOptions) (*InitResult, error) {
	result := &InitResult{}

	info, err := uc.resolveAndDetect(opts)
	if err != nil {
		return nil, err
	}

	result.ProjectInfo = info

	return result, nil
}

// Execute executes the init command
func (uc *InitUseCase) Execute(opts InitOptions) (*InitResult, error) {
	result := &InitResult{}

	// 1. Check if already initialized
	memoryPath := filepath.Join(opts.RootPath, ".ohmymem", "memory.md")
	if fileExists(memoryPath) && !opts.Force {
		return nil, fmt.Errorf("already initialized. Use '--force' to overwrite")
	}

	// 2. Resolve and detect project (or reuse provided info)
	info, err := uc.resolveAndDetect(opts)
	if err != nil {
		return nil, err
	}
	result.ProjectInfo = info

	// 3. Generate memory content from templates
	ctx := context.Background()
	repoURLs := opts.RepoURLs
	if len(repoURLs) == 0 {
		repoURLs = template.GetDefaultRepoURLs()
	}

	memoryContent, agentsContent, err := uc.template.InitTemplate(ctx, info, repoURLs)
	if err != nil {
		// If user specifies a single custom repo, surface the raw git error directly.
		if len(opts.RepoURLs) == 1 {
			return nil, err
		}
		return nil, fmt.Errorf("generate template: %w", err)
	}

	// 4. Create .ohmymem directory
	ohmymemDir := filepath.Join(opts.RootPath, ".ohmymem")
	if err := os.MkdirAll(ohmymemDir, 0755); err != nil {
		return nil, fmt.Errorf("create .ohmymem directory: %w", err)
	}

	// 5. Write memory.md
	if err := os.WriteFile(memoryPath, []byte(memoryContent), 0644); err != nil {
		return nil, fmt.Errorf("write memory.md: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, memoryPath)

	// 6. Write/Update AGENTS.md
	agentsPath := filepath.Join(opts.RootPath, "AGENTS.md")
	if err := uc.updateAgentsFile(agentsPath, agentsContent); err != nil {
		return nil, fmt.Errorf("update AGENTS.md: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, agentsPath)

	// 7. Create symlinks
	symlinks := map[string]string{
		".cursorrules": "AGENTS.md",
		"CLAUDE.md":    "AGENTS.md",
	}

	for link, target := range symlinks {
		linkPath := filepath.Join(opts.RootPath, link)
		if err := uc.createSymlink(linkPath, target); err != nil {
			// Symlink failure is not fatal, record warning for caller
			result.Warnings = append(result.Warnings, fmt.Sprintf("Warning: failed to create symlink %s: %v", link, err))
		} else {
			result.CreatedFiles = append(result.CreatedFiles, linkPath+" → "+target)
		}
	}

	return result, nil
}

func (uc *InitUseCase) resolveAndDetect(opts InitOptions) (*domain.ProjectInfo, error) {
	// Initialize template service if not provided
	if uc.template == nil {
		uc.initDefaultTemplateService()
	}

	// Detect project (or reuse provided info)
	if opts.ProjectInfo != nil {
		return opts.ProjectInfo, nil
	}

	info, err := uc.detector.Detect(opts.RootPath)
	if err != nil {
		return nil, fmt.Errorf("detect project: %w", err)
	}

	return info, nil
}

// SetTemplateService sets the template service (for testing)
func (uc *InitUseCase) SetTemplateService(svc *domain.TemplateService) {
	uc.template = svc
}

// initDefaultTemplateService initializes the default template service
func (uc *InitUseCase) initDefaultTemplateService() {
	repo := template.NewGitRepoTemplateRepository(template.DefaultFetchTimeout)
	loader := domain.NewLocalTemplateLoader()
	uc.template = domain.NewTemplateService(repo, loader)
}

// updateAgentsFile updates or creates AGENTS.md
func (uc *InitUseCase) updateAgentsFile(path, agentsContent string) error {
	content := ""

	if fileExists(path) {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content = string(data)
	}

	// Build ohmymem block
	block := fmt.Sprintf(`<!-- ohmymem:start -->
<!-- 
  This section is managed by OhMyMem.
  Manual edits within this block may be overwritten.
  Last updated: %s
-->

%s
<!-- ohmymem:end -->`, time.Now().Format(time.RFC3339), agentsContent)

	// Check if ohmymem block already exists
	if strings.Contains(content, "<!-- ohmymem:start -->") {
		// Replace existing block
		startIdx := strings.Index(content, "<!-- ohmymem:start -->")
		endIdx := strings.Index(content, "<!-- ohmymem:end -->")
		if startIdx != -1 && endIdx != -1 {
			endIdx += len("<!-- ohmymem:end -->")
			content = content[:startIdx] + block + content[endIdx:]
		}
	} else {
		// Append to file
		if content != "" && !strings.HasSuffix(content, "\n\n") {
			if strings.HasSuffix(content, "\n") {
				content += "\n"
			} else {
				content += "\n\n"
			}
		}
		content += block
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// createSymlink creates a symbolic link
func (uc *InitUseCase) createSymlink(linkPath, target string) error {
	// If already exists
	if info, err := os.Lstat(linkPath); err == nil {
		// If it's a symlink, check target
		if info.Mode()&os.ModeSymlink != 0 {
			existingTarget, _ := os.Readlink(linkPath)
			if existingTarget == target {
				return nil // Already correctly linked
			}
			// Remove old link
			os.Remove(linkPath)
		} else {
			// It's a real file, don't overwrite
			return fmt.Errorf("file exists and is not a symlink")
		}
	}

	return os.Symlink(target, linkPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
