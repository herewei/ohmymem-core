package init

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/herewei/ohmymem-core/cmd"
	initApp "github.com/herewei/ohmymem-core/internal/application/usecase"
	"github.com/herewei/ohmymem-core/internal/infrastructure/detector"
	"github.com/herewei/ohmymem-core/internal/infrastructure/huh"
	"github.com/herewei/ohmymem-core/internal/infrastructure/template"
)

var (
	initForce bool
	initYes   bool
	initRepo  string
)

func init() {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize OhMyMem in current project",
		Long:  "Initialize OhMyMem with smart detection of your project stack.",
		RunE:  runInit,
	}

	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing files")
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Skip confirmation prompts")
	initCmd.Flags().StringVar(&initRepo, "repo", "", "Custom template repository URL")

	cmd.RootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// 1. Get working directory
	rootPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// 1.1 Check if already initialized (interactive unless --yes or --force)
	memoryPath := filepath.Join(rootPath, ".ohmymem", "memory.md")
	if fileExists(memoryPath) && !initForce {
		if initYes {
			return fmt.Errorf("already initialized. Use '--force' to overwrite")
		}
		confirmed, err := huh.Confirm("Already initialized. Re-initialize?", true)
		if err != nil {
			if err == huh.ErrCancelled {
				fmt.Println("Cancelled.")
				return nil
			}
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			fmt.Println("Cancelled.")
			return nil
		}
		initForce = true
	}

	// 2. Create dependencies
	projectDetector := detector.NewCompositeDetector()
	iuc := initApp.NewInitUseCase(projectDetector)

	// 3. Resolve init mode (interactive by default)
	emptyProject := false
	var repoURLs []string
	if initRepo != "" {
		repo := strings.TrimSpace(initRepo)
		if repo != "" {
			repoURLs = []string{repo}
		}
	} else if !initYes {
		options := []string{
			"Empty project (no template)",
			"Use default template",
			"Custom template repo",
		}
		_, choice, err := huh.SelectOne("How would you like to initialize?", options)
		if err != nil {
			if err == huh.ErrCancelled {
				fmt.Println("Cancelled.")
				return nil
			}
			return fmt.Errorf("prompt failed: %w", err)
		}
		switch choice {
		case options[0]:
			emptyProject = true
		case options[1]:
			_, repoChoice, err := huh.SelectOne("Choose default template repo", []string{"GitHub", "Gitee"})
			if err != nil {
				if err == huh.ErrCancelled {
					fmt.Println("Cancelled.")
					return nil
				}
				return fmt.Errorf("prompt failed: %w", err)
			}
			if repoChoice == "Gitee" {
				repoURLs = []string{template.DefaultGiteeRepo}
			} else {
				repoURLs = []string{template.DefaultGitHubRepo}
			}
		case options[2]:
			repo, err := huh.PromptInput("Custom template repo URL", "")
			if err != nil {
				if err == huh.ErrCancelled {
					fmt.Println("Cancelled.")
					return nil
				}
				return fmt.Errorf("prompt failed: %w", err)
			}
			repo = strings.TrimSpace(repo)
			if repo == "" {
				fmt.Println("Cancelled.")
				return nil
			}
			repoURLs = []string{repo}
		}
	}

	opts := initApp.InitOptions{
		RootPath: rootPath,
		Force:    initForce,
		Yes:      initYes,
		RepoURLs: repoURLs,
	}

	if emptyProject {
		ohmymemDir := filepath.Join(rootPath, ".ohmymem")
		if err := os.MkdirAll(ohmymemDir, 0755); err != nil {
			return fmt.Errorf("create .ohmymem directory: %w", err)
		}
		if err := os.WriteFile(memoryPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("write memory.md: %w", err)
		}
		fmt.Println("✨ Creating files...")
		fmt.Println()
		fmt.Printf("   Created: %s\n", memoryPath)
		fmt.Println()
		fmt.Println("✅ Initialization complete!")
		fmt.Println()
		fmt.Println("   Next steps:")
		fmt.Println("   1. Review:  .ohmymem/memory.md")
		fmt.Println("   2. Config:  Add MCP server to your AI tool")
		fmt.Println("   3. Code:    Start with AI that remembers!")
		fmt.Println()
		fmt.Println("💡 Tips:")
		fmt.Println("   • Edit memory.md anytime with your editor")
		return nil
	}

	// 4. Preview (detect project)
	fmt.Println("🔍 Detecting project...")
	fmt.Println()

	preview, err := iuc.Preview(opts)
	if err != nil {
		return err
	}

	info := preview.ProjectInfo
	if info.IsDetected() {
		fmt.Printf("   Language:   %s\n", info.Language)
		if info.Framework != "" {
			fmt.Printf("   Framework:  %s\n", info.Framework)
		}
		if info.Database != "" {
			fmt.Printf("   Database:   %s\n", info.Database)
		}
		if info.ProjectType != "" {
			fmt.Printf("   Type:       %s\n", info.ProjectType)
		}
		fmt.Println()

		// 5. Confirm detection (unless --yes)
		if !initYes {
			confirmed, err := huh.Confirm("Is this correct?", true)
			if err != nil {
				if err == huh.ErrCancelled {
					fmt.Println("Cancelled.")
					return nil
				}
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				fmt.Println("Cancelled.")
				return nil
			}
		}
	} else {
		fmt.Println("   Could not detect project type.")
		fmt.Println()
	}

	opts.ProjectInfo = info

	result, err := iuc.Execute(opts)
	if err != nil {
		return err
	}

	// 7. Display results
	fmt.Println("✨ Creating files...")
	fmt.Println()
	for _, f := range result.CreatedFiles {
		fmt.Printf("   Created: %s\n", f)
	}
	fmt.Println()

	fmt.Println("✅ Initialization complete!")
	fmt.Println()
	fmt.Println("   Next steps:")
	fmt.Println("   1. Review:  .ohmymem/memory.md")
	fmt.Println("   2. Config:  Add MCP server to your AI tool")
	fmt.Println("   3. Code:    Start with AI that remembers!")
	fmt.Println()
	fmt.Println("💡 Tips:")
	fmt.Println("   • Edit memory.md anytime with your editor")
	if len(result.Warnings) > 0 {
		fmt.Println()
		for _, w := range result.Warnings {
			fmt.Println(w)
		}
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
