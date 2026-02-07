package domain

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"log/slog"
)

// MemoryService handles business logic and template rendering
type MemoryService struct {
	repo MemoryRepository
}

// NewMemoryService creates a new memory service
func NewMemoryService(repo MemoryRepository) *MemoryService {
	return &MemoryService{repo: repo}
}

// ValidateInput validates append input against schema constraints
func (s *MemoryService) ValidateInput(input AppendInput) error {
	// Default to "note" if category is empty
	if input.Category == "" {
		input.Category = "note"
	}

	// Validate Category
	sectionType := SectionType(input.Category)
	if !sectionType.IsValid() {
		return fmt.Errorf("%w: %s (must be constraints, decisions, patterns, anti-patterns or note)", ErrInvalidCategory, input.Category)
	}

	// Validate Tag
	if len(input.Tag) == 0 {
		return fmt.Errorf("%w: tag cannot be empty", ErrInvalidTag)
	}
	if len(input.Tag) > 50 {
		return fmt.Errorf("%w: tag must be 50 characters or less (got %d)", ErrInvalidTag, len(input.Tag))
	}

	// Validate Content
	if len(input.Content) == 0 {
		return fmt.Errorf("%w: content cannot be empty", ErrInvalidContent)
	}
	if len(input.Content) > 2000 {
		return fmt.Errorf("%w: content must be 2000 characters or less (got %d)", ErrInvalidContent, len(input.Content))
	}
	if err := ValidateContent(input.Content); err != nil {
		return err
	}

	// Validate Rationale
	if len(input.Rationale) > 500 {
		return fmt.Errorf("%w: rationale must be 500 characters or less (got %d)", ErrInvalidRationale, len(input.Rationale))
	}

	return nil
}

// ValidateContent checks for forbidden content patterns
func ValidateContent(content string) error {
	forbidden := []string{"\n", "\r", "###", "```", "<", ">"}
	for _, char := range forbidden {
		if strings.Contains(content, char) {
			return fmt.Errorf("%w: contains %q", ErrForbiddenContent, char)
		}
	}

	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return fmt.Errorf("%w: content cannot be a list item", ErrListItem)
	}

	return nil
}

// EntryView is the data structure for template rendering
type EntryView struct {
	ID        string
	Tag       string
	TagName   string
	Content   string
	Rationale string
	Time      string
}

const entryTemplate = `<!-- entry-id: {{.ID}}, tag: {{.Tag}}, time: {{.Time}} -->
* **[{{.TagName}}]** {{.Content}}{{if .Rationale}} (*Rationale: {{.Rationale}}*){{end}}
<!-- entry-end -->`

// RenderEntry renders an entry to the 4-line anchored format
func (s *MemoryService) RenderEntry(entry Entry) (string, error) {
	view := EntryView{
		ID:        entry.ID,
		Tag:       entry.Tag,
		TagName:   entry.TagName,
		Content:   entry.Content,
		Rationale: entry.Rationale,
		Time:      entry.CreatedAt.Format(time.RFC3339),
	}

	tmpl, err := template.New("entry").Parse(entryTemplate)
	if err != nil {
		slog.Error("failed to parse entry template", "error", err)
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, view); err != nil {
		slog.Error("failed to execute entry template", "error", err)
		return "", fmt.Errorf("failed to render entry: %w", err)
	}

	return buf.String(), nil
}

// PrepareEntry creates a new Entry from AppendInput
func (s *MemoryService) PrepareEntry(input AppendInput, id string, now time.Time) Entry {
	// Auto-wrap tag with brackets
	tag := input.Tag
	if !strings.HasPrefix(tag, "[") {
		tag = "[" + tag + "]"
	}

	// Remove brackets for TagName
	tagName := strings.Trim(tag, "[]")

	return Entry{
		ID:        id,
		Tag:       tag,
		TagName:   tagName,
		Content:   input.Content,
		Rationale: input.Rationale,
		CreatedAt: now,
	}
}

// ReadSection retrieves entries from a section
func (s *MemoryService) ReadSection(ctx context.Context, sectionType SectionType) (*Section, error) {
	return s.repo.GetSection(ctx, sectionType)
}

// GetMemoryPath returns the memory file path
func (s *MemoryService) GetMemoryPath() string {
	return s.repo.FilePath()
}

// ReadMemory returns raw content of the memory file
func (s *MemoryService) ReadMemory(ctx context.Context) (string, error) {
	return s.repo.ReadAll(ctx)
}

// AppendMemory appends an entry to the memory file
func (s *MemoryService) AppendMemory(ctx context.Context, input AppendInput, id string, now time.Time) error {
	entry := s.PrepareEntry(input, id, now)
	return s.repo.AppendEntry(ctx, SectionType(input.Category), &entry)
}
