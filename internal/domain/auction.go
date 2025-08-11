package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuctionStatus represents the status of an auction
type AuctionStatus string

const (
	AuctionStatusPending   AuctionStatus = "pending"
	AuctionStatusActive    AuctionStatus = "active"
	AuctionStatusCompleted AuctionStatus = "completed"
	AuctionStatusCancelled AuctionStatus = "cancelled"
)

// Auction represents an auction for a listing
type Auction struct {
	ID              uuid.UUID     `json:"id"`
	ListingID       uuid.UUID     `json:"listing_id"`
	Status          AuctionStatus `json:"status"`
	StartTime       *time.Time    `json:"start_time,omitempty"`
	EndTime         *time.Time    `json:"end_time,omitempty"`
	Duration        time.Duration `json:"duration"`
	CurrentBid      int64         `json:"current_bid"` // in cents
	CurrentBidderID *uuid.UUID    `json:"current_bidder_id,omitempty"`
	BidCount        int           `json:"bid_count"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// NewAuction creates a new auction for a listing
func NewAuction(listingID uuid.UUID, duration time.Duration) *Auction {
	now := time.Now()
	return &Auction{
		ID:         uuid.New(),
		ListingID:  listingID,
		Status:     AuctionStatusPending,
		Duration:   duration,
		CurrentBid: 0,
		BidCount:   0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Validate validates the auction data
func (a *Auction) Validate() error {
	if a.ListingID == uuid.Nil {
		return ErrInvalidListingID
	}
	if a.Duration <= 0 {
		return ErrInvalidDuration
	}
	return nil
}

// Start starts the auction
func (a *Auction) Start() error {
	if a.Status != AuctionStatusPending {
		return ErrInvalidStatusTransition
	}
	
	now := time.Now()
	endTime := now.Add(a.Duration)
	
	a.Status = AuctionStatusActive
	a.StartTime = &now
	a.EndTime = &endTime
	a.UpdatedAt = now
	
	return nil
}

// Complete completes the auction
func (a *Auction) Complete() error {
	if a.Status != AuctionStatusActive {
		return ErrInvalidStatusTransition
	}
	
	a.Status = AuctionStatusCompleted
	a.UpdatedAt = time.Now()
	
	return nil
}

// Cancel cancels the auction
func (a *Auction) Cancel() error {
	if a.Status != AuctionStatusPending && a.Status != AuctionStatusActive {
		return ErrInvalidStatusTransition
	}
	
	a.Status = AuctionStatusCancelled
	a.UpdatedAt = time.Now()
	
	return nil
}

// PlaceBid places a new bid on the auction
func (a *Auction) PlaceBid(bidderID uuid.UUID, amount int64) error {
	if a.Status != AuctionStatusActive {
		return ErrAuctionNotActive
	}
	
	if amount <= a.CurrentBid {
		return ErrBidTooLow
	}
	
	a.CurrentBid = amount
	a.CurrentBidderID = &bidderID
	a.BidCount++
	a.UpdatedAt = time.Now()
	
	return nil
}

// IsActive returns true if the auction is active
func (a *Auction) IsActive() bool {
	return a.Status == AuctionStatusActive
}

// IsCompleted returns true if the auction is completed
func (a *Auction) IsCompleted() bool {
	return a.Status == AuctionStatusCompleted
}

// HasBids returns true if the auction has received bids
func (a *Auction) HasBids() bool {
	return a.BidCount > 0
}

// TimeRemaining returns the time remaining in the auction
func (a *Auction) TimeRemaining() time.Duration {
	if a.EndTime == nil || a.Status != AuctionStatusActive {
		return 0
	}
	
	remaining := time.Until(*a.EndTime)
	if remaining < 0 {
		return 0
	}
	
	return remaining
}

// IsExpired returns true if the auction has expired
func (a *Auction) IsExpired() bool {
	if a.EndTime == nil {
		return false
	}
	return time.Now().After(*a.EndTime)
}

// ExtendDuration extends the auction duration (for anti-sniping)
func (a *Auction) ExtendDuration(extension time.Duration) error {
	if a.Status != AuctionStatusActive || a.EndTime == nil {
		return ErrCannotExtendAuction
	}
	
	newEndTime := a.EndTime.Add(extension)
	a.EndTime = &newEndTime
	a.UpdatedAt = time.Now()
	
	return nil
}