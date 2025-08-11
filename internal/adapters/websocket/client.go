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
	maxMessageSize = 2048 // Increased from 512 to 2048 bytes
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
    // Accept any JSON type for inbound timestamps (number or string) to avoid
    // strict decoding errors from clients that send epoch milliseconds.
    // When sending messages, we continue providing time.Time which will marshal
    // to RFC3339 strings for clients.
    Timestamp interface{} `json:"timestamp,omitempty"`
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
				// Log ALL errors, not just unexpected ones
				c.logger.Error("websocket read error", 
					zap.String("client_id", c.ID.String()),
					zap.Error(err),
					zap.String("error_type", "ReadJSON"))
				return
			}

			c.logger.Info("received websocket message",
				zap.String("client_id", c.ID.String()),
				zap.String("message_type", message.Type),
				zap.Any("message_data", message.Data))

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
	c.logger.Info("handling websocket message",
		zap.String("client_id", c.ID.String()),
		zap.String("message_type", message.Type))
		
	switch message.Type {
  case "ping":
      // Respond to client heartbeats
      c.Send <- Message{
          Type:      "pong",
          Data:      map[string]interface{}{},
          Timestamp: time.Now(),
      }
	case "join_auction":
		c.handleJoinAuction(message)
	case "leave_auction":
		c.handleLeaveAuction(message)
	case "place_bid":
		c.handlePlaceBid(message)
  case "subscribe", "unsubscribe":
      // Forward subscription messages to hub for processing by the WebSocket handler
      c.Hub.ProcessClientMessage <- &ClientMessage{
          Client:  c,
          Message: message,
      }
	case "authenticate":
		c.handleAuthenticate(message)
	default:
		c.logger.Warn("unknown message type",
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
	c.logger.Info("bid message received, forwarding to hub",
		zap.String("client_id", c.ID.String()),
		zap.Any("data", message.Data),
		zap.String("message_type", message.Type))
	
	// Forward to hub for processing by WebSocket handler
	select {
	case c.Hub.ProcessClientMessage <- &ClientMessage{
		Client:  c,
		Message: message,
	}:
		c.logger.Info("bid message successfully queued to hub",
			zap.String("client_id", c.ID.String()))
	default:
		c.logger.Error("failed to queue bid message to hub - channel full",
			zap.String("client_id", c.ID.String()))
		c.sendError("server busy, please try again")
	}
}

// handleAuthenticate handles client authentication
func (c *Client) handleAuthenticate(message Message) {
	c.logger.Info("processing authentication message",
		zap.String("client_id", c.ID.String()),
		zap.Any("message_data", message.Data))
		
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		c.logger.Warn("invalid message data format for authentication",
			zap.String("client_id", c.ID.String()))
		c.sendError("invalid message data")
		return
	}

	userIDStr, ok := data["user_id"].(string)
	if !ok {
		c.logger.Warn("missing or invalid user_id in authentication message",
			zap.String("client_id", c.ID.String()),
			zap.Any("user_id", data["user_id"]))
		c.sendError("missing or invalid user_id")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.logger.Warn("invalid user_id format",
			zap.String("client_id", c.ID.String()),
			zap.String("user_id_str", userIDStr),
			zap.Error(err))
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

	c.logger.Info("client authenticated successfully",
		zap.String("client_id", c.ID.String()),
		zap.String("user_id", userID.String()))
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	c.logger.Warn("sending error to client",
		zap.String("client_id", c.ID.String()),
		zap.String("error_message", errorMsg))
		
	select {
	case c.Send <- Message{
		Type: "error",
		Data: map[string]interface{}{
			"message": errorMsg,
		},
		Timestamp: time.Now(),
	}:
	default:
		// Channel is full, log but don't close connection immediately
		c.logger.Warn("client send channel full, message dropped",
			zap.String("client_id", c.ID.String()),
			zap.String("error_message", errorMsg))
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