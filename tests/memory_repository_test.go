package main_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/herewei/ohmymem-core/internal/domain"
	"github.com/herewei/ohmymem-core/internal/infrastructure/persistence"
)

type testClock struct {
	currentTime time.Time
}

func (c *testClock) Now() time.Time {
	if c.currentTime.IsZero() {
		c.currentTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	}
	return c.currentTime
}

type testUUID struct {
	counter int
}

func (u *testUUID) NewV7() (string, error) {
	u.counter++
	return "test-uuid-" + string(rune('0'+u.counter)), nil
}

func setupTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "ohmymem-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return tmpDir
}

func TestNewMemoryRepository(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	uuidGen := &testUUID{}
	timeProvider := &testClock{}

	repo := persistence.NewMemoryRepository(tmpDir, uuidGen, timeProvider)

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}

	expectedPath := filepath.Join(tmpDir, ".ohmymem", "memory.md")
	if repo.FilePath() != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, repo.FilePath())
	}
}

func TestMemoryRepository_ReadAll_Empty(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := persistence.NewMemoryRepository(tmpDir, &testUUID{}, &testClock{})

	content, err := repo.ReadAll(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty content, got %s", content)
	}
}

func TestMemoryRepository_GetSection_Empty(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := persistence.NewMemoryRepository(tmpDir, &testUUID{}, &testClock{})

	section, err := repo.GetSection(context.Background(), domain.SectionConstraints)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if section.Type != domain.SectionConstraints {
		t.Errorf("expected section type constraints, got %s", section.Type)
	}
	if len(section.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(section.Entries))
	}
}

func TestMemoryRepository_AppendEntry(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	uuidGen := &testUUID{}
	timeProvider := &testClock{}
	repo := persistence.NewMemoryRepository(tmpDir, uuidGen, timeProvider)

	entry := &domain.Entry{
		ID:        "test-uuid-1",
		Tag:       "[Architecture]",
		TagName:   "Architecture",
		Content:   "Use hexagonal architecture",
		Rationale: "For separation of concerns",
		CreatedAt: timeProvider.Now(),
	}

	err := repo.AppendEntry(context.Background(), domain.SectionConstraints, entry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify file was created
	filePath := repo.FilePath()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected memory.md to be created")
	}

	// Verify section now has entry
	section, err := repo.GetSection(context.Background(), domain.SectionConstraints)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(section.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(section.Entries))
	}
	if section.Entries[0].Tag != "[Architecture]" {
		t.Errorf("expected tag [Architecture], got %s", section.Entries[0].Tag)
	}
}

func TestMemoryRepository_AppendMultipleEntries(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	uuidGen := &testUUID{}
	timeProvider := &testClock{}
	repo := persistence.NewMemoryRepository(tmpDir, uuidGen, timeProvider)

	entries := []*domain.Entry{
		{
			ID:        "test-uuid-1",
			Tag:       "[Architecture]",
			TagName:   "Architecture",
			Content:   "Use hexagonal architecture",
			CreatedAt: timeProvider.Now(),
		},
		{
			ID:        "test-uuid-2",
			Tag:       "[Storage]",
			TagName:   "Storage",
			Content:   "Use Markdown for storage",
			CreatedAt: timeProvider.Now(),
		},
	}

	for _, entry := range entries {
		err := repo.AppendEntry(context.Background(), domain.SectionConstraints, entry)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	section, err := repo.GetSection(context.Background(), domain.SectionConstraints)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(section.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(section.Entries))
	}
}

func TestMemoryRepository_GetSection_MultipleSections(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	uuidGen := &testUUID{}
	timeProvider := &testClock{}
	repo := persistence.NewMemoryRepository(tmpDir, uuidGen, timeProvider)

	// Add to constraints
	constraintEntry := &domain.Entry{
		ID:      "test-uuid-1",
		Tag:     "[Architecture]",
		TagName: "Architecture",
		Content: "Use hexagonal architecture",
	}
	err := repo.AppendEntry(context.Background(), domain.SectionConstraints, constraintEntry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Add to decisions
	decisionEntry := &domain.Entry{
		ID:      "test-uuid-2",
		Tag:     "[Storage]",
		TagName: "Storage",
		Content: "Use Markdown for storage",
	}
	err = repo.AppendEntry(context.Background(), domain.SectionDecisions, decisionEntry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify constraints section
	constraints, err := repo.GetSection(context.Background(), domain.SectionConstraints)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(constraints.Entries) != 1 {
		t.Errorf("expected 1 constraint entry, got %d", len(constraints.Entries))
	}

	// Verify decisions section
	decisions, err := repo.GetSection(context.Background(), domain.SectionDecisions)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(decisions.Entries) != 1 {
		t.Errorf("expected 1 decision entry, got %d", len(decisions.Entries))
	}

	// Verify patterns section is empty
	patterns, err := repo.GetSection(context.Background(), domain.SectionPatterns)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(patterns.Entries) != 0 {
		t.Errorf("expected 0 pattern entries, got %d", len(patterns.Entries))
	}
	// Verify anti-patterns section is empty
	anti_patterns, err := repo.GetSection(context.Background(), domain.SectionAntiPatterns)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(anti_patterns.Entries) != 0 {
		t.Errorf("expected 0 anti-pattern entries, got %d", len(anti_patterns.Entries))
	}
}

func TestMemoryRepository_ReadAll_WithContent(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	uuidGen := &testUUID{}
	timeProvider := &testClock{}
	repo := persistence.NewMemoryRepository(tmpDir, uuidGen, timeProvider)

	// Add an entry
	entry := &domain.Entry{
		ID:        "test-uuid-1",
		Tag:       "[Test]",
		TagName:   "Test",
		Content:   "Test content",
		CreatedAt: timeProvider.Now(),
	}
	err := repo.AppendEntry(context.Background(), domain.SectionConstraints, entry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Read all content
	content, err := repo.ReadAll(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify content contains expected sections
	if !strings.Contains(content, "## Constraints") {
		t.Error("expected content to contain ## Constraints section")
	}
	if !strings.Contains(content, "## Decisions") {
		t.Error("expected content to contain ## Decisions section")
	}
	if !strings.Contains(content, "## Patterns") {
		t.Error("expected content to contain ## Patterns section")
	}
	if !strings.Contains(content, "## Anti-Patterns") {
		t.Error("expected content to contain ## Anti-Patterns section")
	}
	if !strings.Contains(content, "[Test]") {
		t.Error("expected content to contain [Test] tag")
	}
}

func TestMemoryRepository_DirPath(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer os.RemoveAll(tmpDir)

	repo := persistence.NewMemoryRepository(tmpDir, &testUUID{}, &testClock{})

	expectedDir := filepath.Join(tmpDir, ".ohmymem")
	if repo.DirPath() != expectedDir {
		t.Errorf("expected dir path %s, got %s", expectedDir, repo.DirPath())
	}
}
