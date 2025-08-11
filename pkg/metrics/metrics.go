package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all metrics for the auction microservice
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPActiveConnections prometheus.Gauge

	// WebSocket metrics
	WebSocketConnections     prometheus.Gauge
	WebSocketMessages        *prometheus.CounterVec
	WebSocketConnectionsTotal *prometheus.CounterVec

	// Auction metrics
	AuctionsTotal          *prometheus.CounterVec
	ActiveAuctions         prometheus.Gauge
	AuctionDuration        *prometheus.HistogramVec
	BidsTotal              *prometheus.CounterVec
	BidAmount              *prometheus.HistogramVec

	// Business metrics
	ListingsTotal          *prometheus.CounterVec
	SubscriptionsTotal     *prometheus.CounterVec
	AuctionTimeRemaining   *prometheus.GaugeVec

	// System metrics
	DatabaseConnections    prometheus.Gauge
	DatabaseQueries        *prometheus.CounterVec
	DatabaseQueryDuration  *prometheus.HistogramVec
}

// New creates a new Metrics instance
func New() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),

		// WebSocket metrics
		WebSocketConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "websocket_active_connections",
				Help: "Number of active WebSocket connections",
			},
		),
		WebSocketMessages: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_messages_total",
				Help: "Total number of WebSocket messages",
			},
			[]string{"type", "direction"}, // direction: inbound/outbound
		),
		WebSocketConnectionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_connections_total",
				Help: "Total number of WebSocket connections",
			},
			[]string{"status"}, // status: connected/disconnected
		),

		// Auction metrics
		AuctionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auctions_total",
				Help: "Total number of auctions",
			},
			[]string{"status"}, // status: created/started/completed/cancelled
		),
		ActiveAuctions: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_auctions",
				Help: "Number of currently active auctions",
			},
		),
		AuctionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auction_duration_seconds",
				Help:    "Duration of completed auctions in seconds",
				Buckets: []float64{3600, 7200, 14400, 28800, 43200, 86400, 172800}, // 1h, 2h, 4h, 8h, 12h, 1d, 2d
			},
			[]string{"completion_reason"}, // completed/cancelled
		),
		BidsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "bids_total",
				Help: "Total number of bids placed",
			},
			[]string{"status"}, // status: successful/rejected
		),
		BidAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "bid_amount_cents",
				Help:    "Bid amounts in cents",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000}, // $1, $5, $10, $50, $100, $500, $1000, $5000, $10000
			},
			[]string{"auction_id"},
		),

		// Business metrics
		ListingsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "listings_total",
				Help: "Total number of listings",
			},
			[]string{"status"}, // status: created/active/sold/cancelled
		),
		SubscriptionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subscriptions_total",
				Help: "Total number of subscriptions",
			},
			[]string{"action"}, // action: subscribe/unsubscribe
		),
		AuctionTimeRemaining: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "auction_time_remaining_seconds",
				Help: "Time remaining for active auctions in seconds",
			},
			[]string{"auction_id"},
		),

		// System metrics
		DatabaseConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "database_connections_active",
				Help: "Number of active database connections",
			},
		),
		DatabaseQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"}, // operation: select/insert/update/delete, status: success/error
		),
		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"operation", "table"},
		),
	}
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// SetActiveConnections sets the number of active connections
func (m *Metrics) SetActiveConnections(count float64) {
	m.WebSocketConnections.Set(count)
}

// RecordWebSocketConnection records a WebSocket connection event
func (m *Metrics) RecordWebSocketConnection(status string) {
	m.WebSocketConnectionsTotal.WithLabelValues(status).Inc()
}

// RecordWebSocketMessage records a WebSocket message
func (m *Metrics) RecordWebSocketMessage(messageType, direction string) {
	m.WebSocketMessages.WithLabelValues(messageType, direction).Inc()
}

// RecordAuction records an auction event
func (m *Metrics) RecordAuction(status string) {
	m.AuctionsTotal.WithLabelValues(status).Inc()
}

