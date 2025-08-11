package clock

import (
	"time"

	"auction-microservice/internal/ports"
)

// RealClock implements the Clock interface using the standard time package
type RealClock struct{}

// NewRealClock creates a new RealClock instance
func NewRealClock() ports.Clock {
	return &RealClock{}
}

// Now returns the current time
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// After returns a channel that sends the current time after the specified duration
func (c *RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// Sleep pauses the current goroutine for at least the duration d
func (c *RealClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Since returns the time elapsed since t
func (c *RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// Until returns the duration until t
func (c *RealClock) Until(t time.Time) time.Duration {
	return time.Until(t)
}