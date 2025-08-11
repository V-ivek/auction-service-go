package domain

import (
	"time"

	"github.com/google/uuid"
)

// ListingStatus represents the status of a listing
type ListingStatus string

const (
	ListingStatusDraft     ListingStatus = "draft"
	ListingStatusActive    ListingStatus = "active"
	ListingStatusCancelled ListingStatus = "cancelled"
	ListingStatusSold      ListingStatus = "sold"
)

// Listing represents an item that can be auctioned
type Listing struct {
	ID           uuid.UUID     `json:"id"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	StartingBid  int64         `json:"starting_bid"` // in cents
	ReservePrice int64         `json:"reserve_price"` // in cents
	OwnerID      uuid.UUID     `json:"owner_id"`
	Status       ListingStatus `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// NewListing creates a new listing
func NewListing(title, description string, startingBid, reservePrice int64, ownerID uuid.UUID) *Listing {
	now := time.Now()
	return &Listing{
		ID:           uuid.New(),
		Title:        title,
		Description:  description,
		StartingBid:  startingBid,
		ReservePrice: reservePrice,
		OwnerID:      ownerID,
		Status:       ListingStatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Validate validates the listing data
func (l *Listing) Validate() error {
	if l.Title == "" {
		return ErrInvalidTitle
	}
	if l.StartingBid <= 0 {
		return ErrInvalidStartingBid
	}
	if l.ReservePrice < l.StartingBid {
		return ErrReserveBelowStarting
	}
	if l.OwnerID == uuid.Nil {
		return ErrInvalidOwner
	}
	return nil
}

// Activate activates the listing
func (l *Listing) Activate() error {
	if l.Status != ListingStatusDraft {
		return ErrInvalidStatusTransition
	}
	l.Status = ListingStatusActive
	l.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the listing
func (l *Listing) Cancel() error {
	if l.Status != ListingStatusDraft && l.Status != ListingStatusActive {
		return ErrInvalidStatusTransition
	}
	l.Status = ListingStatusCancelled
	l.UpdatedAt = time.Now()
	return nil
}

// MarkAsSold marks the listing as sold
func (l *Listing) MarkAsSold() error {
	if l.Status != ListingStatusActive {
		return ErrInvalidStatusTransition
	}
	l.Status = ListingStatusSold
	l.UpdatedAt = time.Now()
	return nil
}

// IsActive returns true if the listing is active
func (l *Listing) IsActive() bool {
	return l.Status == ListingStatusActive
}

// Update updates the listing details
func (l *Listing) Update(title, description string, startingBid, reservePrice int64) error {
	if l.Status != ListingStatusDraft {
		return ErrCannotUpdateActiveListing
	}
	
	l.Title = title
	l.Description = description
	l.StartingBid = startingBid
	l.ReservePrice = reservePrice
	l.UpdatedAt = time.Now()
	
	return l.Validate()
}