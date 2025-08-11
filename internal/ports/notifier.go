package ports

import (
	"auction-microservice/internal/domain"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	// NotifyBidPlaced notifies all subscribers about a new bid
	NotifyBidPlaced(event *domain.DomainEvent) error

	// NotifyAuctionStarted notifies all subscribers about an auction start
	NotifyAuctionStarted(event *domain.DomainEvent) error

	// NotifyAuctionEnded notifies all subscribers about an auction end
	NotifyAuctionEnded(event *domain.DomainEvent) error

	// NotifyAuctionExtended notifies all subscribers about an auction extension
	NotifyAuctionExtended(event *domain.DomainEvent) error

	// NotifyBidOutbid notifies a specific user that their bid was outbid
	NotifyBidOutbid(event *domain.DomainEvent) error

	// NotifyListingUpdated notifies all subscribers about a listing update
	NotifyListingUpdated(event *domain.DomainEvent) error

	// NotifyGeneric sends a generic notification to all connected clients for a specific auction
	NotifyGeneric(auctionID string, messageType string, data interface{}) error

	// NotifyUser sends a notification to a specific user
	NotifyUser(userID string, messageType string, data interface{}) error
}