package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/internal/domain"
	"auction-microservice/internal/ports"
)

// ListingServiceImpl implements the ListingService interface
type ListingServiceImpl struct {
	listingRepo ports.ListingRepository
	userRepo    ports.UserRepository
	logger      *zap.Logger
}

// NewListingService creates a new instance of ListingService
func NewListingService(listingRepo ports.ListingRepository, userRepo ports.UserRepository, logger *zap.Logger) *ListingServiceImpl {
	return &ListingServiceImpl{
		listingRepo: listingRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// CreateListing creates a new listing
func (s *ListingServiceImpl) CreateListing(ctx context.Context, title, description string, startingBid, reservePrice int64, ownerID uuid.UUID) (*domain.Listing, error) {
	// Validate owner exists
	_, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		s.logger.Error("owner not found", zap.Error(err), zap.String("owner_id", ownerID.String()))
		return nil, domain.ErrUserNotFound
	}

	// Create listing
	listing := domain.NewListing(title, description, startingBid, reservePrice, ownerID)
	if err := listing.Validate(); err != nil {
		s.logger.Error("invalid listing data", zap.Error(err))
		return nil, err
	}

	// Save to repository
	if err := s.listingRepo.Create(ctx, listing); err != nil {
		s.logger.Error("failed to create listing", zap.Error(err))
		return nil, err
	}

	s.logger.Info("listing created", 
		zap.String("listing_id", listing.ID.String()),
		zap.String("title", listing.Title),
		zap.String("owner_id", ownerID.String()))

	return listing, nil
}

// GetListing retrieves a listing by ID
func (s *ListingServiceImpl) GetListing(ctx context.Context, id uuid.UUID) (*domain.Listing, error) {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get listing", zap.Error(err), zap.String("listing_id", id.String()))
		return nil, err
	}
	return listing, nil
}

// GetListings retrieves all listings with pagination
func (s *ListingServiceImpl) GetListings(ctx context.Context, offset, limit int) ([]*domain.Listing, error) {
	listings, err := s.listingRepo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("failed to get listings", zap.Error(err))
		return nil, err
	}
	return listings, nil
}

// GetListingsByOwner retrieves listings by owner with pagination
func (s *ListingServiceImpl) GetListingsByOwner(ctx context.Context, ownerID uuid.UUID, offset, limit int) ([]*domain.Listing, error) {
	listings, err := s.listingRepo.ListByOwner(ctx, ownerID, offset, limit)
	if err != nil {
		s.logger.Error("failed to get listings by owner", zap.Error(err), zap.String("owner_id", ownerID.String()))
		return nil, err
	}
	return listings, nil
}

// UpdateListing updates an existing listing
func (s *ListingServiceImpl) UpdateListing(ctx context.Context, id uuid.UUID, title, description string, startingBid, reservePrice int64) error {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("listing not found", zap.Error(err), zap.String("listing_id", id.String()))
		return domain.ErrListingNotFound
	}

	if err := listing.Update(title, description, startingBid, reservePrice); err != nil {
		s.logger.Error("failed to update listing", zap.Error(err))
		return err
	}

	if err := s.listingRepo.Update(ctx, listing); err != nil {
		s.logger.Error("failed to save updated listing", zap.Error(err))
		return err
	}

	s.logger.Info("listing updated", 
		zap.String("listing_id", id.String()),
		zap.String("title", title))

	return nil
}

// DeleteListing deletes a listing
func (s *ListingServiceImpl) DeleteListing(ctx context.Context, id uuid.UUID) error {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("listing not found", zap.Error(err), zap.String("listing_id", id.String()))
		return domain.ErrListingNotFound
	}

	if listing.Status == domain.ListingStatusActive {
		s.logger.Error("cannot delete active listing", zap.String("listing_id", id.String()))
		return domain.ErrInvalidStatusTransition
	}

	if err := s.listingRepo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete listing", zap.Error(err))
		return err
	}

	s.logger.Info("listing deleted", zap.String("listing_id", id.String()))
	return nil
}

// ActivateListing activates a listing
func (s *ListingServiceImpl) ActivateListing(ctx context.Context, id uuid.UUID) error {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("listing not found", zap.Error(err), zap.String("listing_id", id.String()))
		return domain.ErrListingNotFound
	}

	if err := listing.Activate(); err != nil {
		s.logger.Error("failed to activate listing", zap.Error(err))
		return err
	}

	if err := s.listingRepo.Update(ctx, listing); err != nil {
		s.logger.Error("failed to save activated listing", zap.Error(err))
		return err
	}

	s.logger.Info("listing activated", zap.String("listing_id", id.String()))
	return nil
}

// CancelListing cancels a listing
func (s *ListingServiceImpl) CancelListing(ctx context.Context, id uuid.UUID) error {
	listing, err := s.listingRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("listing not found", zap.Error(err), zap.String("listing_id", id.String()))
		return domain.ErrListingNotFound
	}

	if err := listing.Cancel(); err != nil {
		s.logger.Error("failed to cancel listing", zap.Error(err))
		return err
	}

	if err := s.listingRepo.Update(ctx, listing); err != nil {
		s.logger.Error("failed to save cancelled listing", zap.Error(err))
		return err
	}

	s.logger.Info("listing cancelled", zap.String("listing_id", id.String()))
	return nil
}