// SetActiveAuctions sets the number of active auctions
func (m *Metrics) SetActiveAuctions(count float64) {
	m.ActiveAuctions.Set(count)
}

// RecordAuctionDuration records the duration of a completed auction
func (m *Metrics) RecordAuctionDuration(completionReason string, duration float64) {
	m.AuctionDuration.WithLabelValues(completionReason).Observe(duration)
}

// RecordBid records a bid event
func (m *Metrics) RecordBid(status string, auctionID string, amount float64) {
	m.BidsTotal.WithLabelValues(status).Inc()
	m.BidAmount.WithLabelValues(auctionID).Observe(amount)
}

// RecordListing records a listing event
func (m *Metrics) RecordListing(status string) {
	m.ListingsTotal.WithLabelValues(status).Inc()
}

// RecordSubscription records a subscription event
func (m *Metrics) RecordSubscription(action string) {
	m.SubscriptionsTotal.WithLabelValues(action).Inc()
}

// SetAuctionTimeRemaining sets the time remaining for an auction
func (m *Metrics) SetAuctionTimeRemaining(auctionID string, seconds float64) {
	m.AuctionTimeRemaining.WithLabelValues(auctionID).Set(seconds)
}

// RemoveAuctionTimeRemaining removes the time remaining metric for a completed auction
func (m *Metrics) RemoveAuctionTimeRemaining(auctionID string) {
	m.AuctionTimeRemaining.DeleteLabelValues(auctionID)
}

// SetDatabaseConnections sets the number of active database connections
func (m *Metrics) SetDatabaseConnections(count float64) {
	m.DatabaseConnections.Set(count)
}

// RecordDatabaseQuery records a database query
func (m *Metrics) RecordDatabaseQuery(operation, table, status string, duration float64) {
	m.DatabaseQueries.WithLabelValues(operation, table, status).Inc()
	m.DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// Helper methods for common operations

// IncrementHTTPRequests increments the HTTP request counter
func (m *Metrics) IncrementHTTPRequests(method, endpoint, statusCode string) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// ObserveHTTPDuration observes HTTP request duration
func (m *Metrics) ObserveHTTPDuration(method, endpoint string, duration float64) {
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// IncrementWebSocketConnections increments WebSocket connection counter
func (m *Metrics) IncrementWebSocketConnections() {
	m.WebSocketConnectionsTotal.WithLabelValues("connected").Inc()
}

// DecrementWebSocketConnections decrements WebSocket connection counter
func (m *Metrics) DecrementWebSocketConnections() {
	m.WebSocketConnectionsTotal.WithLabelValues("disconnected").Inc()
}

// IncrementBidsPlaced increments successful bids counter
func (m *Metrics) IncrementBidsPlaced(auctionID string, amount float64) {
	m.BidsTotal.WithLabelValues("successful").Inc()
	m.BidAmount.WithLabelValues(auctionID).Observe(amount)
}

// IncrementBidsRejected increments rejected bids counter
func (m *Metrics) IncrementBidsRejected() {
	m.BidsTotal.WithLabelValues("rejected").Inc()
}

// IncrementAuctionsCreated increments created auctions counter
func (m *Metrics) IncrementAuctionsCreated() {
	m.AuctionsTotal.WithLabelValues("created").Inc()
}

// IncrementAuctionsStarted increments started auctions counter
func (m *Metrics) IncrementAuctionsStarted() {
	m.AuctionsTotal.WithLabelValues("started").Inc()
}

// IncrementAuctionsCompleted increments completed auctions counter
func (m *Metrics) IncrementAuctionsCompleted(duration float64) {
	m.AuctionsTotal.WithLabelValues("completed").Inc()
	m.AuctionDuration.WithLabelValues("completed").Observe(duration)
}

// IncrementAuctionsCancelled increments cancelled auctions counter
func (m *Metrics) IncrementAuctionsCancelled(duration float64) {
	m.AuctionsTotal.WithLabelValues("cancelled").Inc()
	m.AuctionDuration.WithLabelValues("cancelled").Observe(duration)
}