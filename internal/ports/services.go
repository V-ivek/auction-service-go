package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"auction-microservice/internal/domain"
)

// ListingService defines the interface for listing business logic
type ListingService interface {
	CreateListing(ctx context.Context, title, description string, startingBid, reservePrice int64, ownerID uuid.UUID) (*domain.Listing, error)
	GetListing(ctx context.Context, id uuid.UUID) (*domain.Listing, error)
	GetListings(ctx context.Context, offset, limit int) ([]*domain.Listing, error)
	GetListingsByOwner(ctx context.Context, ownerID uuid.UUID, offset, limit int) ([]*domain.Listing, error)
	UpdateListing(ctx context.Context, id uuid.UUID, title, description string, startingBid, reservePrice int64) error
	DeleteListing(ctx context.Context, id uuid.UUID) error
	ActivateListing(ctx context.Context, id uuid.UUID) error
	CancelListing(ctx context.Context, id uuid.UUID) error
}

// AuctionService defines the interface for auction business logic
type AuctionService interface {
	CreateAuction(ctx context.Context, listingID uuid.UUID, duration time.Duration) (*domain.Auction, error)
	StartAuction(ctx context.Context, auctionID uuid.UUID) error
	GetAuction(ctx context.Context, id uuid.UUID) (*domain.Auction, error)
	GetAuctionByListing(ctx context.Context, listingID uuid.UUID) (*domain.Auction, error)
	GetActiveAuctions(ctx context.Context, offset, limit int) ([]*domain.Auction, error)
	CompleteAuction(ctx context.Context, auctionID uuid.UUID) error
	CancelAuction(ctx context.Context, auctionID uuid.UUID) error
	ExtendAuction(ctx context.Context, auctionID uuid.UUID, extension time.Duration) error
	ProcessExpiredAuctions(ctx context.Context) error
}

// BiddingService defines the interface for bidding business logic
type BiddingService interface {
	PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount int64) (*domain.Bid, error)
	GetBid(ctx context.Context, id uuid.UUID) (*domain.Bid, error)
	GetBidHistory(ctx context.Context, auctionID uuid.UUID) ([]*domain.Bid, error)
	GetBidsByBidder(ctx context.Context, bidderID uuid.UUID, offset, limit int) ([]*domain.Bid, error)
	GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*domain.Bid, error)
	ValidateBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount int64) error
}

// SubscriptionService defines the interface for subscription business logic
type SubscriptionService interface {
	Subscribe(ctx context.Context, userID, listingID uuid.UUID) error
	Unsubscribe(ctx context.Context, userID, listingID uuid.UUID) error
	GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetSubscribers(ctx context.Context, listingID uuid.UUID) ([]uuid.UUID, error)
	IsSubscribed(ctx context.Context, userID, listingID uuid.UUID) (bool, error)
}