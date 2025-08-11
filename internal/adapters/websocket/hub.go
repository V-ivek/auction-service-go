package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/internal/ports"
)

// AuctionMessage represents a message for auction operations
type AuctionMessage struct {
	Client    *Client
	AuctionID uuid.UUID
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Auction rooms - maps auction ID to set of clients
	auctionRooms map[uuid.UUID]map[*Client]bool

	// User connections - maps user ID to set of clients
	userConnections map[uuid.UUID]map[*Client]bool

	// Inbound messages from the clients
	Broadcast chan Message

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Join auction room
	JoinAuction chan *AuctionMessage

	// Leave auction room
	LeaveAuction chan *AuctionMessage

	// Send message to specific auction
	SendToAuction chan *AuctionMessage

	// Send message to specific user
	SendToUser chan *UserMessage

	// Process client messages
	ProcessClientMessage chan *ClientMessage

	logger             *zap.Logger
	subscriptionRepo   ports.SubscriptionRepository
	wsHandler          WebSocketMessageHandler
	mu                 sync.RWMutex
	ctx                context.Context
	cancel             context.CancelFunc
	activeConnections  int
}

// UserMessage represents a message for a specific user
type UserMessage struct {
	UserID  uuid.UUID
	Message Message
}

// ClientMessage represents a message from a client that needs processing
type ClientMessage struct {
	Client  *Client
	Message Message
}

// WebSocketMessageHandler interface for handling WebSocket messages
type WebSocketMessageHandler interface {
	ProcessBidMessage(ctx context.Context, client *Client, data map[string]interface{})
	ProcessSubscriptionMessage(ctx context.Context, client *Client, messageType string, data map[string]interface{})
}

// NewHub creates a new Hub
func NewHub(logger *zap.Logger, subscriptionRepo ports.SubscriptionRepository) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Hub{
		clients:              make(map[*Client]bool),
		auctionRooms:         make(map[uuid.UUID]map[*Client]bool),
		userConnections:      make(map[uuid.UUID]map[*Client]bool),
		Broadcast:            make(chan Message, 100),
		Register:             make(chan *Client, 10),
		Unregister:           make(chan *Client, 10),
		JoinAuction:          make(chan *AuctionMessage, 10),
		LeaveAuction:         make(chan *AuctionMessage, 10),
		SendToAuction:        make(chan *AuctionMessage, 10),
		SendToUser:           make(chan *UserMessage, 10),
		ProcessClientMessage: make(chan *ClientMessage, 100),
		logger:               logger,
		subscriptionRepo:     subscriptionRepo,
		ctx:                  ctx,
		cancel:               cancel,
	}
}

// SetMessageHandler sets the WebSocket message handler
func (h *Hub) SetMessageHandler(handler WebSocketMessageHandler) {
	h.wsHandler = handler
}

// Run starts the hub and processes messages
func (h *Hub) Run(ctx context.Context) {
	defer h.cancel()

	h.logger.Info("starting websocket hub")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("websocket hub stopping")
			return

		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)

		case auctionMsg := <-h.JoinAuction:
			h.joinAuction(auctionMsg)

		case auctionMsg := <-h.LeaveAuction:
			h.leaveAuction(auctionMsg)

		case auctionMsg := <-h.SendToAuction:
			h.sendToAuction(auctionMsg)

		case userMsg := <-h.SendToUser:
			h.sendToUser(userMsg)

		case clientMsg := <-h.ProcessClientMessage:
			h.logger.Info("received client message in hub main loop",
				zap.String("client_id", clientMsg.Client.ID.String()),
				zap.String("message_type", clientMsg.Message.Type))
			h.processClientMessage(clientMsg)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	h.activeConnections++

	// If client has a user ID, add to user connections
	if userID := client.GetUserID(); userID != nil {
		if h.userConnections[*userID] == nil {
			h.userConnections[*userID] = make(map[*Client]bool)
		}
		h.userConnections[*userID][client] = true
	}

	h.logger.Info("client registered",
		zap.String("client_id", client.ID.String()),
		zap.Int("total_connections", h.activeConnections))

	// Send welcome message
	client.SendMessage(Message{
		Type: "welcome",
		Data: map[string]interface{}{
			"client_id": client.ID.String(),
			"message":   "Connected to auction service",
		},
		Timestamp: time.Now(),
	})
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		h.activeConnections--

		// Remove from user connections
		if userID := client.GetUserID(); userID != nil {
			if userClients := h.userConnections[*userID]; userClients != nil {
				delete(userClients, client)
				if len(userClients) == 0 {
					delete(h.userConnections, *userID)
				}
			}
		}

		// Remove from auction rooms
		if auctionID := client.GetAuctionID(); auctionID != nil {
			if auctionClients := h.auctionRooms[*auctionID]; auctionClients != nil {
				delete(auctionClients, client)
				if len(auctionClients) == 0 {
					delete(h.auctionRooms, *auctionID)
				}
			}
		}

		close(client.Send)

		h.logger.Info("client unregistered",
			zap.String("client_id", client.ID.String()),
			zap.Int("total_connections", h.activeConnections))
	}
}

