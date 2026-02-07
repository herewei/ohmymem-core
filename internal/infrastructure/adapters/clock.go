package adapters

import (
	"time"

	"github.com/herewei/ohmymem-core/internal/domain"
)

// SystemClock provides current time using system clock
type SystemClock struct{}

// NewSystemClock creates a new system clock adapter
func NewSystemClock() domain.TimeProvider {
	return &SystemClock{}
}

// Now returns the current time
func (c *SystemClock) Now() time.Time {
	return time.Now()
}

// Ensure SystemClock implements TimeProvider
var _ domain.TimeProvider = (*SystemClock)(nil)
