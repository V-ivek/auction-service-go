package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"auction-microservice/internal/domain"
	"auction-microservice/internal/ports"
)

// HTTPHandlers contains all HTTP handler methods
type HTTPHandlers struct {
	listingService ports.ListingService
	auctionService ports.AuctionService
	biddingService ports.BiddingService
	logger         *zap.Logger
}

// NewHTTPHandlers creates a new HTTPHandlers instance
func NewHTTPHandlers(
	listingService ports.ListingService,
	auctionService ports.AuctionService,
	biddingService ports.BiddingService,
	logger *zap.Logger,
) *HTTPHandlers {
	return &HTTPHandlers{
		listingService: listingService,
		auctionService: auctionService,
		biddingService: biddingService,
		logger:         logger,
	}
}

// CreateListingRequest represents the request body for creating a listing
type CreateListingRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	StartingBid  int64  `json:"starting_bid"`
	ReservePrice int64  `json:"reserve_price"`
}

// CreateAuctionRequest represents the request body for creating an auction
type CreateAuctionRequest struct {
	Duration string `json:"duration"` // e.g., "24h", "48h"
}

// PlaceBidRequest represents the request body for placing a bid
type PlaceBidRequest struct {
	Amount int64 `json:"amount"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// HealthCheck handles health check requests
func (h *HTTPHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "auction-microservice",
		},
	})
}

// CreateListing handles POST /api/listings
func (h *HTTPHandlers) CreateListing(w http.ResponseWriter, r *http.Request) {
	var req CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "user not authenticated", err)
		return
	}

	// Create listing
	listing, err := h.listingService.CreateListing(
		r.Context(),
		req.Title,
		req.Description,
		req.StartingBid,
		req.ReservePrice,
		userID,
	)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create listing", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    listing,
		Message: "listing created successfully",
	})
}

// GetListings handles GET /api/listings
func (h *HTTPHandlers) GetListings(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offset, limit := h.parsePaginationParams(r)

	listings, err := h.listingService.GetListings(r.Context(), offset, limit)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get listings", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"listings": listings,
			"offset":   offset,
			"limit":    limit,
			"count":    len(listings),
		},
	})
}

// GetListing handles GET /api/listings/{id}
func (h *HTTPHandlers) GetListing(w http.ResponseWriter, r *http.Request) {
	// Extract listing ID from URL path
	id, err := h.extractIDFromPath(r.URL.Path, "/api/listings/")
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid listing ID", err)
		return
	}

	listing, err := h.listingService.GetListing(r.Context(), id)
	if err != nil {
		if err == domain.ErrListingNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "listing not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get listing", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    listing,
	})
}

// CreateAuction handles POST /api/auctions/{listing_id}
func (h *HTTPHandlers) CreateAuction(w http.ResponseWriter, r *http.Request) {
	// Extract listing ID from URL path
	listingID, err := h.extractIDFromPath(r.URL.Path, "/api/auctions/")
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid listing ID", err)
		return
	}

	var req CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Parse duration
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid duration format", err)
		return
	}

	// Create auction
	auction, err := h.auctionService.CreateAuction(r.Context(), listingID, duration)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create auction", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    auction,
		Message: "auction created successfully",
	})
}

// GetAuction handles GET /api/auctions/{id}
func (h *HTTPHandlers) GetAuction(w http.ResponseWriter, r *http.Request) {
	// Extract auction ID from URL path
	id, err := h.extractIDFromPath(r.URL.Path, "/api/auctions/")
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid auction ID", err)
		return
	}

	auction, err := h.auctionService.GetAuction(r.Context(), id)
	if err != nil {
		if err == domain.ErrAuctionNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "auction not found", err)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get auction", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    auction,
	})
}

// GetActiveAuctions handles GET /api/auctions/active
func (h *HTTPHandlers) GetActiveAuctions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offset, limit := h.parsePaginationParams(r)

	auctions, err := h.auctionService.GetActiveAuctions(r.Context(), offset, limit)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get active auctions", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"auctions": auctions,
			"offset":   offset,
			"limit":    limit,
			"count":    len(auctions),
		},
	})
}

// PlaceBid handles POST /api/auctions/{auction_id}/bids
func (h *HTTPHandlers) PlaceBid(w http.ResponseWriter, r *http.Request) {
	// Extract auction ID from URL path
	auctionID, err := h.extractIDFromPath(r.URL.Path, "/api/auctions/")
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid auction ID", err)
		return
	}

	// Get user ID from context
	bidderID, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "user not authenticated", err)
		return
	}

	var req PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Place bid
	bid, err := h.biddingService.PlaceBid(r.Context(), auctionID, bidderID, req.Amount)
	if err != nil {
		switch err {
		case domain.ErrAuctionNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "auction not found", err)
		case domain.ErrAuctionNotActive:
			h.writeErrorResponse(w, http.StatusBadRequest, "auction is not active", err)
		case domain.ErrBidTooLow:
			h.writeErrorResponse(w, http.StatusBadRequest, "bid amount is too low", err)
		case domain.ErrCannotBidOnOwnItem:
			h.writeErrorResponse(w, http.StatusForbidden, "cannot bid on your own item", err)
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to place bid", err)
		}
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    bid,
		Message: "bid placed successfully",
	})
}

// Helper methods

// extractIDFromPath extracts UUID from URL path
func (h *HTTPHandlers) extractIDFromPath(path, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, fmt.Errorf("invalid path")
	}

	idStr := strings.TrimPrefix(path, prefix)
	// Remove any trailing path segments (e.g., /bids)
	if idx := strings.Index(idStr, "/"); idx != -1 {
		idStr = idStr[:idx]
	}

	return uuid.Parse(idStr)
}

// getUserIDFromContext extracts user ID from request context
func (h *HTTPHandlers) getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID := ctx.Value("user_id")
	if userID == nil {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	switch v := userID.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fmt.Errorf("invalid user ID type in context")
	}
}

// parsePaginationParams parses offset and limit from query parameters
func (h *HTTPHandlers) parsePaginationParams(r *http.Request) (int, int) {
	offset := 0
	limit := 50 // default limit

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	return offset, limit
}

// writeJSONResponse writes a JSON response
func (h *HTTPHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

// writeErrorResponse writes an error response
func (h *HTTPHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	h.logger.Error("HTTP error response",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.Error(err))

	errorResponse := ErrorResponse{
		Error:   err.Error(),
		Code:    statusCode,
		Message: message,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}