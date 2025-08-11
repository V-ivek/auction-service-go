package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"auction-microservice/internal/domain"
	wsAdapter "auction-microservice/internal/adapters/websocket"
	"auction-microservice/internal/ports"
)

// WebSocketHandler handles WebSocket connections and messages
type WebSocketHandler struct {
	hub                 *wsAdapter.Hub
	biddingService      ports.BiddingService
	subscriptionService ports.SubscriptionService
	auctionService      ports.AuctionService
	authMiddleware      AuthMiddleware
	logger              *zap.Logger
}

// AuthMiddleware interface for authentication
type AuthMiddleware interface {
	ValidateToken(tokenString string) (uuid.UUID, error)
}

// BidMessage represents a bid message from WebSocket
type BidMessage struct {
	AuctionID string `json:"auction_id"`
	Amount    int64  `json:"amount"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	hub *wsAdapter.Hub,
	biddingService ports.BiddingService,
	subscriptionService ports.SubscriptionService,
	auctionService ports.AuctionService,
	authMiddleware AuthMiddleware,
	logger *zap.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:                 hub,
		biddingService:      biddingService,
		subscriptionService: subscriptionService,
		auctionService:      auctionService,
		authMiddleware:      authMiddleware,
		logger:              logger,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade to websocket", zap.Error(err))
		return
	}

	// Create new client
	client := wsAdapter.NewClient(h.hub, conn, h.logger)

	// Try to authenticate client from query parameters or headers
	h.authenticateClient(r, client)

	// Register client with hub
	h.hub.Register <- client

	h.logger.Info("websocket client connected",
		zap.String("client_id", client.ID.String()),
		zap.String("remote_addr", r.RemoteAddr))

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	// Start message handler for this client
	go h.handleClientMessages(client)
}

// authenticateClient attempts to authenticate the client
func (h *WebSocketHandler) authenticateClient(r *http.Request, client *wsAdapter.Client) {
	// Try to get token from query parameters
	token := r.URL.Query().Get("token")
	
	// If not in query params, try Authorization header
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token != "" && h.authMiddleware != nil {
		if userID, err := h.authMiddleware.ValidateToken(token); err == nil {
			client.SetUserID(&userID)
			h.logger.Info("websocket client authenticated",
				zap.String("client_id", client.ID.String()),
				zap.String("user_id", userID.String()))
		} else {
			h.logger.Debug("websocket client authentication failed",
				zap.String("client_id", client.ID.String()),
				zap.Error(err))
		}
	}
}

// handleClientMessages handles messages from a specific client
func (h *WebSocketHandler) handleClientMessages(client *wsAdapter.Client) {
	// This function could be extended to handle custom client-specific message processing
	// For now, the client handles messages directly in its ReadPump method
	
	// Wait for client context to be done
	<-client.Ctx.Done()
	
	h.logger.Debug("client message handler stopped",
		zap.String("client_id", client.ID.String()))
}

// ProcessBidMessage processes a bid message from WebSocket
func (h *WebSocketHandler) ProcessBidMessage(ctx context.Context, client *wsAdapter.Client, data map[string]interface{}) {
	h.logger.Info("🚀 ProcessBidMessage called",
		zap.String("client_id", client.ID.String()),
		zap.Any("data", data))

	// Extract bid data
	auctionIDStr, ok := data["auction_id"].(string)
	if !ok {
		h.logger.Error("❌ missing or invalid auction_id in bid data",
			zap.String("client_id", client.ID.String()),
			zap.Any("data", data))
		h.sendError(client, "missing or invalid auction_id")
		return
	}

	amountFloat, ok := data["amount"].(float64)
	if !ok {
		h.logger.Error("❌ missing or invalid amount in bid data",
			zap.String("client_id", client.ID.String()),
			zap.Any("data", data),
			zap.Any("amount", data["amount"]))
		h.sendError(client, "missing or invalid amount")
		return
	}
	amount := int64(amountFloat)

	h.logger.Info("📊 bid data extracted",
		zap.String("client_id", client.ID.String()),
		zap.String("auction_id", auctionIDStr),
		zap.Int64("amount", amount),
		zap.Float64("amount_float", amountFloat))

	// Parse auction ID
	auctionID, err := uuid.Parse(auctionIDStr)
	if err != nil {
		h.logger.Error("❌ invalid auction_id format",
			zap.String("client_id", client.ID.String()),
			zap.String("auction_id_str", auctionIDStr),
			zap.Error(err))
		h.sendError(client, "invalid auction_id format")
		return
	}

	// Get user ID
	userID := client.GetUserID()
	if userID == nil {
		h.logger.Error("❌ user not authenticated",
			zap.String("client_id", client.ID.String()))
		h.sendError(client, "user not authenticated")
		return
	}

	h.logger.Info("✅ proceeding with bid placement",
		zap.String("client_id", client.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("auction_id", auctionID.String()),
		zap.Int64("amount", amount))

	// Place bid
	bid, err := h.biddingService.PlaceBid(ctx, auctionID, *userID, amount)
	if err != nil {
		h.logger.Error("❌ failed to place bid",
			zap.String("client_id", client.ID.String()),
			zap.String("user_id", userID.String()),
			zap.String("auction_id", auctionID.String()),
			zap.Int64("amount", amount),
			zap.Error(err))
		h.sendError(client, "failed to place bid: "+err.Error())
		return
	}

	h.logger.Info("🎉 bid placed successfully via websocket",
		zap.String("client_id", client.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("auction_id", auctionID.String()),
		zap.String("bid_id", bid.ID.String()),
		zap.Int64("amount", amount))

	// Send success response to the bidding client
	client.SendMessage(wsAdapter.Message{
		Type: "bid_placed_success",
		Data: map[string]interface{}{
			"bid_id":     bid.ID.String(),
			"auction_id": auctionID.String(),
			"amount":     amount,
			"status":     "success",
		},
		Timestamp: bid.PlacedAt,
	})

	// 🚀 EVENT-DRIVEN: Broadcast bid update to ALL connected users
	h.broadcastBidUpdate(auctionID, bid, *userID)

	h.logger.Info("📢 bid update broadcasted to all users",
		zap.String("auction_id", auctionID.String()),
		zap.Int64("amount", amount),
		zap.String("bidder_user_id", userID.String()))
}

// ProcessSubscriptionMessage processes a subscription message
func (h *WebSocketHandler) ProcessSubscriptionMessage(ctx context.Context, client *wsAdapter.Client, messageType string, data map[string]interface{}) {
	listingIDStr, ok := data["listing_id"].(string)
	if !ok {
		h.sendError(client, "missing or invalid listing_id")
		return
	}

	listingID, err := uuid.Parse(listingIDStr)
	if err != nil {
		h.sendError(client, "invalid listing_id format")
		return
	}

	userID := client.GetUserID()
	if userID == nil {
		h.sendError(client, "user not authenticated")
		return
	}

	var responseType string
	var successMessage string

	switch messageType {
	case "subscribe":
		err = h.subscriptionService.Subscribe(ctx, *userID, listingID)
		responseType = "subscription_success"
		successMessage = "successfully subscribed to listing"

	case "unsubscribe":
		err = h.subscriptionService.Unsubscribe(ctx, *userID, listingID)
		responseType = "unsubscription_success"
		successMessage = "successfully unsubscribed from listing"

	default:
		h.sendError(client, "unknown subscription action")
		return
	}

	if err != nil {
		h.sendError(client, "subscription operation failed: "+err.Error())
		return
	}

	// Send success response to client
	client.SendMessage(wsAdapter.Message{
		Type: responseType,
		Data: map[string]interface{}{
			"listing_id": listingID.String(),
			"user_id":    userID.String(),
			"message":    successMessage,
		},
		Timestamp: h.getCurrentTime(),
	})

	// 🚀 EVENT-DRIVEN: Broadcast subscription update to all users
	if messageType == "subscribe" {
		h.BroadcastToAll("user_subscribed", map[string]interface{}{
			"listing_id": listingID.String(),
			"user_id":    userID.String(),
			"timestamp":  h.getCurrentTime(),
		})
	}

	h.logger.Info("subscription operation completed",
		zap.String("client_id", client.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("listing_id", listingID.String()),
		zap.String("operation", messageType))
}

// ProcessAuctionInfoRequest processes a request for auction information
func (h *WebSocketHandler) ProcessAuctionInfoRequest(ctx context.Context, client *wsAdapter.Client, data map[string]interface{}) {
	auctionIDStr, ok := data["auction_id"].(string)
	if !ok {
		h.sendError(client, "missing or invalid auction_id")
		return
	}

	auctionID, err := uuid.Parse(auctionIDStr)
	if err != nil {
		h.sendError(client, "invalid auction_id format")
		return
	}

	// Get auction information
	auction, err := h.auctionService.GetAuction(ctx, auctionID)
	if err != nil {
		h.sendError(client, "failed to get auction info: "+err.Error())
		return
	}

	// Get bid history
	bidHistory, err := h.biddingService.GetBidHistory(ctx, auctionID)
	if err != nil {
		h.logger.Error("failed to get bid history", zap.Error(err))
		bidHistory = nil // Continue without bid history
	}

	// Send auction info
	client.SendMessage(wsAdapter.Message{
		Type: "auction_info",
		Data: map[string]interface{}{
			"auction":     auction,
			"bid_history": bidHistory,
			"room_size":   h.hub.GetAuctionRoomSize(auctionID),
		},
		Timestamp: h.getCurrentTime(),
	})
}

// sendError sends an error message to a client
func (h *WebSocketHandler) sendError(client *wsAdapter.Client, errorMsg string) {
	client.SendMessage(wsAdapter.Message{
		Type: "error",
		Data: map[string]interface{}{
			"message": errorMsg,
		},
		Timestamp: h.getCurrentTime(),
	})
}

// getCurrentTime returns the current time (can be mocked for testing)
func (h *WebSocketHandler) getCurrentTime() time.Time {
	return time.Now()
}

// BroadcastAuctionUpdate broadcasts an auction update to all clients in the auction room
func (h *WebSocketHandler) BroadcastAuctionUpdate(auctionID uuid.UUID, updateType string, data interface{}) {
	h.hub.BroadcastToAuction(auctionID, updateType, data)
}

// BroadcastToAll broadcasts a message to all connected clients
func (h *WebSocketHandler) BroadcastToAll(messageType string, data interface{}) {
	message := wsAdapter.Message{
		Type:      messageType,
		Data:      data,
		Timestamp: h.getCurrentTime(),
	}

	h.hub.Broadcast <- message
}

// SendToUser sends a message to a specific user
func (h *WebSocketHandler) SendToUser(userID uuid.UUID, messageType string, data interface{}) {
	h.hub.SendToUserID(userID, messageType, data)
}

// GetActiveConnections returns the number of active connections
func (h *WebSocketHandler) GetActiveConnections() int {
	return h.hub.GetActiveConnections()
}

// GetAuctionRoomSize returns the number of clients in an auction room
func (h *WebSocketHandler) GetAuctionRoomSize(auctionID uuid.UUID) int {
	return h.hub.GetAuctionRoomSize(auctionID)
}

// broadcastBidUpdate broadcasts a bid update to all connected users (event-driven)
func (h *WebSocketHandler) broadcastBidUpdate(auctionID uuid.UUID, bid *domain.Bid, bidderUserID uuid.UUID) {
	// Get auction info to include in the broadcast
	ctx := context.Background()
	auction, err := h.auctionService.GetAuction(ctx, auctionID)
	if err != nil {
		h.logger.Error("failed to get auction for bid broadcast", 
			zap.String("auction_id", auctionID.String()),
			zap.Error(err))
		return
	}

	// Create the bid update event data
	bidUpdateData := map[string]interface{}{
		"bid_id":         bid.ID.String(),
		"auction_id":     auctionID.String(), 
		"bidder_user_id": bidderUserID.String(),
		"amount":         bid.Amount,
		"timestamp":      bid.PlacedAt,
		"current_bid":    auction.CurrentBid,
		"bid_count":      auction.BidCount,
	}

	// 🚀 Broadcast to ALL connected users
	h.BroadcastToAll("bid_placed", bidUpdateData)

	h.logger.Debug("broadcasted bid update to all users",
		zap.String("auction_id", auctionID.String()),
		zap.String("bidder_user_id", bidderUserID.String()),
		zap.Int64("amount", bid.Amount),
		zap.Int("active_connections", h.GetActiveConnections()))
}