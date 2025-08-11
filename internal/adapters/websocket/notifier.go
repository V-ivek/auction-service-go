package websocket

import (
	"github.com/google/uuid"

	"auction-microservice/internal/domain"
	"auction-microservice/internal/ports"
)

// Notifier implements the ports.Notifier interface using WebSocket
type Notifier struct {
	hub *Hub
}

// NewNotifier creates a new WebSocket notifier
func NewNotifier(hub *Hub) ports.Notifier {
	return &Notifier{
		hub: hub,
	}
}

// NotifyBidPlaced notifies all subscribers about a new bid
func (n *Notifier) NotifyBidPlaced(event *domain.DomainEvent) error {
	data := event.Data
	auctionIDStr, ok := data["auction_id"].(uuid.UUID)
	if !ok {
		return nil
	}

	// Prepare notification data
	notificationData := map[string]interface{}{
		"event_id":     event.ID.String(),
		"event_type":   event.Type,
		"timestamp":    event.Timestamp,
		"auction_id":   data["auction_id"],
		"bid_id":       data["bid_id"],
		"bidder_id":    data["bidder_id"],
		"amount":       data["amount"],
		"bid_time":     data["bid_time"],
		"new_high":     data["new_high"],
		"bid_count":    data["bid_count"],
	}

	// Send to auction room
	n.hub.BroadcastToAuction(auctionIDStr, "bid_placed", notificationData)

	return nil
}

// NotifyAuctionStarted notifies all subscribers about an auction start
func (n *Notifier) NotifyAuctionStarted(event *domain.DomainEvent) error {
	data := event.Data
	auctionIDStr, ok := data["auction_id"].(uuid.UUID)
	if !ok {
		return nil
	}

	// Prepare notification data
	notificationData := map[string]interface{}{
		"event_id":     event.ID.String(),
		"event_type":   event.Type,
		"timestamp":    event.Timestamp,
		"auction_id":   data["auction_id"],
		"listing_id":   data["listing_id"],
		"start_time":   data["start_time"],
		"end_time":     data["end_time"],
		"duration":     data["duration"],
	}

	// Send to auction room
	n.hub.BroadcastToAuction(auctionIDStr, "auction_started", notificationData)

	// Also broadcast to all connected clients
	n.hub.Broadcast <- Message{
		Type:      "auction_started",
		Data:      notificationData,
		Timestamp: event.Timestamp,
	}

	return nil
}

// NotifyAuctionEnded notifies all subscribers about an auction end
func (n *Notifier) NotifyAuctionEnded(event *domain.DomainEvent) error {
	data := event.Data
	auctionIDStr, ok := data["auction_id"].(uuid.UUID)
	if !ok {
		return nil
	}

	// Prepare notification data
	notificationData := map[string]interface{}{
		"event_id":          event.ID.String(),
		"event_type":        event.Type,
		"timestamp":         event.Timestamp,
		"auction_id":        data["auction_id"],
		"listing_id":        data["listing_id"],
		"final_bid":         data["final_bid"],
		"winner_id":         data["winner_id"],
		"bid_count":         data["bid_count"],
		"completion_time":   data["completion_time"],
	}

	// Send to auction room
	n.hub.BroadcastToAuction(auctionIDStr, "auction_ended", notificationData)

	// Notify winner specifically if there is one
	if winnerIDData, exists := data["winner_id"]; exists && winnerIDData != nil {
		if winnerID, ok := winnerIDData.(uuid.UUID); ok {
			winnerNotification := map[string]interface{}{
				"event_id":      event.ID.String(),
				"event_type":    "auction_won",
				"timestamp":     event.Timestamp,
				"auction_id":    data["auction_id"],
				"listing_id":    data["listing_id"],
				"winning_bid":   data["final_bid"],
				"congratulations": "You won the auction!",
			}

			n.hub.SendToUserID(winnerID, "auction_won", winnerNotification)
		}
	}

	return nil
}

// NotifyAuctionExtended notifies all subscribers about an auction extension
func (n *Notifier) NotifyAuctionExtended(event *domain.DomainEvent) error {
	data := event.Data
	auctionIDStr, ok := data["auction_id"].(uuid.UUID)
	if !ok {
		return nil
	}

	// Prepare notification data
	notificationData := map[string]interface{}{
		"event_id":      event.ID.String(),
		"event_type":    "auction_extended",
		"timestamp":     event.Timestamp,
		"auction_id":    data["auction_id"],
		"new_end_time":  data["new_end_time"],
		"extension":     data["extension"],
		"reason":        "Anti-sniping protection",
	}

	// Send to auction room
	n.hub.BroadcastToAuction(auctionIDStr, "auction_extended", notificationData)

	return nil
}

// NotifyBidOutbid notifies a specific user that their bid was outbid
func (n *Notifier) NotifyBidOutbid(event *domain.DomainEvent) error {
	data := event.Data
	outbidBidderIDData, ok := data["outbid_bidder_id"]
	if !ok {
		return nil
	}

	outbidBidderID, ok := outbidBidderIDData.(uuid.UUID)
	if !ok {
		return nil
	}

	// Prepare notification data
	notificationData := map[string]interface{}{
		"event_id":         event.ID.String(),
		"event_type":       event.Type,
		"timestamp":        event.Timestamp,
		"auction_id":       data["auction_id"],
		"outbid_bid_id":    data["outbid_bid_id"],
		"outbid_amount":    data["outbid_amount"],
		"new_high_amount":  data["new_high_amount"],
		"message":          "Your bid has been outbid",
	}

	// Send to specific user
	n.hub.SendToUserID(outbidBidderID, "bid_outbid", notificationData)

	return nil
}

// NotifyListingUpdated notifies all subscribers about a listing update
func (n *Notifier) NotifyListingUpdated(event *domain.DomainEvent) error {
	// Prepare notification data
	notificationData := map[string]interface{}{
		"event_id":     event.ID.String(),
		"event_type":   event.Type,
		"timestamp":    event.Timestamp,
		"listing_id":   event.Data["listing_id"],
		"title":        event.Data["title"],
		"status":       event.Data["status"],
		"update_time":  event.Data["update_time"],
	}

	// Broadcast to all connected clients
	n.hub.Broadcast <- Message{
		Type:      "listing_updated",
		Data:      notificationData,
		Timestamp: event.Timestamp,
	}

	return nil
}

// NotifyGeneric sends a generic notification to all connected clients for a specific auction
func (n *Notifier) NotifyGeneric(auctionID string, messageType string, data interface{}) error {
	auctionUUID, err := uuid.Parse(auctionID)
	if err != nil {
		return err
	}

	n.hub.BroadcastToAuction(auctionUUID, messageType, data)
	return nil
}

// NotifyUser sends a notification to a specific user
func (n *Notifier) NotifyUser(userID string, messageType string, data interface{}) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	n.hub.SendToUserID(userUUID, messageType, data)
	return nil
}