// broadcastMessage broadcasts a message to all connected clients
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, close it
			delete(h.clients, client)
			close(client.Send)
		}
	}

	h.logger.Debug("message broadcasted to all clients",
		zap.String("message_type", message.Type),
		zap.Int("client_count", len(h.clients)))
}

// joinAuction adds a client to an auction room
func (h *Hub) joinAuction(auctionMsg *AuctionMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := auctionMsg.Client
	auctionID := auctionMsg.AuctionID

	if h.auctionRooms[auctionID] == nil {
		h.auctionRooms[auctionID] = make(map[*Client]bool)
	}

	h.auctionRooms[auctionID][client] = true

	h.logger.Debug("client joined auction room",
		zap.String("client_id", client.ID.String()),
		zap.String("auction_id", auctionID.String()),
		zap.Int("room_size", len(h.auctionRooms[auctionID])))

	// Send confirmation to client
	client.SendMessage(Message{
		Type: "auction_joined",
		Data: map[string]interface{}{
			"auction_id": auctionID.String(),
			"room_size":  len(h.auctionRooms[auctionID]),
		},
		Timestamp: time.Now(),
	})

	// Notify other clients in the room
	joinMessage := Message{
		Type: "user_joined",
		Data: map[string]interface{}{
			"auction_id": auctionID.String(),
			"client_id":  client.ID.String(),
			"room_size":  len(h.auctionRooms[auctionID]),
		},
		Timestamp: time.Now(),
	}

	for roomClient := range h.auctionRooms[auctionID] {
		if roomClient != client {
			roomClient.SendMessage(joinMessage)
		}
	}
}

// leaveAuction removes a client from an auction room
func (h *Hub) leaveAuction(auctionMsg *AuctionMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := auctionMsg.Client
	auctionID := auctionMsg.AuctionID

	if auctionClients := h.auctionRooms[auctionID]; auctionClients != nil {
		delete(auctionClients, client)

		h.logger.Debug("client left auction room",
			zap.String("client_id", client.ID.String()),
			zap.String("auction_id", auctionID.String()),
			zap.Int("room_size", len(auctionClients)))

		// Send confirmation to client
		client.SendMessage(Message{
			Type: "auction_left",
			Data: map[string]interface{}{
				"auction_id": auctionID.String(),
			},
			Timestamp: time.Now(),
		})

		// Notify other clients in the room
		leaveMessage := Message{
			Type: "user_left",
			Data: map[string]interface{}{
				"auction_id": auctionID.String(),
				"client_id":  client.ID.String(),
				"room_size":  len(auctionClients),
			},
			Timestamp: time.Now(),
		}

		for roomClient := range auctionClients {
			roomClient.SendMessage(leaveMessage)
		}

		// Clean up empty room
		if len(auctionClients) == 0 {
			delete(h.auctionRooms, auctionID)
		}
	}
}

