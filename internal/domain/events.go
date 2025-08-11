package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of domain event
type EventType string

const (
	EventTypeAuctionStarted   EventType = "auction_started"
	EventTypeAuctionEnded     EventType = "auction_ended"
	EventTypeAuctionCancelled EventType = "auction_cancelled"
	EventTypeBidPlaced        EventType = "bid_placed"
	EventTypeBidOutbid        EventType = "bid_outbid"
	EventTypeListingCreated   EventType = "listing_created"
	EventTypeListingUpdated   EventType = "listing_updated"
	EventTypeUserJoined       EventType = "user_joined"
	EventTypeUserLeft         EventType = "user_left"
)

// DomainEvent represents a domain event
type DomainEvent struct {
	ID          uuid.UUID              `json:"id"`
	Type        EventType              `json:"type"`
	AggregateID uuid.UUID              `json:"aggregate_id"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType EventType, aggregateID uuid.UUID, data map[string]interface{}) *DomainEvent {
	return &DomainEvent{
		ID:          uuid.New(),
		Type:        eventType,
		AggregateID: aggregateID,
		Data:        data,
		Timestamp:   time.Now(),
	}
}

// AuctionStartedEvent creates an auction started event
func AuctionStartedEvent(auction *Auction) *DomainEvent {
	data := map[string]interface{}{
		"auction_id": auction.ID,
		"listing_id": auction.ListingID,
		"start_time": auction.StartTime,
		"end_time":   auction.EndTime,
		"duration":   auction.Duration.String(),
	}
	return NewDomainEvent(EventTypeAuctionStarted, auction.ID, data)
}

// AuctionEndedEvent creates an auction ended event
func AuctionEndedEvent(auction *Auction) *DomainEvent {
	data := map[string]interface{}{
		"auction_id":         auction.ID,
		"listing_id":         auction.ListingID,
		"final_bid":          auction.CurrentBid,
		"winner_id":          auction.CurrentBidderID,
		"bid_count":          auction.BidCount,
		"completion_time":    time.Now(),
	}
	return NewDomainEvent(EventTypeAuctionEnded, auction.ID, data)
}

// AuctionCancelledEvent creates an auction cancelled event
func AuctionCancelledEvent(auction *Auction) *DomainEvent {
	data := map[string]interface{}{
		"auction_id":      auction.ID,
		"listing_id":      auction.ListingID,
		"current_bid":     auction.CurrentBid,
		"bid_count":       auction.BidCount,
		"cancellation_time": time.Now(),
	}
	return NewDomainEvent(EventTypeAuctionCancelled, auction.ID, data)
}

// BidPlacedEvent creates a bid placed event
func BidPlacedEvent(bid *Bid, auction *Auction) *DomainEvent {
	data := map[string]interface{}{
		"bid_id":     bid.ID,
		"auction_id": bid.AuctionID,
		"bidder_id":  bid.BidderID,
		"amount":     bid.Amount,
		"bid_time":   bid.PlacedAt,
		"new_high":   auction.CurrentBid == bid.Amount,
		"bid_count":  auction.BidCount,
	}
	return NewDomainEvent(EventTypeBidPlaced, bid.AuctionID, data)
}

// BidOutbidEvent creates a bid outbid event
func BidOutbidEvent(bid *Bid, newHighBid *Bid) *DomainEvent {
	data := map[string]interface{}{
		"outbid_bid_id":    bid.ID,
		"outbid_bidder_id": bid.BidderID,
		"outbid_amount":    bid.Amount,
		"new_high_bid_id":  newHighBid.ID,
		"new_high_amount":  newHighBid.Amount,
		"auction_id":       bid.AuctionID,
	}
	return NewDomainEvent(EventTypeBidOutbid, bid.AuctionID, data)
}

// ListingCreatedEvent creates a listing created event
func ListingCreatedEvent(listing *Listing) *DomainEvent {
	data := map[string]interface{}{
		"listing_id":     listing.ID,
		"title":          listing.Title,
		"starting_bid":   listing.StartingBid,
		"reserve_price":  listing.ReservePrice,
		"owner_id":       listing.OwnerID,
		"creation_time":  listing.CreatedAt,
	}
	return NewDomainEvent(EventTypeListingCreated, listing.ID, data)
}

// ListingUpdatedEvent creates a listing updated event
func ListingUpdatedEvent(listing *Listing) *DomainEvent {
	data := map[string]interface{}{
		"listing_id":     listing.ID,
		"title":          listing.Title,
		"status":         listing.Status,
		"owner_id":       listing.OwnerID,
		"update_time":    listing.UpdatedAt,
	}
	return NewDomainEvent(EventTypeListingUpdated, listing.ID, data)
}

// UserJoinedEvent creates a user joined event (for WebSocket connections)
func UserJoinedEvent(userID, auctionID uuid.UUID) *DomainEvent {
	data := map[string]interface{}{
		"user_id":    userID,
		"auction_id": auctionID,
		"join_time":  time.Now(),
	}
	return NewDomainEvent(EventTypeUserJoined, auctionID, data)
}

// UserLeftEvent creates a user left event (for WebSocket connections)
func UserLeftEvent(userID, auctionID uuid.UUID) *DomainEvent {
	data := map[string]interface{}{
		"user_id":    userID,
		"auction_id": auctionID,
		"leave_time": time.Now(),
	}
	return NewDomainEvent(EventTypeUserLeft, auctionID, data)
}