package domain

import (
	"time"

	"github.com/google/uuid"
)

// BidStatus represents the status of a bid
type BidStatus string

const (
	BidStatusActive    BidStatus = "active"
	BidStatusOutbid    BidStatus = "outbid"
	BidStatusWinning   BidStatus = "winning"
	BidStatusRejected  BidStatus = "rejected"
)

// Bid represents a bid placed on an auction
type Bid struct {
	ID        uuid.UUID `json:"id"`
	AuctionID uuid.UUID `json:"auction_id"`
	BidderID  uuid.UUID `json:"bidder_id"`
	Amount    int64     `json:"amount"` // in cents
	Status    BidStatus `json:"status"`
	PlacedAt  time.Time `json:"placed_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewBid creates a new bid
func NewBid(auctionID, bidderID uuid.UUID, amount int64) *Bid {
	now := time.Now()
	return &Bid{
		ID:        uuid.New(),
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    amount,
		Status:    BidStatusActive,
		PlacedAt:  now,
		UpdatedAt: now,
	}
}

// Validate validates the bid data
func (b *Bid) Validate() error {
	if b.AuctionID == uuid.Nil {
		return ErrInvalidAuctionID
	}
	if b.BidderID == uuid.Nil {
		return ErrInvalidBidder
	}
	if b.Amount <= 0 {
		return ErrInvalidBidAmount
	}
	return nil
}

// MarkAsOutbid marks the bid as outbid by a higher bid
func (b *Bid) MarkAsOutbid() {
	b.Status = BidStatusOutbid
	b.UpdatedAt = time.Now()
}

// MarkAsWinning marks the bid as the winning bid
func (b *Bid) MarkAsWinning() {
	b.Status = BidStatusWinning
	b.UpdatedAt = time.Now()
}

// MarkAsRejected marks the bid as rejected
func (b *Bid) MarkAsRejected() {
	b.Status = BidStatusRejected
	b.UpdatedAt = time.Now()
}

// IsActive returns true if the bid is active
func (b *Bid) IsActive() bool {
	return b.Status == BidStatusActive
}

// IsWinning returns true if the bid is winning
func (b *Bid) IsWinning() bool {
	return b.Status == BidStatusWinning
}