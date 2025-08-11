package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (configure for production)
		return true
	},
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	ID         uuid.UUID
	UserID     *uuid.UUID
	AuctionID  *uuid.UUID
	Hub        *Hub
	Conn       *websocket.Conn
	Send       chan Message
	logger     *zap.Logger
	mu         sync.RWMutex
	Ctx        context.Context
	cancel     context.CancelFunc
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, logger *zap.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &Client{
		ID:     uuid.New(),
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan Message, 256),
		logger: logger,
		Ctx:    ctx,
		cancel: cancel,
	}

	return client
}

// GetUserID returns the user ID in a thread-safe manner
func (c *Client) GetUserID() *uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.UserID
}

// SetUserID sets the user ID in a thread-safe manner
func (c *Client) SetUserID(userID *uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.UserID = userID
}

// GetAuctionID returns the auction ID in a thread-safe manner
func (c *Client) GetAuctionID() *uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AuctionID
}

// SetAuctionID sets the auction ID in a thread-safe manner
func (c *Client) SetAuctionID(auctionID *uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AuctionID = auctionID
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
		c.cancel()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.Ctx.Done():
			return
		default:
			var message Message
			err := c.Conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("websocket unexpected close error", zap.Error(err))
				}
				return
			}

			c.logger.Debug("received websocket message",
				zap.String("client_id", c.ID.String()),
				zap.String("message_type", message.Type))

			// Handle different message types
			c.handleMessage(message)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.cancel()
	}()

	for {
		select {
		case <-c.Ctx.Done():
			return

		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				c.logger.Error("failed to write websocket message", zap.Error(err))
				return
			}

			c.logger.Debug("sent websocket message",
				zap.String("client_id", c.ID.String()),
				zap.String("message_type", message.Type))

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (c *Client) handleMessage(message Message) {
	switch message.Type {
	case "join_auction":
		c.handleJoinAuction(message)
	case "leave_auction":
		c.handleLeaveAuction(message)
	case "place_bid":
		c.handlePlaceBid(message)
	case "authenticate":
		c.handleAuthenticate(message)
	default:
		c.logger.Debug("unknown message type",
			zap.String("client_id", c.ID.String()),
			zap.String("message_type", message.Type))
	}
}

// handleJoinAuction handles joining an auction room
func (c *Client) handleJoinAuction(message Message) {
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		c.sendError("invalid message data")
		return
	}

	auctionIDStr, ok := data["auction_id"].(string)
	if !ok {
		c.sendError("missing or invalid auction_id")
		return
	}

	auctionID, err := uuid.Parse(auctionIDStr)
	if err != nil {
		c.sendError("invalid auction_id format")
		return
	}

	// Leave current auction if any
	if currentAuctionID := c.GetAuctionID(); currentAuctionID != nil {
		c.Hub.LeaveAuction <- &AuctionMessage{
			Client:    c,
			AuctionID: *currentAuctionID,
		}
	}

	// Join new auction
	c.SetAuctionID(&auctionID)
	c.Hub.JoinAuction <- &AuctionMessage{
		Client:    c,
		AuctionID: auctionID,
	}

	c.logger.Info("client joined auction",
		zap.String("client_id", c.ID.String()),
		zap.String("auction_id", auctionID.String()))
}

// handleLeaveAuction handles leaving an auction room
func (c *Client) handleLeaveAuction(message Message) {
	auctionID := c.GetAuctionID()
	if auctionID == nil {
		c.sendError("not in any auction")
		return
	}

	c.Hub.LeaveAuction <- &AuctionMessage{
		Client:    c,
		AuctionID: *auctionID,
	}

	c.SetAuctionID(nil)

	c.logger.Info("client left auction",
		zap.String("client_id", c.ID.String()),
		zap.String("auction_id", auctionID.String()))
}

// handlePlaceBid handles placing a bid
func (c *Client) handlePlaceBid(message Message) {
	// This would typically forward the bid to the bidding service
	// For now, just log it
	c.logger.Debug("bid message received",
		zap.String("client_id", c.ID.String()),
		zap.Any("data", message.Data))
}

// handleAuthenticate handles client authentication
func (c *Client) handleAuthenticate(message Message) {
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		c.sendError("invalid message data")
		return
	}

	userIDStr, ok := data["user_id"].(string)
	if !ok {
		c.sendError("missing or invalid user_id")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.sendError("invalid user_id format")
		return
	}

	c.SetUserID(&userID)

	// Send confirmation
	c.Send <- Message{
		Type: "authenticated",
		Data: map[string]interface{}{
			"user_id":   userID.String(),
			"client_id": c.ID.String(),
		},
		Timestamp: time.Now(),
	}

	c.logger.Info("client authenticated",
		zap.String("client_id", c.ID.String()),
		zap.String("user_id", userID.String()))
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	select {
	case c.Send <- Message{
		Type: "error",
		Data: map[string]interface{}{
			"message": errorMsg,
		},
		Timestamp: time.Now(),
	}:
	default:
		// Channel is full, close the client
		close(c.Send)
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message Message) {
	select {
	case c.Send <- message:
	default:
		// Channel is full, close the client
		close(c.Send)
		c.Hub.Unregister <- c
	}
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	close(c.Send)
	c.Conn.Close()
}