package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/internal/domain"
	"auction-microservice/internal/ports"
)

const (
	// AntiSnipingExtension is the extension time added when a bid is placed in the last minutes
	AntiSnipingExtension = 5 * time.Minute
	// AntiSnipingThreshold is the time threshold before auction end that triggers extension
	AntiSnipingThreshold = 5 * time.Minute
)

// BiddingServiceImpl implements the BiddingService interface
type BiddingServiceImpl struct {
	bidRepo        ports.BidRepository
	auctionRepo    ports.AuctionRepository
	listingRepo    ports.ListingRepository
	notifier       ports.Notifier
	timeoutManager ports.TimeoutManager
	clock          ports.Clock
	logger         *zap.Logger
}

// NewBiddingService creates a new instance of BiddingService
func NewBiddingService(
	bidRepo ports.BidRepository,
	auctionRepo ports.AuctionRepository,
	listingRepo ports.ListingRepository,
	notifier ports.Notifier,
	timeoutManager ports.TimeoutManager,
	clock ports.Clock,
	logger *zap.Logger,
) *BiddingServiceImpl {
	return &BiddingServiceImpl{
		bidRepo:        bidRepo,
		auctionRepo:    auctionRepo,
		listingRepo:    listingRepo,
		notifier:       notifier,
		timeoutManager: timeoutManager,
		clock:          clock,
		logger:         logger,
	}
}

// PlaceBid places a new bid on an auction
func (s *BiddingServiceImpl) PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount int64) (*domain.Bid, error) {
	// Validate bid
	if err := s.ValidateBid(ctx, auctionID, bidderID, amount); err != nil {
		return nil, err
	}

	// Get auction
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		s.logger.Error("auction not found", zap.Error(err), zap.String("auction_id", auctionID.String()))
		return nil, domain.ErrAuctionNotFound
	}

	// Get the current highest bid to mark it as outbid
	var previousHighBid *domain.Bid
	if auction.HasBids() {
		previousHighBid, _ = s.bidRepo.GetHighestBid(ctx, auctionID)
	}

	// Create new bid
	bid := domain.NewBid(auctionID, bidderID, amount)
	if err := bid.Validate(); err != nil {
		s.logger.Error("invalid bid data", zap.Error(err))
		return nil, err
	}

	// Place bid on auction (this updates current bid and count)
	if err := auction.PlaceBid(bidderID, amount); err != nil {
		s.logger.Error("failed to place bid on auction", zap.Error(err))
		return nil, err
	}

	// Save bid
	if err := s.bidRepo.Create(ctx, bid); err != nil {
		s.logger.Error("failed to save bid", zap.Error(err))
		return nil, err
	}

	// Update auction
	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		s.logger.Error("failed to update auction", zap.Error(err))
		return nil, err
	}

	// Mark previous high bid as outbid
	if previousHighBid != nil {
		previousHighBid.MarkAsOutbid()
		if err := s.bidRepo.Update(ctx, previousHighBid); err != nil {
			s.logger.Error("failed to mark previous bid as outbid", zap.Error(err))
		}

		// Notify previous bidder they were outbid
		if s.notifier != nil {
			event := domain.BidOutbidEvent(previousHighBid, bid)
			if err := s.notifier.NotifyBidOutbid(event); err != nil {
				s.logger.Error("failed to notify bid outbid", zap.Error(err))
			}
		}
	}

	// Check for anti-sniping (if bid is placed in the last few minutes)
	timeRemaining := auction.TimeRemaining()
	if timeRemaining > 0 && timeRemaining <= AntiSnipingThreshold {
		if err := auction.ExtendDuration(AntiSnipingExtension); err != nil {
			s.logger.Error("failed to extend auction for anti-sniping", zap.Error(err))
		} else {
			// Update auction with extended time
			if err := s.auctionRepo.Update(ctx, auction); err != nil {
				s.logger.Error("failed to save extended auction", zap.Error(err))
			}

			// Update timeout
			if s.timeoutManager != nil {
				newRemaining := auction.TimeRemaining()
				if err := s.timeoutManager.UpdateTimeout(auctionID, newRemaining); err != nil {
					s.logger.Error("failed to update timeout for extended auction", zap.Error(err))
				}
			}

			s.logger.Info("auction extended due to anti-sniping", 
				zap.String("auction_id", auctionID.String()),
				zap.Duration("extension", AntiSnipingExtension),
				zap.Time("new_end_time", *auction.EndTime))
		}
	}

	// Notify about the new bid
	if s.notifier != nil {
		event := domain.BidPlacedEvent(bid, auction)
		if err := s.notifier.NotifyBidPlaced(event); err != nil {
			s.logger.Error("failed to notify bid placed", zap.Error(err))
		}
	}

	s.logger.Info("bid placed", 
		zap.String("bid_id", bid.ID.String()),
		zap.String("auction_id", auctionID.String()),
		zap.String("bidder_id", bidderID.String()),
		zap.Int64("amount", amount),
		zap.Int("bid_count", auction.BidCount))

	return bid, nil
}

