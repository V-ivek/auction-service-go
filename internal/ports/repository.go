package ports

import (
	"context"

	"github.com/google/uuid"

	"auction-microservice/internal/domain"
)

// UserRepository defines the interface for user data persistence
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}

// ListingRepository defines the interface for listing data persistence
type ListingRepository interface {
	Create(ctx context.Context, listing *domain.Listing) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Listing, error)
	Update(ctx context.Context, listing *domain.Listing) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*domain.Listing, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, offset, limit int) ([]*domain.Listing, error)
	ListByStatus(ctx context.Context, status domain.ListingStatus, offset, limit int) ([]*domain.Listing, error)
}

// AuctionRepository defines the interface for auction data persistence
type AuctionRepository interface {
	Create(ctx context.Context, auction *domain.Auction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Auction, error)
	GetByListingID(ctx context.Context, listingID uuid.UUID) (*domain.Auction, error)
	Update(ctx context.Context, auction *domain.Auction) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*domain.Auction, error)
	ListByStatus(ctx context.Context, status domain.AuctionStatus, offset, limit int) ([]*domain.Auction, error)
	ListActive(ctx context.Context, offset, limit int) ([]*domain.Auction, error)
	ListExpired(ctx context.Context) ([]*domain.Auction, error)
}

// BidRepository defines the interface for bid data persistence
type BidRepository interface {
	Create(ctx context.Context, bid *domain.Bid) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bid, error)
	Update(ctx context.Context, bid *domain.Bid) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByAuction(ctx context.Context, auctionID uuid.UUID, offset, limit int) ([]*domain.Bid, error)
	ListByBidder(ctx context.Context, bidderID uuid.UUID, offset, limit int) ([]*domain.Bid, error)
	GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*domain.Bid, error)
	GetBidHistory(ctx context.Context, auctionID uuid.UUID) ([]*domain.Bid, error)
	CountByAuction(ctx context.Context, auctionID uuid.UUID) (int, error)
}

// SubscriptionRepository defines the interface for subscription data persistence
type SubscriptionRepository interface {
	Create(ctx context.Context, userID, listingID uuid.UUID) error
	Delete(ctx context.Context, userID, listingID uuid.UUID) error
	GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetSubscribers(ctx context.Context, listingID uuid.UUID) ([]uuid.UUID, error)
	IsSubscribed(ctx context.Context, userID, listingID uuid.UUID) (bool, error)
}