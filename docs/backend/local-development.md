# Backend Local Development Guide

## Prerequisites

### Required Software
- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Git**: For version control
- **PostgreSQL 13+**: For production database (optional for local development)
- **Docker**: For containerized development (optional)

### Verify Installation
```bash
# Check Go version
go version

# Check Git version
git version

# Check PostgreSQL (if installed)
psql --version

# Check Docker (if using)
docker --version
docker-compose --version
```

## Quick Start

### 1. Clone Repository
```bash
git clone <repository-url>
cd auction-microservice
```

### 2. Install Dependencies
```bash
go mod download
go mod verify
```

### 3. Run Backend Server
```bash
# Run with in-memory storage (default for development)
go run cmd/backend/main.go
```

The server will start with the following services:
- **Backend API**: http://localhost:8081
- **WebSocket**: ws://localhost:8081/ws
- **Health Check**: http://localhost:8081/health
- **Metrics**: http://localhost:9091/metrics

### 4. Test the Setup
```bash
# Test health endpoint
curl http://localhost:8081/health

# Test API endpoint
curl http://localhost:8081/api/auctions/active

# Test metrics endpoint
curl http://localhost:9091/metrics
```

## Development Modes

### In-Memory Mode (Default)

Perfect for local development and testing:

```bash
go run cmd/backend/main.go
```

**Features:**
- No database setup required
- Pre-loaded sample data
- All data stored in memory
- Resets on restart

**Sample Data Includes:**
- 3 users: alice, bob, charlie
- 3 listings: Vintage Watch, Classic Car, Art Painting
- 2 auctions: 1 active, 1 pending

### Database Mode (Production-like)

For testing with persistent storage:

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=auction_db
export DB_USER=auction_user
export DB_PASSWORD=your_password

go run cmd/backend/main.go
```

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8081
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Database Configuration (optional for local development)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=auction_db
DB_USER=auction_user
DB_PASSWORD=secure_password
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=10

# JWT Configuration
JWT_SECRET_KEY=development-secret-key-do-not-use-in-production
JWT_EXPIRATION=24h

# Logging Configuration
LOG_LEVEL=debug
LOG_FORMAT=console

# Metrics Configuration
METRICS_ENABLED=true
METRICS_PORT=9091
```

### Loading Environment Variables
```bash
# Using direnv (recommended)
echo "source_env .env" > .envrc
direnv allow

# Or export manually
export $(cat .env | xargs)

# Or use go-dotenv
go get github.com/joho/godotenv
```

## Database Setup (Optional)

### Using PostgreSQL

#### 1. Install PostgreSQL
```bash
# macOS with Homebrew
brew install postgresql
brew services start postgresql

# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib
sudo systemctl start postgresql

# Windows
# Download from https://www.postgresql.org/download/windows/
```

#### 2. Create Database and User
```sql
-- Connect as postgres superuser
sudo -u postgres psql

-- Create user
CREATE USER auction_user WITH PASSWORD 'secure_password';

-- Create database
CREATE DATABASE auction_db OWNER auction_user;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE auction_db TO auction_user;

-- Exit
\q
```

#### 3. Create Tables
```sql
-- Connect to auction database
psql -h localhost -U auction_user -d auction_db

-- Run the schema creation scripts
-- (You would typically have migration files)
\i database/migrations/001_initial_schema.sql
```

#### 4. Run with Database
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=auction_db
export DB_USER=auction_user
export DB_PASSWORD=secure_password

go run cmd/backend/main.go
```

### Using Docker for Database

#### 1. Docker Compose Setup
Create `docker-compose.dev.yml`:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: auction_db
      POSTGRES_USER: auction_user
      POSTGRES_PASSWORD: secure_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U auction_user -d auction_db"]
      interval: 30s
      timeout: 10s
      retries: 5

volumes:
  postgres_data:
```

#### 2. Start Database
```bash
docker-compose -f docker-compose.dev.yml up -d postgres
```

#### 3. Wait for Database to be Ready
```bash
# Wait for health check to pass
docker-compose -f docker-compose.dev.yml ps

# Or check logs
docker-compose -f docker-compose.dev.yml logs postgres
```

#### 4. Run Backend
```bash
export DB_HOST=localhost
go run cmd/backend/main.go
```

## Development Workflow

### 1. Code Structure
```
cmd/
  backend/
    main.go           # Backend server entry point
  frontend/
    main.go           # Frontend server entry point
internal/
  domain/             # Business entities
  ports/              # Interface definitions
  adapters/           # External integrations
  services/           # Business logic
  handlers/           # HTTP/WebSocket handlers
pkg/
  config/             # Configuration management
  middleware/         # HTTP middleware
  metrics/            # Prometheus metrics
```

### 2. Adding New Features

#### Step 1: Define Domain Entity
```go
// internal/domain/new_entity.go
type NewEntity struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

func NewEntity(name string) *NewEntity {
    return &NewEntity{
        ID:        uuid.New(),
        Name:      name,
        CreatedAt: time.Now(),
    }
}
```

#### Step 2: Define Port Interface
```go
// internal/ports/new_entity_repository.go
type NewEntityRepository interface {
    Create(ctx context.Context, entity *domain.NewEntity) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.NewEntity, error)
    GetAll(ctx context.Context) ([]*domain.NewEntity, error)
}
```

