package services

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"

	"auction-microservice/internal/domain"
)

// MockUserRepository is an in-memory implementation of UserRepository for testing
type MockUserRepository struct {
	users map[uuid.UUID]*domain.User
	mu    sync.RWMutex
}

// NewMockUserRepository creates a new MockUserRepository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return domain.ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[id]; !exists {
		return domain.ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}

func (r *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	users := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	
	// Sort by creation time
	sort.Slice(users, func(i, j int) bool {
		return users[i].CreatedAt.Before(users[j].CreatedAt)
	})
	
	start := offset
	if start > len(users) {
		return []*domain.User{}, nil
	}
	
	end := offset + limit
	if end > len(users) {
		end = len(users)
	}
	
	return users[start:end], nil
}

// MockListingRepository is an in-memory implementation of ListingRepository for testing
type MockListingRepository struct {
	listings map[uuid.UUID]*domain.Listing
	mu       sync.RWMutex
}

// NewMockListingRepository creates a new MockListingRepository
func NewMockListingRepository() *MockListingRepository {
	return &MockListingRepository{
		listings: make(map[uuid.UUID]*domain.Listing),
	}
}

func (r *MockListingRepository) Create(ctx context.Context, listing *domain.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listings[listing.ID] = listing
	return nil
}

func (r *MockListingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	listing, exists := r.listings[id]
	if !exists {
		return nil, domain.ErrListingNotFound
	}
	return listing, nil
}

func (r *MockListingRepository) Update(ctx context.Context, listing *domain.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.listings[listing.ID]; !exists {
		return domain.ErrListingNotFound
	}
	r.listings[listing.ID] = listing
	return nil
}

func (r *MockListingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.listings[id]; !exists {
		return domain.ErrListingNotFound
	}
	delete(r.listings, id)
	return nil
}

func (r *MockListingRepository) List(ctx context.Context, offset, limit int) ([]*domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	listings := make([]*domain.Listing, 0, len(r.listings))
	for _, listing := range r.listings {
		listings = append(listings, listing)
	}
	
	sort.Slice(listings, func(i, j int) bool {
		return listings[i].CreatedAt.Before(listings[j].CreatedAt)
	})
	
	start := offset
	if start > len(listings) {
		return []*domain.Listing{}, nil
	}
	
	end := offset + limit
	if end > len(listings) {
		end = len(listings)
	}
	
	return listings[start:end], nil
}

func (r *MockListingRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, offset, limit int) ([]*domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	listings := make([]*domain.Listing, 0)
	for _, listing := range r.listings {
		if listing.OwnerID == ownerID {
			listings = append(listings, listing)
		}
	}
	
	sort.Slice(listings, func(i, j int) bool {
		return listings[i].CreatedAt.Before(listings[j].CreatedAt)
	})
	
	start := offset
	if start > len(listings) {
		return []*domain.Listing{}, nil
	}
	
	end := offset + limit
	if end > len(listings) {
		end = len(listings)
	}
	
	return listings[start:end], nil
}

func (r *MockListingRepository) ListByStatus(ctx context.Context, status domain.ListingStatus, offset, limit int) ([]*domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	listings := make([]*domain.Listing, 0)
	for _, listing := range r.listings {
		if listing.Status == status {
			listings = append(listings, listing)
		}
	}
	
	sort.Slice(listings, func(i, j int) bool {
		return listings[i].CreatedAt.Before(listings[j].CreatedAt)
	})
	
	start := offset
	if start > len(listings) {
		return []*domain.Listing{}, nil
	}
	
	end := offset + limit
	if end > len(listings) {
		end = len(listings)
	}
	
	return listings[start:end], nil
}

// MockAuctionRepository is an in-memory implementation of AuctionRepository for testing
type MockAuctionRepository struct {
	auctions map[uuid.UUID]*domain.Auction
	mu       sync.RWMutex
}

// NewMockAuctionRepository creates a new MockAuctionRepository
func NewMockAuctionRepository() *MockAuctionRepository {
	return &MockAuctionRepository{
		auctions: make(map[uuid.UUID]*domain.Auction),
	}
}

