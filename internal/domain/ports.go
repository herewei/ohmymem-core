package domain

import (
	"context"
	"time"
)

// MemoryRepository defines the storage interface
type MemoryRepository interface {
	// GetSection retrieves all entries from a specific section
	GetSection(ctx context.Context, sectionType SectionType) (*Section, error)

	// AppendEntry adds a new entry to the specified section
	AppendEntry(ctx context.Context, sectionType SectionType, entry *Entry) error

	// ReadAll returns the raw content of the entire memory file
	ReadAll(ctx context.Context) (string, error)

	// FilePath returns the path to the memory file
	FilePath() string
}

// UUIDGenerator interface for generating UUIDv7
type UUIDGenerator interface {
	NewV7() (string, error)
}

// TimeProvider interface for getting current time
type TimeProvider interface {
	Now() time.Time
}
