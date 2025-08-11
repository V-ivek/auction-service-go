package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Clock defines the interface for time operations
type Clock interface {
	// Now returns the current time
	Now() time.Time

	// After returns a channel that sends the current time after the specified duration
	After(d time.Duration) <-chan time.Time

	// Sleep pauses the current goroutine for at least the duration d
	Sleep(d time.Duration)

	// Since returns the time elapsed since t
	Since(t time.Time) time.Duration

	// Until returns the duration until t
	Until(t time.Time) time.Duration
}

// TimeoutCallback is a function that gets called when a timeout occurs
type TimeoutCallback func(ctx context.Context, auctionID uuid.UUID)

// TimeoutManager defines the interface for managing auction timeouts
type TimeoutManager interface {
	// ScheduleTimeout schedules a timeout for an auction
	ScheduleTimeout(ctx context.Context, auctionID uuid.UUID, duration time.Duration, callback TimeoutCallback) error

	// CancelTimeout cancels a previously scheduled timeout
	CancelTimeout(auctionID uuid.UUID) error

	// UpdateTimeout updates the duration of an existing timeout (for auction extensions)
	UpdateTimeout(auctionID uuid.UUID, newDuration time.Duration) error

	// GetRemainingTime returns the remaining time for a timeout
	GetRemainingTime(auctionID uuid.UUID) (time.Duration, bool)

	// IsActive checks if a timeout is currently active
	IsActive(auctionID uuid.UUID) bool

	// Stop stops the timeout manager and cancels all pending timeouts
	Stop()
}