func (r *MockAuctionRepository) Create(ctx context.Context, auction *domain.Auction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.auctions[auction.ID] = auction
	return nil
}

func (r *MockAuctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Auction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	auction, exists := r.auctions[id]
	if !exists {
		return nil, domain.ErrAuctionNotFound
	}
	return auction, nil
}

func (r *MockAuctionRepository) GetByListingID(ctx context.Context, listingID uuid.UUID) (*domain.Auction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, auction := range r.auctions {
		if auction.ListingID == listingID {
			return auction, nil
		}
	}
	return nil, domain.ErrAuctionNotFound
}

func (r *MockAuctionRepository) Update(ctx context.Context, auction *domain.Auction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.auctions[auction.ID]; !exists {
		return domain.ErrAuctionNotFound
	}
	r.auctions[auction.ID] = auction
	return nil
}

func (r *MockAuctionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.auctions[id]; !exists {
		return domain.ErrAuctionNotFound
	}
	delete(r.auctions, id)
	return nil
}

func (r *MockAuctionRepository) List(ctx context.Context, offset, limit int) ([]*domain.Auction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	auctions := make([]*domain.Auction, 0, len(r.auctions))
	for _, auction := range r.auctions {
		auctions = append(auctions, auction)
	}
	
	sort.Slice(auctions, func(i, j int) bool {
		return auctions[i].CreatedAt.Before(auctions[j].CreatedAt)
	})
	
	start := offset
	if start > len(auctions) {
		return []*domain.Auction{}, nil
	}
	
	end := offset + limit
	if end > len(auctions) {
		end = len(auctions)
	}
	
	return auctions[start:end], nil
}

func (r *MockAuctionRepository) ListByStatus(ctx context.Context, status domain.AuctionStatus, offset, limit int) ([]*domain.Auction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	auctions := make([]*domain.Auction, 0)
	for _, auction := range r.auctions {
		if auction.Status == status {
			auctions = append(auctions, auction)
		}
	}
	
	sort.Slice(auctions, func(i, j int) bool {
		return auctions[i].CreatedAt.Before(auctions[j].CreatedAt)
	})
	
	start := offset
	if start > len(auctions) {
		return []*domain.Auction{}, nil
	}
	
	end := offset + limit
	if end > len(auctions) {
		end = len(auctions)
	}
	
	return auctions[start:end], nil
}

func (r *MockAuctionRepository) ListActive(ctx context.Context, offset, limit int) ([]*domain.Auction, error) {
	return r.ListByStatus(ctx, domain.AuctionStatusActive, offset, limit)
}

func (r *MockAuctionRepository) ListExpired(ctx context.Context) ([]*domain.Auction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var expired []*domain.Auction
	for _, auction := range r.auctions {
		if auction.Status == domain.AuctionStatusActive && auction.IsExpired() {
			expired = append(expired, auction)
		}
	}
	
	return expired, nil
}

// MockBidRepository is an in-memory implementation of BidRepository for testing
type MockBidRepository struct {
	bids map[uuid.UUID]*domain.Bid
	mu   sync.RWMutex
}

// NewMockBidRepository creates a new MockBidRepository
func NewMockBidRepository() *MockBidRepository {
	return &MockBidRepository{
		bids: make(map[uuid.UUID]*domain.Bid),
	}
}

func (r *MockBidRepository) Create(ctx context.Context, bid *domain.Bid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bids[bid.ID] = bid
	return nil
}

func (r *MockBidRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bid, exists := r.bids[id]
	if !exists {
		return nil, domain.ErrBidNotFound
	}
	return bid, nil
}

func (r *MockBidRepository) Update(ctx context.Context, bid *domain.Bid) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.bids[bid.ID]; !exists {
		return domain.ErrBidNotFound
	}
	r.bids[bid.ID] = bid
	return nil
}

func (r *MockBidRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.bids[id]; !exists {
		return domain.ErrBidNotFound
	}
	delete(r.bids, id)
	return nil
}

