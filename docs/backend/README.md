# Backend Documentation

## Overview

The auction microservice backend is built using Go following hexagonal architecture principles. It provides a production-quality, real-time auction and bidding system with WebSocket support, structured logging, and comprehensive metrics.

## Architecture

### Hexagonal Architecture (Ports & Adapters)

The application follows clean architecture principles with clear separation of concerns:

```
internal/
├── domain/          # Core business logic
├── ports/           # Interface definitions
├── adapters/        # External integrations
├── services/        # Business services
└── handlers/        # HTTP/WebSocket handlers
```

### Core Components

#### Domain Layer (`internal/domain/`)
- **Auction**: Core auction entity with lifecycle management
- **Bid**: Bid entity with validation rules
- **Listing**: Item listing entity
- **User**: User entity for authentication
- **Events**: Domain events for real-time notifications

#### Ports (`internal/ports/`)
- **Repository**: Data persistence interfaces
- **Notifier**: Real-time notification interface
- **Clock**: Time management interface
- **Services**: Business logic interfaces

#### Adapters (`internal/adapters/`)
- **Clock**: Real-time clock and timeout management
- **WebSocket**: Real-time communication hub
- **PostgreSQL**: Database adapters (for production)

#### Services (`internal/services/`)
- **AuctionService**: Manages auction lifecycle
- **BiddingService**: Handles bid placement and validation
- **ListingService**: Manages item listings
- **SubscriptionService**: Manages user subscriptions

## API Endpoints

### REST API

#### Health Check
```http
GET /health
```
Returns service health status.

#### Listings
```http
GET /api/listings                    # Get all listings
POST /api/listings                   # Create new listing (authenticated)
GET /api/listings/{id}               # Get specific listing
```

#### Auctions
```http
GET /api/auctions/active             # Get all active auctions
GET /api/auctions/{id}               # Get specific auction
POST /api/auctions/{id}              # Create new auction (authenticated)
```

### WebSocket API

#### Connection
```
ws://localhost:8081/ws?token={jwt_token}
```

#### Message Format
```json
{
  "action": "subscribe|bid|unsubscribe",
  "listing_id": "uuid",
  "amount": 100.00
}
```

#### Response Format
```json
{
  "event": "bid_placed|auction_started|auction_ended|error",
  "data": { ... }
}
```

## Configuration

The service uses environment-based configuration with sensible defaults:

### Environment Variables
```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8081
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Database Configuration (Production)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=auction_db
DB_USER=auction_user
DB_PASSWORD=secure_password

# JWT Configuration
JWT_SECRET_KEY=your-secret-key
JWT_EXPIRATION=24h

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9091
```

### Configuration Loading
Configuration is loaded via `pkg/config/config.go` with the following precedence:
1. Environment variables
2. Default values

## Database Schema

### Tables

#### users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### listings
```sql
CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    reserve_price INTEGER NOT NULL, -- in cents
    buy_now_price INTEGER,
    seller_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### auctions
```sql
CREATE TABLE auctions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID NOT NULL REFERENCES listings(id),
    status VARCHAR(50) DEFAULT 'pending',
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    duration BIGINT NOT NULL, -- in nanoseconds
    current_bid INTEGER DEFAULT 0,
    bid_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_auctions_status ON auctions(status);
CREATE INDEX idx_auctions_end_time ON auctions(end_time);
```

#### bids
```sql
CREATE TABLE bids (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id UUID NOT NULL REFERENCES auctions(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount INTEGER NOT NULL, -- in cents
    placed_at TIMESTAMP DEFAULT NOW(),
    is_winning BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_bids_auction_amount ON bids(auction_id, amount DESC);
CREATE INDEX idx_bids_user ON bids(user_id);
```

#### subscriptions
```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    listing_id UUID NOT NULL REFERENCES listings(id),
    subscribed_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, listing_id)
);
```

## Business Rules

### Auction Rules
1. **Reserve Price**: All listings have a minimum starting bid (reserve price)
2. **Bid Increment**: Each bid must exceed the current highest bid by at least $1
3. **Time Extension**: Bids in the last 30 seconds extend the auction by 30 seconds (anti-sniping)
4. **Authentication Required**: All bid placement requires valid JWT authentication
5. **Persistence Guarantees**: Bids are only broadcast after successful database write

### Auction Lifecycle
1. **Created**: Auction is created but not started
2. **Active**: Auction is running and accepting bids
3. **Extended**: Auction time extended due to late bid
4. **Ended**: Auction completed, winner determined

## Concurrency Model

### Goroutine Management
- Each active auction runs in its own goroutine
- WebSocket hub manages client connections in separate goroutines
- Context cancellation for graceful shutdown
- Channel-based communication between components

### Race Condition Prevention
- Mutex locks for shared state
- Atomic operations for counters
- Database transactions for consistency
- Proper error handling and rollback

## Error Handling

### Error Types
```go
type DomainError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`
}
```

