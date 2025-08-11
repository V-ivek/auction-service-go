package services

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/internal/domain"
	"auction-microservice/internal/ports"
)

// SubscriptionServiceImpl implements the SubscriptionService interface
type SubscriptionServiceImpl struct {
	subscriptionRepo ports.SubscriptionRepository
	listingRepo      ports.ListingRepository
	userRepo         ports.UserRepository
	logger           *zap.Logger
}

// NewSubscriptionService creates a new instance of SubscriptionService
func NewSubscriptionService(
	subscriptionRepo ports.SubscriptionRepository,
	listingRepo ports.ListingRepository,
	userRepo ports.UserRepository,
	logger *zap.Logger,
) *SubscriptionServiceImpl {
	return &SubscriptionServiceImpl{
		subscriptionRepo: subscriptionRepo,
		listingRepo:      listingRepo,
		userRepo:         userRepo,
		logger:           logger,
	}
}

// Subscribe subscribes a user to a listing
func (s *SubscriptionServiceImpl) Subscribe(ctx context.Context, userID, listingID uuid.UUID) error {
	// Validate user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("user not found", zap.Error(err), zap.String("user_id", userID.String()))
		return domain.ErrUserNotFound
	}

	// Validate listing exists
	_, err = s.listingRepo.GetByID(ctx, listingID)
	if err != nil {
		s.logger.Error("listing not found", zap.Error(err), zap.String("listing_id", listingID.String()))
		return domain.ErrListingNotFound
	}

	// Check if already subscribed
	isSubscribed, err := s.subscriptionRepo.IsSubscribed(ctx, userID, listingID)
	if err != nil {
		s.logger.Error("failed to check subscription status", zap.Error(err))
		return err
	}

	if isSubscribed {
		s.logger.Debug("user already subscribed", 
			zap.String("user_id", userID.String()),
			zap.String("listing_id", listingID.String()))
		return domain.ErrAlreadySubscribed
	}

	// Create subscription
	if err := s.subscriptionRepo.Create(ctx, userID, listingID); err != nil {
		s.logger.Error("failed to create subscription", zap.Error(err))
		return err
	}

	s.logger.Info("user subscribed to listing", 
		zap.String("user_id", userID.String()),
		zap.String("listing_id", listingID.String()))

	return nil
}

// Unsubscribe unsubscribes a user from a listing
func (s *SubscriptionServiceImpl) Unsubscribe(ctx context.Context, userID, listingID uuid.UUID) error {
	// Check if subscribed
	isSubscribed, err := s.subscriptionRepo.IsSubscribed(ctx, userID, listingID)
	if err != nil {
		s.logger.Error("failed to check subscription status", zap.Error(err))
		return err
	}

	if !isSubscribed {
		s.logger.Debug("user not subscribed", 
			zap.String("user_id", userID.String()),
			zap.String("listing_id", listingID.String()))
		return domain.ErrSubscriptionNotFound
	}

	// Remove subscription
	if err := s.subscriptionRepo.Delete(ctx, userID, listingID); err != nil {
		s.logger.Error("failed to delete subscription", zap.Error(err))
		return err
	}

	s.logger.Info("user unsubscribed from listing", 
		zap.String("user_id", userID.String()),
		zap.String("listing_id", listingID.String()))

	return nil
}

// GetSubscriptions retrieves all listings a user is subscribed to
func (s *SubscriptionServiceImpl) GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("user not found", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, domain.ErrUserNotFound
	}

	subscriptions, err := s.subscriptionRepo.GetSubscriptions(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get subscriptions", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}

	return subscriptions, nil
}

// GetSubscribers retrieves all users subscribed to a listing
func (s *SubscriptionServiceImpl) GetSubscribers(ctx context.Context, listingID uuid.UUID) ([]uuid.UUID, error) {
	// Validate listing exists
	_, err := s.listingRepo.GetByID(ctx, listingID)
	if err != nil {
		s.logger.Error("listing not found", zap.Error(err), zap.String("listing_id", listingID.String()))
		return nil, domain.ErrListingNotFound
	}

	subscribers, err := s.subscriptionRepo.GetSubscribers(ctx, listingID)
	if err != nil {
		s.logger.Error("failed to get subscribers", zap.Error(err), zap.String("listing_id", listingID.String()))
		return nil, err
	}

	return subscribers, nil
}

// IsSubscribed checks if a user is subscribed to a listing
func (s *SubscriptionServiceImpl) IsSubscribed(ctx context.Context, userID, listingID uuid.UUID) (bool, error) {
	isSubscribed, err := s.subscriptionRepo.IsSubscribed(ctx, userID, listingID)
	if err != nil {
		s.logger.Error("failed to check subscription status", zap.Error(err))
		return false, err
	}

	return isSubscribed, nil
}