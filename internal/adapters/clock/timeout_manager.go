package clock

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/internal/ports"
)

// timeoutEntry represents a single timeout entry
type timeoutEntry struct {
	auctionID uuid.UUID
	timer     *time.Timer
	callback  ports.TimeoutCallback
	deadline  time.Time
	ctx       context.Context
	cancel    context.CancelFunc
}

// TimeoutManager manages auction timeouts
type TimeoutManager struct {
	clock     ports.Clock
	logger    *zap.Logger
	timeouts  map[uuid.UUID]*timeoutEntry
	mu        sync.RWMutex
	stopCh    chan struct{}
	stopped   bool
}

// NewTimeoutManager creates a new TimeoutManager
func NewTimeoutManager(clock ports.Clock, logger *zap.Logger) ports.TimeoutManager {
	return &TimeoutManager{
		clock:    clock,
		logger:   logger,
		timeouts: make(map[uuid.UUID]*timeoutEntry),
		stopCh:   make(chan struct{}),
	}
}

// ScheduleTimeout schedules a timeout for an auction
func (tm *TimeoutManager) ScheduleTimeout(ctx context.Context, auctionID uuid.UUID, duration time.Duration, callback ports.TimeoutCallback) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.stopped {
		tm.logger.Debug("timeout manager is stopped, ignoring schedule request",
			zap.String("auction_id", auctionID.String()))
		return nil
	}

	// Cancel existing timeout if it exists
	if existing, exists := tm.timeouts[auctionID]; exists {
		existing.cancel()
		existing.timer.Stop()
	}

	// Create new timeout context
	timeoutCtx, cancel := context.WithCancel(ctx)
	
	// Calculate deadline
	deadline := tm.clock.Now().Add(duration)

	// Create timer
	timer := time.NewTimer(duration)

	entry := &timeoutEntry{
		auctionID: auctionID,
		timer:     timer,
		callback:  callback,
		deadline:  deadline,
		ctx:       timeoutCtx,
		cancel:    cancel,
	}

	tm.timeouts[auctionID] = entry

	// Start goroutine to handle timeout
	go tm.handleTimeout(entry)

	tm.logger.Debug("timeout scheduled",
		zap.String("auction_id", auctionID.String()),
		zap.Duration("duration", duration),
		zap.Time("deadline", deadline))

	return nil
}

// handleTimeout handles the timeout for an auction
func (tm *TimeoutManager) handleTimeout(entry *timeoutEntry) {
	select {
	case <-entry.timer.C:
		// Timeout occurred
		tm.logger.Info("auction timeout triggered",
			zap.String("auction_id", entry.auctionID.String()),
			zap.Time("deadline", entry.deadline))

		// Execute callback
		if entry.callback != nil {
			entry.callback(entry.ctx, entry.auctionID)
		}

		// Clean up
		tm.mu.Lock()
		delete(tm.timeouts, entry.auctionID)
		tm.mu.Unlock()

	case <-entry.ctx.Done():
		// Timeout was cancelled
		tm.logger.Debug("auction timeout cancelled",
			zap.String("auction_id", entry.auctionID.String()))

	case <-tm.stopCh:
		// Timeout manager is stopping
		tm.logger.Debug("timeout manager stopping, cancelling timeout",
			zap.String("auction_id", entry.auctionID.String()))
		entry.timer.Stop()
	}

	// Cleanup
	entry.cancel()
}

// CancelTimeout cancels a previously scheduled timeout
func (tm *TimeoutManager) CancelTimeout(auctionID uuid.UUID) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry, exists := tm.timeouts[auctionID]
	if !exists {
		tm.logger.Debug("no timeout found to cancel",
			zap.String("auction_id", auctionID.String()))
		return nil
	}

	// Cancel the timeout
	entry.cancel()
	entry.timer.Stop()
	delete(tm.timeouts, auctionID)

	tm.logger.Debug("timeout cancelled",
		zap.String("auction_id", auctionID.String()))

	return nil
}

// UpdateTimeout updates the duration of an existing timeout
func (tm *TimeoutManager) UpdateTimeout(auctionID uuid.UUID, newDuration time.Duration) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entry, exists := tm.timeouts[auctionID]
	if !exists {
		tm.logger.Debug("no timeout found to update",
			zap.String("auction_id", auctionID.String()))
		return nil
	}

	// Cancel existing timer
	entry.cancel()
	entry.timer.Stop()

	// Create new timeout context
	timeoutCtx, cancel := context.WithCancel(context.Background())
	
	// Calculate new deadline
	newDeadline := tm.clock.Now().Add(newDuration)

	// Create new timer
	newTimer := time.NewTimer(newDuration)

	// Update entry
	entry.timer = newTimer
	entry.deadline = newDeadline
	entry.ctx = timeoutCtx
	entry.cancel = cancel

	// Start new timeout handler
	go tm.handleTimeout(entry)

	tm.logger.Debug("timeout updated",
		zap.String("auction_id", auctionID.String()),
		zap.Duration("new_duration", newDuration),
		zap.Time("new_deadline", newDeadline))

	return nil
}

// GetRemainingTime returns the remaining time for a timeout
func (tm *TimeoutManager) GetRemainingTime(auctionID uuid.UUID) (time.Duration, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	entry, exists := tm.timeouts[auctionID]
	if !exists {
		return 0, false
	}

	remaining := tm.clock.Until(entry.deadline)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, true
}

// IsActive checks if a timeout is currently active
func (tm *TimeoutManager) IsActive(auctionID uuid.UUID) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, exists := tm.timeouts[auctionID]
	return exists
}

// Stop stops the timeout manager and cancels all pending timeouts
func (tm *TimeoutManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.stopped {
		return
	}

	tm.stopped = true
	close(tm.stopCh)

	// Cancel all existing timeouts
	for auctionID, entry := range tm.timeouts {
		entry.cancel()
		entry.timer.Stop()
		tm.logger.Debug("timeout cancelled during shutdown",
			zap.String("auction_id", auctionID.String()))
	}

	// Clear timeouts map
	tm.timeouts = make(map[uuid.UUID]*timeoutEntry)

	tm.logger.Info("timeout manager stopped")
}