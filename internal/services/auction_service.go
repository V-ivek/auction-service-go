package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/internal/domain"
	"auction-microservice/internal/ports"
)

// AuctionServiceImpl implements the AuctionService interface
type AuctionServiceImpl struct {
	auctionRepo    ports.AuctionRepository
	listingRepo    ports.ListingRepository
	bidRepo        ports.BidRepository
	notifier       ports.Notifier
	timeoutManager ports.TimeoutManager
	clock          ports.Clock
	logger         *zap.Logger
}

// NewAuctionService creates a new instance of AuctionService
func NewAuctionService(
	auctionRepo ports.AuctionRepository,
	listingRepo ports.ListingRepository,
	bidRepo ports.BidRepository,
	notifier ports.Notifier,
	timeoutManager ports.TimeoutManager,
	clock ports.Clock,
	logger *zap.Logger,
) *AuctionServiceImpl {
	return &AuctionServiceImpl{
		auctionRepo:    auctionRepo,
		listingRepo:    listingRepo,
		bidRepo:        bidRepo,
		notifier:       notifier,
		timeoutManager: timeoutManager,
		clock:          clock,
		logger:         logger,
	}
}

// CreateAuction creates a new auction for a listing
func (s *AuctionServiceImpl) CreateAuction(ctx context.Context, listingID uuid.UUID, duration time.Duration) (*domain.Auction, error) {
	// Validate listing exists and is active
	listing, err := s.listingRepo.GetByID(ctx, listingID)
	if err != nil {
		s.logger.Error("listing not found", zap.Error(err), zap.String("listing_id", listingID.String()))
		return nil, domain.ErrListingNotFound
	}

	if !listing.IsActive() {
		s.logger.Error("listing is not active", zap.String("listing_id", listingID.String()))
		return nil, domain.ErrInvalidStatusTransition
	}

	// Check if auction already exists for this listing
	existingAuction, err := s.auctionRepo.GetByListingID(ctx, listingID)
	if err == nil && existingAuction != nil {
		s.logger.Error("auction already exists for listing", zap.String("listing_id", listingID.String()))
		return nil, domain.ErrAuctionAlreadyStarted
	}

	// Create auction
	auction := domain.NewAuction(listingID, duration)
	if err := auction.Validate(); err != nil {
		s.logger.Error("invalid auction data", zap.Error(err))
		return nil, err
	}

	// Save to repository
	if err := s.auctionRepo.Create(ctx, auction); err != nil {
		s.logger.Error("failed to create auction", zap.Error(err))
		return nil, err
	}

	s.logger.Info("auction created", 
		zap.String("auction_id", auction.ID.String()),
		zap.String("listing_id", listingID.String()),
		zap.Duration("duration", duration))

	return auction, nil
}

// StartAuction starts a pending auction
func (s *AuctionServiceImpl) StartAuction(ctx context.Context, auctionID uuid.UUID) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		s.logger.Error("auction not found", zap.Error(err), zap.String("auction_id", auctionID.String()))
		return domain.ErrAuctionNotFound
	}

	if err := auction.Start(); err != nil {
		s.logger.Error("failed to start auction", zap.Error(err))
		return err
	}

	// Save updated auction
	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		s.logger.Error("failed to save started auction", zap.Error(err))
		return err
	}

	// Schedule timeout for auction end
	if s.timeoutManager != nil {
		callback := func(ctx context.Context, auctionID uuid.UUID) {
			if err := s.CompleteAuction(ctx, auctionID); err != nil {
				s.logger.Error("failed to complete auction on timeout", zap.Error(err))
			}
		}
		
		if err := s.timeoutManager.ScheduleTimeout(ctx, auctionID, auction.Duration, callback); err != nil {
			s.logger.Error("failed to schedule auction timeout", zap.Error(err))
		}
	}

	// Notify about auction start
	if s.notifier != nil {
		event := domain.AuctionStartedEvent(auction)
		if err := s.notifier.NotifyAuctionStarted(event); err != nil {
			s.logger.Error("failed to notify auction start", zap.Error(err))
		}
	}

	s.logger.Info("auction started", 
		zap.String("auction_id", auctionID.String()),
		zap.Time("end_time", *auction.EndTime))

	return nil
}

// GetAuction retrieves an auction by ID
func (s *AuctionServiceImpl) GetAuction(ctx context.Context, id uuid.UUID) (*domain.Auction, error) {
	auction, err := s.auctionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get auction", zap.Error(err), zap.String("auction_id", id.String()))
		return nil, err
	}
	return auction, nil
}

// GetAuctionByListing retrieves an auction by listing ID
func (s *AuctionServiceImpl) GetAuctionByListing(ctx context.Context, listingID uuid.UUID) (*domain.Auction, error) {
	auction, err := s.auctionRepo.GetByListingID(ctx, listingID)
	if err != nil {
		s.logger.Error("failed to get auction by listing", zap.Error(err), zap.String("listing_id", listingID.String()))
		return nil, err
	}
	return auction, nil
}

