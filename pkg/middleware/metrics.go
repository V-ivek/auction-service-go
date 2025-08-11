package middleware

import (
	"net/http"
	"strconv"
	"time"

	"auction-microservice/pkg/metrics"
)

// MetricsMiddleware records HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code
			wrapped := newResponseWriter(w)

			// Process the request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Extract path without query parameters for metrics
			path := r.URL.Path
			method := r.Method
			statusCode := strconv.Itoa(wrapped.statusCode)

			// Record metrics
			m.RecordHTTPRequest(method, path, statusCode, duration)

			// Record additional metrics based on endpoints
			recordEndpointSpecificMetrics(m, path, method, wrapped.statusCode, r)
		})
	}
}

// recordEndpointSpecificMetrics records metrics specific to auction endpoints
func recordEndpointSpecificMetrics(m *metrics.Metrics, path, method string, statusCode int, r *http.Request) {
	// Record auction-specific metrics
	switch {
	case path == "/api/listings" && method == "POST" && statusCode == 201:
		m.RecordListing("created")
	case path == "/api/listings" && method == "GET":
		// Listing view - could track popularity
	case path == "/api/auctions/active" && method == "GET":
		// Active auctions view
	case path == "/api/auctions" && method == "POST" && statusCode == 201:
		m.RecordAuction("created")
	case path == "/ws":
		// WebSocket connection attempt
		if statusCode == 101 { // Switching Protocols
			m.RecordWebSocketConnection("connected")
		}
	}

	// Track specific auction or listing access
	if method == "GET" && statusCode == 200 {
		switch {
		case len(path) > 14 && path[:14] == "/api/listings/":
			// Individual listing access
		case len(path) > 14 && path[:14] == "/api/auctions/":
			// Individual auction access
		}
	}
}

// WebSocketMetricsWrapper wraps WebSocket operations with metrics
type WebSocketMetricsWrapper struct {
	metrics *metrics.Metrics
}

// NewWebSocketMetricsWrapper creates a new WebSocket metrics wrapper
func NewWebSocketMetricsWrapper(m *metrics.Metrics) *WebSocketMetricsWrapper {
	return &WebSocketMetricsWrapper{
		metrics: m,
	}
}

// RecordConnection records a WebSocket connection
func (w *WebSocketMetricsWrapper) RecordConnection() {
	w.metrics.IncrementWebSocketConnections()
}

// RecordDisconnection records a WebSocket disconnection
func (w *WebSocketMetricsWrapper) RecordDisconnection() {
	w.metrics.DecrementWebSocketConnections()
}

// RecordMessage records a WebSocket message
func (w *WebSocketMetricsWrapper) RecordMessage(messageType, direction string) {
	w.metrics.RecordWebSocketMessage(messageType, direction)
}

// RecordBid records a bid placement
func (w *WebSocketMetricsWrapper) RecordBid(successful bool, auctionID string, amount float64) {
	if successful {
		w.metrics.IncrementBidsPlaced(auctionID, amount)
	} else {
		w.metrics.IncrementBidsRejected()
	}
}

// AuctionMetricsWrapper wraps auction operations with metrics
type AuctionMetricsWrapper struct {
	metrics *metrics.Metrics
}

// NewAuctionMetricsWrapper creates a new auction metrics wrapper
func NewAuctionMetricsWrapper(m *metrics.Metrics) *AuctionMetricsWrapper {
	return &AuctionMetricsWrapper{
		metrics: m,
	}
}

// RecordAuctionCreated records an auction creation
func (w *AuctionMetricsWrapper) RecordAuctionCreated() {
	w.metrics.IncrementAuctionsCreated()
}

// RecordAuctionStarted records an auction start
func (w *AuctionMetricsWrapper) RecordAuctionStarted() {
	w.metrics.IncrementAuctionsStarted()
}

// RecordAuctionCompleted records an auction completion
func (w *AuctionMetricsWrapper) RecordAuctionCompleted(duration time.Duration) {
	w.metrics.IncrementAuctionsCompleted(duration.Seconds())
}

// RecordAuctionCancelled records an auction cancellation
func (w *AuctionMetricsWrapper) RecordAuctionCancelled(duration time.Duration) {
	w.metrics.IncrementAuctionsCancelled(duration.Seconds())
}

// UpdateTimeRemaining updates the time remaining for an auction
func (w *AuctionMetricsWrapper) UpdateTimeRemaining(auctionID string, remaining time.Duration) {
	w.metrics.SetAuctionTimeRemaining(auctionID, remaining.Seconds())
}

// RemoveTimeRemaining removes the time remaining metric for a completed auction
func (w *AuctionMetricsWrapper) RemoveTimeRemaining(auctionID string) {
	w.metrics.RemoveAuctionTimeRemaining(auctionID)
}

// DatabaseMetricsWrapper wraps database operations with metrics
type DatabaseMetricsWrapper struct {
	metrics *metrics.Metrics
}

// NewDatabaseMetricsWrapper creates a new database metrics wrapper
func NewDatabaseMetricsWrapper(m *metrics.Metrics) *DatabaseMetricsWrapper {
	return &DatabaseMetricsWrapper{
		metrics: m,
	}
}

// RecordQuery records a database query
func (w *DatabaseMetricsWrapper) RecordQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	w.metrics.RecordDatabaseQuery(operation, table, status, duration.Seconds())
}

// UpdateConnectionCount updates the database connection count
func (w *DatabaseMetricsWrapper) UpdateConnectionCount(count int) {
	w.metrics.SetDatabaseConnections(float64(count))
}

// HealthCheckMetrics records health check metrics
func HealthCheckMetrics(m *metrics.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Simple health check response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))

		// Record the health check
		duration := time.Since(start).Seconds()
		m.RecordHTTPRequest("GET", "/health", "200", duration)
	}
}