### HTTP Response Format
```json
{
  "success": false,
  "error": {
    "code": "INVALID_BID_AMOUNT",
    "message": "Bid amount must be greater than current bid",
    "field": "amount"
  }
}
```

## Logging

Uses structured logging with Uber's Zap library:

### Log Levels
- **DEBUG**: Detailed debugging information
- **INFO**: General information about service operation
- **WARN**: Warning conditions
- **ERROR**: Error conditions
- **FATAL**: Critical errors causing service shutdown

### Log Fields
- Timestamp
- Level
- Message
- Request ID (for tracing)
- User ID (when available)
- Auction/Listing ID (when relevant)

### Example Log Entry
```json
{
  "level": "info",
  "ts": 1640995200.123456,
  "caller": "handlers/websocket.go:45",
  "msg": "new bid placed",
  "auction_id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "alice",
  "amount": 150.00,
  "request_id": "req_123"
}
```

## Metrics

### Prometheus Metrics

#### Counter Metrics
- `auction_bids_total`: Total number of bids placed
- `auction_auctions_started_total`: Total auctions started
- `auction_auctions_completed_total`: Total auctions completed
- `auction_websocket_connections_total`: Total WebSocket connections

#### Gauge Metrics
- `auction_active_connections`: Current active WebSocket connections
- `auction_active_auctions`: Current number of active auctions

#### Histogram Metrics
- `auction_bid_processing_duration_seconds`: Time to process bid requests
- `auction_websocket_message_duration_seconds`: WebSocket message processing time

### Metrics Endpoint
```http
GET /metrics
```
Exposes Prometheus-formatted metrics on port 9091.

## Testing

### Unit Tests
```bash
go test ./internal/domain/...
go test ./internal/services/...
```

### Integration Tests
```bash
go test ./internal/adapters/...
go test ./internal/handlers/...
```

### Test Coverage
```bash
go test -cover ./...
```

Target: 100% coverage for domain logic, >80% for adapters.

## Development Setup

### Prerequisites
- Go 1.21+
- PostgreSQL 13+ (for production)
- Docker & Docker Compose (optional)

### Local Development
```bash
# Clone repository
git clone <repository-url>
cd auction-microservice

# Install dependencies
go mod download

# Run with in-memory storage (development)
go run cmd/backend/main.go

# Run tests
go test ./...

# Run with race detection
go test -race ./...
```

### Docker Development
```bash
# Build and run with Docker Compose
docker-compose up --build

# Run tests in container
docker-compose exec backend go test ./...
```

## Performance Considerations

### Database Optimization
- Proper indexing on frequently queried columns
- Connection pooling for database connections
- Prepared statements to prevent SQL injection
- Query optimization for high-traffic endpoints

### Memory Management
- Connection limits for WebSocket clients
- Graceful handling of disconnected clients
- Proper cleanup of goroutines and channels

### Scalability
- Stateless design allows horizontal scaling
- Database as single source of truth
- Redis can be added for session storage and caching

## Security

### Authentication
- JWT-based authentication
- Token expiration and refresh
- Secure token transmission

### Input Validation
- All inputs validated against business rules
- SQL injection prevention with prepared statements
- XSS prevention in WebSocket messages

### Rate Limiting
- Per-user rate limiting for bid placement
- Connection limits for WebSocket clients
- API rate limiting (can be added)

## Monitoring

### Health Checks
- `/health` endpoint for load balancer health checks
- Database connectivity check
- Service startup validation

### Observability
- Structured logging for debugging
- Metrics for performance monitoring
- Tracing can be added with OpenTelemetry

## Production Considerations

### Deployment
- Docker containerization
- Kubernetes deployment manifests
- Database migrations
- Configuration management
- Secrets management

### Reliability
- Graceful shutdown handling
- Circuit breaker patterns for external services
- Retry logic with exponential backoff
- Dead letter queues for failed operations

### Performance
- Connection pooling
- Caching strategies
- Background job processing
- Database query optimization