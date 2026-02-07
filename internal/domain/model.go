package domain

import "time"

// SectionType defines valid memory categories
type SectionType string

const (
	SectionConstraints  SectionType = "constraints"
	SectionDecisions    SectionType = "decisions"
	SectionPatterns     SectionType = "patterns"
	SectionAntiPatterns SectionType = "anti-patterns"
	SectionNote         SectionType = "note"
)

// ValidSections returns all valid section types
func ValidSections() []SectionType {
	return []SectionType{SectionConstraints, SectionDecisions, SectionPatterns, SectionAntiPatterns, SectionNote}
}

// IsValid checks if the section type is valid
func (s SectionType) IsValid() bool {
	switch s {
	case SectionConstraints, SectionDecisions, SectionPatterns, SectionAntiPatterns, SectionNote:
		return true
	default:
		return false
	}
}

// FileFormat represents the memory file format version
type FileFormat string

const (
	FormatAnchored     FileFormat = "anchored"
	FormatLegacyInline FileFormat = "legacy_inline"
)

// ParseMetadata contains metadata about the parsing process
type ParseMetadata struct {
	Format  FileFormat
	Warning string // Optional warning message
}

// Entry represents a single memory entry
type Entry struct {
	ID        string    // UUID v7
	Tag       string    // With brackets: "[Architecture]"
	TagName   string    // Without brackets: "Architecture"
	Content   string    // Cleaned single-line content
	Rationale string    // Optional
	CreatedAt time.Time // RFC3339 format
}

// Section represents a category of entries
type Section struct {
	Type    SectionType
	Entries []Entry
}

// AppendInput represents validated input for appending memory
type AppendInput struct {
	Category  string `json:"category" validate:"omitempty,oneof=constraints decisions patterns anti-patterns note"`
	Tag       string `json:"tag" validate:"required,max=50"`
	Content   string `json:"content" validate:"required,max=2000,ascii"`
	Rationale string `json:"rationale,omitempty" validate:"max=500"`
}
