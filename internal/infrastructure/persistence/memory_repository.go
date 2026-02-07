package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"log/slog"

	"github.com/gofrs/flock"

	"github.com/herewei/ohmymem-core/internal/domain"
)

const (
	DirName  = ".ohmymem"
	FileName = "memory.md"
)

// MarkdownMemoryRepository implements MemoryRepository using Markdown file-based storage with flock
type MarkdownMemoryRepository struct {
	basePath      string
	uuidGenerator domain.UUIDGenerator
	timeProvider  domain.TimeProvider
	lockFilePath  string
}

// NewMemoryRepository creates a new Markdown-based memory repository
func NewMemoryRepository(basePath string, uuidGenerator domain.UUIDGenerator, timeProvider domain.TimeProvider) *MarkdownMemoryRepository {
	return &MarkdownMemoryRepository{
		basePath:      basePath,
		uuidGenerator: uuidGenerator,
		timeProvider:  timeProvider,
		lockFilePath:  filepath.Join(basePath, DirName, ".memory.lock"),
	}
}

// FilePath returns the full path to the memory file
func (r *MarkdownMemoryRepository) FilePath() string {
	return filepath.Join(r.basePath, DirName, FileName)
}

// DirPath returns the full path to the memory directory
func (r *MarkdownMemoryRepository) DirPath() string {
	return filepath.Join(r.basePath, DirName)
}

// EnsureDir creates the memory directory if it doesn't exist
func (r *MarkdownMemoryRepository) EnsureDir() error {
	return os.MkdirAll(r.DirPath(), 0755)
}

// acquireLock acquires an exclusive file lock using flock
func (r *MarkdownMemoryRepository) acquireLock(ctx context.Context) (func() error, error) {
	// Ensure lock file exists
	if err := r.EnsureDir(); err != nil {
		return nil, err
	}

	fl := flock.New(r.lockFilePath)

	// Try to acquire lock with context support
	locked := make(chan struct{})
	var lockErr error

	go func() {
		lockErr = fl.Lock()
		close(locked)
	}()

	select {
	case <-ctx.Done():
		fl.Unlock()
		return nil, ctx.Err()
	case <-locked:
		if lockErr != nil {
			return nil, fmt.Errorf("failed to acquire lock: %w", lockErr)
		}
		return fl.Unlock, nil
	}
}