// GetActiveAuctions retrieves all active auctions with pagination
func (s *AuctionServiceImpl) GetActiveAuctions(ctx context.Context, offset, limit int) ([]*domain.Auction, error) {
	auctions, err := s.auctionRepo.ListActive(ctx, offset, limit)
	if err != nil {
		s.logger.Error("failed to get active auctions", zap.Error(err))
		return nil, err
	}
	return auctions, nil
}

// CompleteAuction completes an auction
func (s *AuctionServiceImpl) CompleteAuction(ctx context.Context, auctionID uuid.UUID) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		s.logger.Error("auction not found", zap.Error(err), zap.String("auction_id", auctionID.String()))
		return domain.ErrAuctionNotFound
	}

	if err := auction.Complete(); err != nil {
		s.logger.Error("failed to complete auction", zap.Error(err))
		return err
	}

	// Save updated auction
	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		s.logger.Error("failed to save completed auction", zap.Error(err))
		return err
	}

	// Cancel timeout if it exists
	if s.timeoutManager != nil {
		s.timeoutManager.CancelTimeout(auctionID)
	}

	// Update listing status if there was a winning bid
	if auction.HasBids() {
		listing, err := s.listingRepo.GetByID(ctx, auction.ListingID)
		if err == nil {
			listing.MarkAsSold()
			s.listingRepo.Update(ctx, listing)
		}
	}

	// Notify about auction end
	if s.notifier != nil {
		event := domain.AuctionEndedEvent(auction)
		if err := s.notifier.NotifyAuctionEnded(event); err != nil {
			s.logger.Error("failed to notify auction end", zap.Error(err))
		}
	}

	s.logger.Info("auction completed", 
		zap.String("auction_id", auctionID.String()),
		zap.Int64("final_bid", auction.CurrentBid),
		zap.Int("bid_count", auction.BidCount))

	return nil
}

// CancelAuction cancels an auction
func (s *AuctionServiceImpl) CancelAuction(ctx context.Context, auctionID uuid.UUID) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		s.logger.Error("auction not found", zap.Error(err), zap.String("auction_id", auctionID.String()))
		return domain.ErrAuctionNotFound
	}

	if err := auction.Cancel(); err != nil {
		s.logger.Error("failed to cancel auction", zap.Error(err))
		return err
	}

	// Save updated auction
	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		s.logger.Error("failed to save cancelled auction", zap.Error(err))
		return err
	}

	// Cancel timeout if it exists
	if s.timeoutManager != nil {
		s.timeoutManager.CancelTimeout(auctionID)
	}

	// Notify about auction cancellation
	if s.notifier != nil {
		event := domain.AuctionCancelledEvent(auction)
		if err := s.notifier.NotifyAuctionEnded(event); err != nil {
			s.logger.Error("failed to notify auction cancellation", zap.Error(err))
		}
	}

	s.logger.Info("auction cancelled", zap.String("auction_id", auctionID.String()))
	return nil
}

// ExtendAuction extends an auction duration (for anti-sniping)
func (s *AuctionServiceImpl) ExtendAuction(ctx context.Context, auctionID uuid.UUID, extension time.Duration) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		s.logger.Error("auction not found", zap.Error(err), zap.String("auction_id", auctionID.String()))
		return domain.ErrAuctionNotFound
	}

	if err := auction.ExtendDuration(extension); err != nil {
		s.logger.Error("failed to extend auction", zap.Error(err))
		return err
	}

	// Save updated auction
	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		s.logger.Error("failed to save extended auction", zap.Error(err))
		return err
	}

	// Update timeout
	if s.timeoutManager != nil {
		remaining := auction.TimeRemaining()
		if err := s.timeoutManager.UpdateTimeout(auctionID, remaining); err != nil {
			s.logger.Error("failed to update auction timeout", zap.Error(err))
		}
	}

	s.logger.Info("auction extended", 
		zap.String("auction_id", auctionID.String()),
		zap.Duration("extension", extension),
		zap.Time("new_end_time", *auction.EndTime))

	return nil
}

// ProcessExpiredAuctions processes all expired auctions
func (s *AuctionServiceImpl) ProcessExpiredAuctions(ctx context.Context) error {
	expiredAuctions, err := s.auctionRepo.ListExpired(ctx)
	if err != nil {
		s.logger.Error("failed to get expired auctions", zap.Error(err))
		return err
	}

	for _, auction := range expiredAuctions {
		if err := s.CompleteAuction(ctx, auction.ID); err != nil {
			s.logger.Error("failed to complete expired auction", 
				zap.Error(err), 
				zap.String("auction_id", auction.ID.String()))
		}
	}

	if len(expiredAuctions) > 0 {
		s.logger.Info("processed expired auctions", zap.Int("count", len(expiredAuctions)))
	}

	return nil
}