func (r *MockBidRepository) ListByAuction(ctx context.Context, auctionID uuid.UUID, offset, limit int) ([]*domain.Bid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	bids := make([]*domain.Bid, 0)
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			bids = append(bids, bid)
		}
	}
	
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].PlacedAt.Before(bids[j].PlacedAt)
	})
	
	start := offset
	if start > len(bids) {
		return []*domain.Bid{}, nil
	}
	
	end := offset + limit
	if end > len(bids) {
		end = len(bids)
	}
	
	return bids[start:end], nil
}

func (r *MockBidRepository) ListByBidder(ctx context.Context, bidderID uuid.UUID, offset, limit int) ([]*domain.Bid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	bids := make([]*domain.Bid, 0)
	for _, bid := range r.bids {
		if bid.BidderID == bidderID {
			bids = append(bids, bid)
		}
	}
	
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].PlacedAt.After(bids[j].PlacedAt)
	})
	
	start := offset
	if start > len(bids) {
		return []*domain.Bid{}, nil
	}
	
	end := offset + limit
	if end > len(bids) {
		end = len(bids)
	}
	
	return bids[start:end], nil
}

func (r *MockBidRepository) GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*domain.Bid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var highestBid *domain.Bid
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			if highestBid == nil || bid.Amount > highestBid.Amount {
				highestBid = bid
			}
		}
	}
	
	if highestBid == nil {
		return nil, domain.ErrBidNotFound
	}
	
	return highestBid, nil
}

func (r *MockBidRepository) GetBidHistory(ctx context.Context, auctionID uuid.UUID) ([]*domain.Bid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	bids := make([]*domain.Bid, 0)
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			bids = append(bids, bid)
		}
	}
	
	sort.Slice(bids, func(i, j int) bool {
		return bids[i].PlacedAt.After(bids[j].PlacedAt)
	})
	
	return bids, nil
}

func (r *MockBidRepository) CountByAuction(ctx context.Context, auctionID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	count := 0
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			count++
		}
	}
	
	return count, nil
}

// MockSubscriptionRepository is an in-memory implementation of SubscriptionRepository for testing
type MockSubscriptionRepository struct {
	// subscriptions maps userID -> set of listingIDs
	subscriptions map[uuid.UUID]map[uuid.UUID]bool
	mu           sync.RWMutex
}

// NewMockSubscriptionRepository creates a new MockSubscriptionRepository
func NewMockSubscriptionRepository() *MockSubscriptionRepository {
	return &MockSubscriptionRepository{
		subscriptions: make(map[uuid.UUID]map[uuid.UUID]bool),
	}
}

func (r *MockSubscriptionRepository) Create(ctx context.Context, userID, listingID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.subscriptions[userID] == nil {
		r.subscriptions[userID] = make(map[uuid.UUID]bool)
	}
	
	r.subscriptions[userID][listingID] = true
	return nil
}

func (r *MockSubscriptionRepository) Delete(ctx context.Context, userID, listingID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.subscriptions[userID] != nil {
		delete(r.subscriptions[userID], listingID)
		if len(r.subscriptions[userID]) == 0 {
			delete(r.subscriptions, userID)
		}
	}
	
	return nil
}

func (r *MockSubscriptionRepository) GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var listingIDs []uuid.UUID
	if userSubscriptions := r.subscriptions[userID]; userSubscriptions != nil {
		for listingID := range userSubscriptions {
			listingIDs = append(listingIDs, listingID)
		}
	}
	
	return listingIDs, nil
}

func (r *MockSubscriptionRepository) GetSubscribers(ctx context.Context, listingID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var userIDs []uuid.UUID
	for userID, userSubscriptions := range r.subscriptions {
		if userSubscriptions[listingID] {
			userIDs = append(userIDs, userID)
		}
	}
	
	return userIDs, nil
}

func (r *MockSubscriptionRepository) IsSubscribed(ctx context.Context, userID, listingID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if userSubscriptions := r.subscriptions[userID]; userSubscriptions != nil {
		return userSubscriptions[listingID], nil
	}
	
	return false, nil
}