// readFile reads the entire memory file
func (r *MarkdownMemoryRepository) readFile() (string, error) {
	if err := r.EnsureDir(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(r.FilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read memory file: %w", err)
	}

	return string(data), nil
}

// atomicWrite writes content atomically using rename
func (r *MarkdownMemoryRepository) atomicWrite(content string) error {
	tmpPath := r.FilePath() + ".tmp"

	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, r.FilePath()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// ReadAll implements MemoryRepository
func (r *MarkdownMemoryRepository) ReadAll(ctx context.Context) (string, error) {
	return r.readFile()
}

// GetSection implements MemoryRepository
func (r *MarkdownMemoryRepository) GetSection(ctx context.Context, sectionType domain.SectionType) (*domain.Section, error) {
	content, err := r.readFile()
	if err != nil {
		return nil, err
	}

	if content == "" {
		return &domain.Section{
			Type:    sectionType,
			Entries: []domain.Entry{},
		}, nil
	}

	// Extract section content
	sectionBlock := extractSection(content, string(sectionType))

	if sectionBlock == "" {
		return &domain.Section{
			Type:    sectionType,
			Entries: []domain.Entry{},
		}, nil
	}

	// Try V1 anchored format first
	entries, err := parseV1Anchored(sectionBlock)
	if err == nil {
		return &domain.Section{
			Type:    sectionType,
			Entries: entries,
		}, nil
	}

	// Fallback to legacy format
	legacyEntries := parseLegacyInline(sectionBlock)
	if len(legacyEntries) > 0 {
		slog.Warn("legacy format detected", "section", sectionType)
		return &domain.Section{
			Type:    sectionType,
			Entries: legacyEntries,
		}, nil
	}

	return &domain.Section{
		Type:    sectionType,
		Entries: []domain.Entry{},
	}, nil
}

// AppendEntry implements MemoryRepository with flock
func (r *MarkdownMemoryRepository) AppendEntry(ctx context.Context, sectionType domain.SectionType, entry *domain.Entry) error {
	// Acquire exclusive lock
	unlock, err := r.acquireLock(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := unlock(); err != nil {
			slog.Error("failed to unlock file", "error", err)
		}
	}()

	// Read current content
	content, err := r.readFile()
	if err != nil {
		return err
	}

	// Check if file needs initialization
	if content == "" {
		content = r.createInitialContent()
	}

	// Render the new entry
	renderedEntry := renderEntry(entry)

	// Find section position
	section := capitalize(string(sectionType))
	sectionStart := findSectionStart(content, section)
	if sectionStart == -1 {
		return fmt.Errorf("section not found: %s", section)
	}

	// Find section end
	sectionEnd := findSectionEnd(content, sectionStart)

	// Build new content
	var newContent strings.Builder
	newContent.WriteString(content[:sectionStart])
	newContent.WriteString(content[sectionStart:sectionEnd])
	newContent.WriteString(renderedEntry)
	newContent.WriteString("\n")
	newContent.WriteString(content[sectionEnd:])

	// Atomic write
	if err := r.atomicWrite(newContent.String()); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	slog.Debug("entry appended successfully",
		"section", sectionType,
		"tag", entry.Tag,
		"id", entry.ID)

	return nil
}

// createInitialContent creates a new memory file with front matter
func (r *MarkdownMemoryRepository) createInitialContent() string {
	now := r.timeProvider.Now().Format(time.RFC3339)
	return fmt.Sprintf(`---
schema_version: "0.1"
entry_format: "anchored"
created_at: "%s"
---

## Constraints

## Decisions

## Patterns

## Anti-Patterns
`, now)
}

// Helper functions

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func findSectionStart(content, section string) int {
	header := fmt.Sprintf("## %s", section)
	return strings.Index(content, header)
}

func findSectionEnd(content string, start int) int {
	remaining := content[start+3:]
	lines := strings.Split(remaining, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			pos := strings.Index(remaining, line)
			return start + 3 + pos
		}
	}

	return len(content)
}

func extractSection(content, sectionType string) string {
	section := capitalize(sectionType)
	start := findSectionStart(content, section)
	if start == -1 {
		return ""
	}

	end := findSectionEnd(content, start)
	if end == -1 {
		return ""
	}

	sectionContent := content[start:end]
	header := fmt.Sprintf("## %s\n", section)
	return strings.TrimPrefix(sectionContent, header)
}

// V1 Parser (anchored format)
var anchoredEntryRegex = regexp.MustCompile(
	`(?m)^<!-- entry-id: ([a-f0-9-]+), tag: \[([^\]]+)\], time: ([^\n]+) -->` + "\n" +
		`^\* \*\*\[([^\]]+)\]\*\* (.+?)(?: ` + regexp.QuoteMeta("(*Rationale:") + `(.+?)` + regexp.QuoteMeta("*)") + `)?` + "\n" +
		`^<!-- entry-end -->$`,
)

func parseV1Anchored(block string) ([]domain.Entry, error) {
	matches := anchoredEntryRegex.FindAllStringSubmatch(block, -1)
	if matches == nil {
		return nil, fmt.Errorf("no anchored entries found")
	}

	var entries []domain.Entry
	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		tag := match[2]
		entries = append(entries, domain.Entry{
			ID:        match[1],
			Tag:       "[" + tag + "]",
			TagName:   tag,
			Content:   match[3],
			Rationale: match[4],
		})
	}

	return entries, nil
}

// Legacy Parser (inline format)
var legacyEntryRegex = regexp.MustCompile(
	`^\* \*\*\[([^\]]+)\]\*\* (.+?)(?: ` + regexp.QuoteMeta("(*Rationale:") + `(.+?)` + regexp.QuoteMeta("*)") + `)?$`,
)

func parseLegacyInline(block string) []domain.Entry {
	lines := strings.Split(block, "\n")
	var entries []domain.Entry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "* **[") {
			if match := legacyEntryRegex.FindStringSubmatch(line); match != nil {
				entries = append(entries, domain.Entry{
					Tag:       "[" + match[1] + "]",
					TagName:   match[1],
					Content:   match[2],
					Rationale: match[3],
				})
			}
		}
	}

	return entries
}

// renderEntry renders an entry to the 4-line anchored format
func renderEntry(entry *domain.Entry) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("<!-- entry-id: %s, tag: %s, time: %s -->\n",
		entry.ID, entry.Tag, entry.CreatedAt.Format(time.RFC3339)))

	buf.WriteString(fmt.Sprintf("* **[%s]** %s", entry.TagName, entry.Content))

	if entry.Rationale != "" {
		buf.WriteString(fmt.Sprintf(" (*Rationale: %s*)", entry.Rationale))
	}

	buf.WriteString("\n<!-- entry-end -->")

	return buf.String()
}