// sendToAuction sends a message to all clients in an auction room
func (h *Hub) sendToAuction(auctionMsg *AuctionMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	auctionID := auctionMsg.AuctionID

	if auctionClients := h.auctionRooms[auctionID]; auctionClients != nil {
		message := Message{
			Type:      "auction_update",
			Data:      auctionMsg,
			Timestamp: time.Now(),
		}
		
		for client := range auctionClients {
			client.SendMessage(message)
		}

		h.logger.Debug("message sent to auction room",
			zap.String("auction_id", auctionID.String()),
			zap.Int("client_count", len(auctionClients)))
	}
}

// sendToUser sends a message to all connections for a specific user
func (h *Hub) sendToUser(userMsg *UserMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userID := userMsg.UserID
	message := userMsg.Message

	if userClients := h.userConnections[userID]; userClients != nil {
		for client := range userClients {
			client.SendMessage(message)
		}

		h.logger.Debug("message sent to user",
			zap.String("user_id", userID.String()),
			zap.String("message_type", message.Type),
			zap.Int("client_count", len(userClients)))
	}
}

// GetActiveConnections returns the number of active connections
func (h *Hub) GetActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.activeConnections
}

// GetAuctionRoomSize returns the number of clients in an auction room
func (h *Hub) GetAuctionRoomSize(auctionID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if auctionClients := h.auctionRooms[auctionID]; auctionClients != nil {
		return len(auctionClients)
	}

	return 0
}

// BroadcastToAuction broadcasts a message to all clients in an auction room
func (h *Hub) BroadcastToAuction(auctionID uuid.UUID, messageType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message := Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}

	if auctionClients := h.auctionRooms[auctionID]; auctionClients != nil {
		for client := range auctionClients {
			client.SendMessage(message)
		}

		h.logger.Debug("message broadcasted to auction room",
			zap.String("auction_id", auctionID.String()),
			zap.String("message_type", messageType),
			zap.Int("client_count", len(auctionClients)))
	}
}

// SendToUserID sends a message to a specific user
func (h *Hub) SendToUserID(userID uuid.UUID, messageType string, data interface{}) {
	message := Message{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case h.SendToUser <- &UserMessage{
		UserID:  userID,
		Message: message,
	}:
		// Message queued successfully
	default:
		h.logger.Warn("failed to queue user message, channel full",
			zap.String("user_id", userID.String()),
			zap.String("message_type", messageType))
	}
}

// Stop gracefully stops the hub
func (h *Hub) Stop() {
	h.cancel()
	
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all client connections
	for client := range h.clients {
		client.Close()
	}

	h.logger.Info("websocket hub stopped")
}

// processClientMessage processes messages from clients
func (h *Hub) processClientMessage(clientMsg *ClientMessage) {
	if h.wsHandler == nil {
		h.logger.Warn("no WebSocket message handler set")
		return
	}

	client := clientMsg.Client
	message := clientMsg.Message

	h.logger.Info("processing client message in hub",
		zap.String("client_id", client.ID.String()),
		zap.String("message_type", message.Type),
		zap.Any("message_data", message.Data))

	data, ok := message.Data.(map[string]interface{})
	if !ok {
		h.logger.Error("invalid message data format",
			zap.String("client_id", client.ID.String()),
			zap.String("message_type", message.Type),
			zap.Any("message_data", message.Data))
		client.SendMessage(Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "invalid message data format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	ctx := context.Background()

	switch message.Type {
	case "place_bid":
		h.logger.Info("processing place_bid message",
			zap.String("client_id", client.ID.String()),
			zap.String("auction_id", data["auction_id"].(string)),
			zap.Any("amount", data["amount"]))
		h.wsHandler.ProcessBidMessage(ctx, client, data)
	case "subscribe", "unsubscribe":
		h.logger.Info("processing subscription message",
			zap.String("client_id", client.ID.String()),
			zap.String("message_type", message.Type),
			zap.String("listing_id", data["listing_id"].(string)))
		h.wsHandler.ProcessSubscriptionMessage(ctx, client, message.Type, data)
	default:
		h.logger.Debug("unhandled message type",
			zap.String("client_id", client.ID.String()),
			zap.String("message_type", message.Type))
	}
}