#### Step 3: Implement Service
```go
// internal/services/new_entity_service.go
type NewEntityService struct {
    repo   ports.NewEntityRepository
    logger *zap.Logger
}

func NewNewEntityService(repo ports.NewEntityRepository, logger *zap.Logger) *NewEntityService {
    return &NewEntityService{
        repo:   repo,
        logger: logger,
    }
}

func (s *NewEntityService) Create(ctx context.Context, name string) (*domain.NewEntity, error) {
    entity := domain.NewEntity(name)
    
    if err := s.repo.Create(ctx, entity); err != nil {
        s.logger.Error("failed to create entity", zap.Error(err))
        return nil, err
    }
    
    return entity, nil
}
```

#### Step 4: Add HTTP Handler
```go
// internal/handlers/new_entity_handler.go
func (h *HTTPHandlers) CreateNewEntity(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name string `json:"name"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    entity, err := h.newEntityService.Create(r.Context(), req.Name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data":    entity,
    })
}
```

### 3. Testing

#### Unit Tests
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/domain
go test ./internal/services

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Integration Tests
```bash
# Run integration tests (requires database)
go test -tags=integration ./...

# Run with test database
TEST_DB_HOST=localhost TEST_DB_PORT=5432 go test -tags=integration ./...
```

#### Test Example
```go
// internal/services/auction_service_test.go
func TestAuctionService_CreateAuction(t *testing.T) {
    // Setup
    logger := zap.NewNop()
    mockRepo := &MockAuctionRepository{}
    service := NewAuctionService(mockRepo, logger)
    
    // Test data
    listingID := uuid.New()
    duration := 24 * time.Hour
    
    // Execute
    auction, err := service.CreateAuction(context.Background(), listingID, duration)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, auction)
    assert.Equal(t, listingID, auction.ListingID)
    assert.Equal(t, duration, auction.Duration)
}
```

### 4. Debugging

#### Enable Debug Logging
```bash
export LOG_LEVEL=debug
go run cmd/backend/main.go
```

#### Use Go Debugger (Delve)
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug cmd/backend/main.go

# Set breakpoints
(dlv) break main.main
(dlv) break internal/services/auction_service.go:45
(dlv) continue
```

#### IDE Debugging
Most IDEs support Go debugging:
- **VS Code**: Go extension with built-in debugger
- **GoLand**: Professional Go IDE with advanced debugging
- **Vim/Neovim**: vim-go plugin with debugging support

### 5. Code Quality

#### Linting
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

#### Formatting
```bash
# Format code
go fmt ./...

# Import organization
goimports -w .

# Use gofumpt for stricter formatting
go install mvdan.cc/gofumpt@latest
gofumpt -w .
```

#### Generate Code
```bash
# Generate mocks
go generate ./...

# Generate swagger docs (if using)
swag init -g cmd/backend/main.go
```

## Development Tools

### Recommended VS Code Extensions
- **Go**: Official Go extension
- **Go Outline**: Code navigation
- **REST Client**: API testing
- **Thunder Client**: API client
- **Docker**: Container management

### Useful Go Tools
```bash
# Code generation
go install github.com/golang/mock/mockgen@latest
go install github.com/swaggo/swag/cmd/swag@latest

# Performance profiling
go install github.com/google/pprof@latest

# Dependency analysis
go install github.com/kisielk/godepgraph@latest
```

### API Testing

#### Using cURL
```bash
# Health check
curl http://localhost:8081/health

# Get active auctions
curl http://localhost:8081/api/auctions/active

# Get listings
curl http://localhost:8081/api/listings

# Create listing (with auth)
curl -X POST http://localhost:8081/api/listings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"title":"Test Item","description":"Test description","reserve_price":10000}'
```

#### Using HTTPie
```bash
# Install HTTPie
pip install httpie

# Test endpoints
http GET localhost:8081/health
http GET localhost:8081/api/auctions/active
http POST localhost:8081/api/listings title="Test Item" reserve_price:=10000 Authorization:"Bearer YOUR_TOKEN"
```

## Common Issues and Solutions

### Port Already in Use
```bash
# Find process using port
lsof -i :8081

# Kill process
kill -9 PID
```

### Database Connection Issues
```bash
# Check PostgreSQL status
brew services list | grep postgresql
sudo systemctl status postgresql

# Test connection
psql -h localhost -U auction_user -d auction_db -c "SELECT 1;"
```

### Module Issues
```bash
# Clean module cache
go clean -modcache

# Tidy dependencies
go mod tidy

# Vendor dependencies
go mod vendor
```

### Memory Issues
```bash
# Check memory usage
go tool pprof http://localhost:8081/debug/pprof/heap

# Profile CPU usage
go tool pprof http://localhost:8081/debug/pprof/profile
```

## Performance Monitoring

### Built-in Profiling
Add pprof to your development server:

```go
import _ "net/http/pprof"

// Add pprof endpoints in debug mode
if cfg.Debug {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```

Access profiling at:
- http://localhost:6060/debug/pprof/
- http://localhost:6060/debug/pprof/heap
- http://localhost:6060/debug/pprof/profile

### Metrics Monitoring
View Prometheus metrics:
```bash
curl http://localhost:9091/metrics
```

## Next Steps

1. **Add Database Persistence**: Implement PostgreSQL adapters
2. **Add Authentication**: Implement JWT middleware
3. **Add Rate Limiting**: Prevent abuse
4. **Add Caching**: Implement Redis for performance
5. **Add Message Queues**: For reliable event processing
6. **Add Monitoring**: Integrate with Prometheus/Grafana
7. **Add Tracing**: Implement distributed tracing