// GetBid retrieves a bid by ID
func (s *BiddingServiceImpl) GetBid(ctx context.Context, id uuid.UUID) (*domain.Bid, error) {
	bid, err := s.bidRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get bid", zap.Error(err), zap.String("bid_id", id.String()))
		return nil, err
	}
	return bid, nil
}

// GetBidHistory retrieves all bids for an auction
func (s *BiddingServiceImpl) GetBidHistory(ctx context.Context, auctionID uuid.UUID) ([]*domain.Bid, error) {
	bids, err := s.bidRepo.GetBidHistory(ctx, auctionID)
	if err != nil {
		s.logger.Error("failed to get bid history", zap.Error(err), zap.String("auction_id", auctionID.String()))
		return nil, err
	}
	return bids, nil
}

// GetBidsByBidder retrieves bids by a specific bidder
func (s *BiddingServiceImpl) GetBidsByBidder(ctx context.Context, bidderID uuid.UUID, offset, limit int) ([]*domain.Bid, error) {
	bids, err := s.bidRepo.ListByBidder(ctx, bidderID, offset, limit)
	if err != nil {
		s.logger.Error("failed to get bids by bidder", zap.Error(err), zap.String("bidder_id", bidderID.String()))
		return nil, err
	}
	return bids, nil
}

// GetHighestBid retrieves the highest bid for an auction
func (s *BiddingServiceImpl) GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*domain.Bid, error) {
	bid, err := s.bidRepo.GetHighestBid(ctx, auctionID)
	if err != nil {
		s.logger.Error("failed to get highest bid", zap.Error(err), zap.String("auction_id", auctionID.String()))
		return nil, err
	}
	return bid, nil
}

// ValidateBid validates a bid before it's placed
func (s *BiddingServiceImpl) ValidateBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount int64) error {
	// Get auction
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return domain.ErrAuctionNotFound
	}

	// Check if auction is active
	if !auction.IsActive() {
		return domain.ErrAuctionNotActive
	}

	// Check if auction has expired
	if auction.IsExpired() {
		return domain.ErrAuctionNotActive
	}

	// Get listing to check ownership
	listing, err := s.listingRepo.GetByID(ctx, auction.ListingID)
	if err != nil {
		return domain.ErrListingNotFound
	}

	// Note: Removed owner validation - listing owners can bid on their own items
	// This allows for legitimate scenarios like testing or self-bidding for reserve price

	// Check if bid amount is valid
	if amount <= 0 {
		return domain.ErrInvalidBidAmount
	}

	// Check if bid is higher than current bid
	if amount <= auction.CurrentBid {
		return domain.ErrBidTooLow
	}

	// Check if bid meets minimum starting bid
	if auction.CurrentBid == 0 && amount < listing.StartingBid {
		return domain.ErrBidTooLow
	}

	return nil
}