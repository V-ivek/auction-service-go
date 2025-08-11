package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

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
	// Extract bid data
	auctionIDStr, ok := data["auction_id"].(string)
	if !ok {
		h.sendError(client, "missing or invalid auction_id")
		return
	}

	amountFloat, ok := data["amount"].(float64)
	if !ok {
		h.sendError(client, "missing or invalid amount")
		return
	}
	amount := int64(amountFloat)

	// Parse auction ID
	auctionID, err := uuid.Parse(auctionIDStr)
	if err != nil {
		h.sendError(client, "invalid auction_id format")
		return
	}

	// Get user ID
	userID := client.GetUserID()
	if userID == nil {
		h.sendError(client, "user not authenticated")
		return
	}

	// Place bid
	bid, err := h.biddingService.PlaceBid(ctx, auctionID, *userID, amount)
	if err != nil {
		h.sendError(client, "failed to place bid: "+err.Error())
		return
	}

	// Send success response to client
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

	h.logger.Info("bid placed via websocket",
		zap.String("client_id", client.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("auction_id", auctionID.String()),
		zap.Int64("amount", amount))
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

	// Send success response
	client.SendMessage(wsAdapter.Message{
		Type: responseType,
		Data: map[string]interface{}{
			"listing_id": listingID.String(),
			"user_id":    userID.String(),
			"message":    successMessage,
		},
		Timestamp: h.getCurrentTime(),
	})

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