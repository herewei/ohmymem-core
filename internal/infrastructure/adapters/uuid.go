package adapters

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/herewei/ohmymem-core/internal/domain"
)

// GoogleUUIDGenerator generates UUIDv7 using google/uuid library
type GoogleUUIDGenerator struct{}

// NewGoogleUUIDGenerator creates a new UUID v7 generator adapter
func NewGoogleUUIDGenerator() domain.UUIDGenerator {
	return &GoogleUUIDGenerator{}
}

// NewV7 generates a new UUID v7
func (g *GoogleUUIDGenerator) NewV7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		slog.Error("failed to generate UUID v7", "error", err)
		return "", err
	}
	return id.String(), nil
}

// Ensure GoogleUUIDGenerator implements UUIDGenerator
var _ domain.UUIDGenerator = (*GoogleUUIDGenerator)(nil)
