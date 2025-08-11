# Railway Deployment Guide

Railway is a modern deployment platform that makes it easy to deploy web applications and databases. This guide covers deploying both the backend and frontend components of the auction microservice to Railway.

## Prerequisites

### 1. Railway Account Setup
- Create account at [railway.app](https://railway.app)
- Install Railway CLI:
```bash
# macOS
brew install railway

# Windows
npm install -g @railway/cli

# Linux
curl -fsSL https://railway.app/install.sh | sh
```

### 2. Authentication
```bash
# Login to Railway
railway login

# Verify authentication
railway whoami
```

## Backend Deployment

### 1. Project Setup

#### Create New Railway Project
```bash
# In your project root directory
railway init

# Or link to existing project
railway link [PROJECT_ID]
```

#### Alternative: Deploy from GitHub
1. Connect GitHub account to Railway
2. Import repository from GitHub
3. Railway will auto-detect Go application

### 2. Environment Configuration

#### Production Environment Variables
```bash
# Set environment variables via CLI
railway variables set SERVER_HOST=0.0.0.0
railway variables set SERVER_PORT=8080
railway variables set JWT_SECRET_KEY=your-super-secure-production-secret-key
railway variables set LOG_LEVEL=info
railway variables set LOG_FORMAT=json
railway variables set METRICS_ENABLED=true
railway variables set METRICS_PORT=9091

# Database will be auto-configured when you add PostgreSQL service
```

#### Or use Railway Dashboard
1. Go to your project dashboard
2. Click on "Variables" tab
3. Add environment variables:
   - `SERVER_HOST=0.0.0.0`
   - `SERVER_PORT=8080`
   - `JWT_SECRET_KEY=your-secure-key`
   - `LOG_LEVEL=info`
   - `METRICS_ENABLED=true`

### 3. Database Setup

#### Add PostgreSQL Service
```bash
# Add PostgreSQL database
railway add postgresql

# Railway automatically sets these environment variables:
# - DATABASE_URL
# - PGHOST
# - PGPORT
# - PGDATABASE
# - PGUSER
# - PGPASSWORD
```

#### Database Migration
Create a migration script `scripts/migrate.sh`:
```bash
#!/bin/bash
set -e

# Wait for database to be ready
until pg_isready -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE; do
  echo "Waiting for database to be ready..."
  sleep 2
done

# Run migrations
psql $DATABASE_URL < database/migrations/001_initial_schema.sql
psql $DATABASE_URL < database/migrations/002_add_indexes.sql

echo "Database migration completed"
```

#### Update Backend Code for Railway
Modify `cmd/backend/main.go` to use Railway's environment variables:

```go
func loadDatabaseConfig() *config.DatabaseConfig {
    // Railway provides DATABASE_URL
    if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
        return parseRailwayDatabaseURL(dbURL)
    }
    
    // Fallback to individual variables
    return &config.DatabaseConfig{
        Host:     getEnv("PGHOST", "localhost"),
        Port:     getEnvAsInt("PGPORT", 5432),
        Name:     getEnv("PGDATABASE", "auction_db"),
        User:     getEnv("PGUSER", "postgres"),
        Password: getEnv("PGPASSWORD", ""),
        SSLMode:  getEnv("DB_SSL_MODE", "require"),
    }
}
```

### 4. Dockerfile for Railway

Create `Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o backend cmd/backend/main.go

# Production stage
FROM alpine:3.18

# Install certificates and postgresql-client for migrations
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

# Copy the binary
COPY --from=builder /app/backend .
COPY --from=builder /app/database ./database
COPY --from=builder /app/scripts ./scripts

# Make scripts executable
RUN chmod +x scripts/*.sh

# Expose port
EXPOSE 8080

# Run the application
CMD ["./backend"]
```

### 5. Railway Configuration

Create `railway.json`:
```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "dockerfile"
  },
  "deploy": {
    "restartPolicyType": "always",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 100
  }
}
```

### 6. Deploy Backend
```bash
# Deploy the backend service
railway up

# Or deploy specific service
railway up --service backend

# Check deployment status
railway status

# View logs
railway logs

# Open deployed application
railway open
```

### 7. Domain Setup (Optional)
```bash
# Add custom domain
railway domain add yourdomain.com

# Or use Railway provided domain
railway domain
```

## Frontend Deployment

### 1. Frontend Service Setup

#### Option A: Separate Frontend Service
```bash
# Create new service for frontend
railway service new frontend

# Switch to frontend service
railway service use frontend
```

#### Option B: Static Site Deployment
Railway can also serve static files directly.

### 2. Frontend Dockerfile

Create `Dockerfile.frontend`:
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build frontend server
RUN CGO_ENABLED=0 GOOS=linux go build -o frontend cmd/frontend/main.go

# Production stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and web files
COPY --from=builder /app/frontend .
COPY --from=builder /app/web ./web

# Expose port
EXPOSE 3000

# Run frontend server
CMD ["./frontend"]
```

### 3. Frontend Environment Variables
```bash
# Set frontend environment variables
railway variables set PORT=3000
railway variables set BACKEND_URL=https://your-backend-url.railway.app
```

Update frontend configuration:
```go
// cmd/frontend/main.go
func main() {
    port := getEnv("PORT", "3000")
    backendURL := getEnv("BACKEND_URL", "http://localhost:8081")
    
    // Update CORS and proxy settings
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        // Inject backend URL into JavaScript
        if strings.HasSuffix(r.URL.Path, ".js") {
            content := injectBackendURL(r.URL.Path, backendURL)
            w.Header().Set("Content-Type", "application/javascript")
            w.Write(content)
            return
        }
        
        fs.ServeHTTP(w, r)
    })
    
    server := &http.Server{
        Addr:    fmt.Sprintf(":%s", port),
        Handler: handler,
    }
    
    log.Printf("Frontend server starting on port %s", port)
    log.Printf("Backend URL: %s", backendURL)
    
    if err := server.ListenAndServe(); err != nil {
        log.Fatal("Frontend server failed:", err)
    }
}
```

### 4. Deploy Frontend
```bash
# Deploy frontend service
railway up --dockerfile Dockerfile.frontend

# Or if using separate service
railway service use frontend
railway up
```

## Production Configuration

### 1. Environment Variables Summary

#### Backend Service
```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database (auto-configured by Railway)
DATABASE_URL=postgresql://user:pass@host:port/db
PGHOST=hostname
PGPORT=5432
PGDATABASE=dbname
PGUSER=username
PGPASSWORD=password

# JWT
JWT_SECRET_KEY=your-super-secure-production-key
JWT_EXPIRATION=24h

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9091

# CORS
CORS_ALLOWED_ORIGINS=https://your-frontend-domain.railway.app,https://yourdomain.com
```

#### Frontend Service
```env
PORT=3000
BACKEND_URL=https://your-backend-service.railway.app
```

### 2. Database Migration Strategy

Create `database/migrations/` directory:
```
database/
  migrations/
    001_initial_schema.sql
    002_add_indexes.sql
    003_add_constraints.sql
```

Example migration script:
```sql
-- 001_initial_schema.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Continue with other tables...
```

### 3. Health Checks

Implement comprehensive health checks:
```go
func (h *HTTPHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "service": "auction-microservice",
        "status":  "healthy",
        "timestamp": time.Now(),
        "version": os.Getenv("RAILWAY_GIT_COMMIT_SHA"),
        "environment": os.Getenv("RAILWAY_ENVIRONMENT"),
    }
    
    // Check database connectivity
    if err := h.db.PingContext(r.Context()); err != nil {
        health["status"] = "unhealthy"
        health["database"] = "disconnected"
        w.WriteHeader(http.StatusServiceUnavailable)
    } else {
        health["database"] = "connected"
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data":    health,
    })
}
```

## Monitoring and Observability

### 1. Railway Metrics
Railway provides built-in monitoring:
- CPU usage
- Memory usage
- Network traffic
- Request metrics
- Error rates

### 2. Custom Metrics Endpoint
```go
// Expose metrics for external monitoring
func setupMetrics() {
    http.Handle("/metrics", promhttp.Handler())
    
    // Custom metrics
    auctionBidsTotal := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auction_bids_total",
            Help: "Total number of bids placed",
        },
        []string{"auction_id", "status"},
    )
    prometheus.MustRegister(auctionBidsTotal)
}
```

### 3. Logging Configuration
```go
func setupLogging() *zap.Logger {
    config := zap.NewProductionConfig()
    
    // Railway log format
    config.Encoding = "json"
    config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    
    // Add Railway-specific fields
    config.InitialFields = map[string]interface{}{
        "service":     "auction-backend",
        "version":     os.Getenv("RAILWAY_GIT_COMMIT_SHA"),
        "environment": os.Getenv("RAILWAY_ENVIRONMENT"),
    }
    
    logger, _ := config.Build()
    return logger
}
```

## Scaling and Performance

### 1. Horizontal Scaling
```bash
# Scale backend service
railway service scale --replicas 3

# Set resource limits
railway service resources --memory 1GB --cpu 1000m
```

### 2. Database Connection Pooling
```go
func setupDatabase() *sql.DB {
    config := pgxpool.Config{
        MaxConns:        30,
        MinConns:        5,
        MaxConnLifetime: time.Hour,
        MaxConnIdleTime: time.Minute * 30,
    }
    
    db, err := pgxpool.ConnectConfig(context.Background(), &config)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    return stdlib.OpenDB(*db.Config().ConnConfig)
}
```

## Security Considerations

### 1. Environment Variables
```bash
# Use Railway's secure variable storage
railway variables set JWT_SECRET_KEY --sensitive
railway variables set DATABASE_PASSWORD --sensitive
```

### 2. CORS Configuration
```go
func setupCORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            for _, allowed := range allowedOrigins {
                if origin == allowed {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    break
                }
            }
            
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Deployment Commands Reference

### Basic Commands
```bash
# Initial setup
railway login
railway init
railway link [PROJECT_ID]

# Service management
railway service
railway service new [NAME]
railway service delete [NAME]

# Deployment
railway up
railway up --service [NAME]
railway up --dockerfile [PATH]

# Environment variables
railway variables
railway variables set KEY=value
railway variables set KEY=value --sensitive
railway variables delete KEY

# Database
railway add postgresql
railway add redis
railway add mysql

# Monitoring
railway logs
railway logs --follow
railway status
railway ps

# Domain management
railway domain
railway domain add yourdomain.com
railway domain delete yourdomain.com

# Project management
railway open
railway whoami
railway logout
```

### Advanced Commands
```bash
# Connect to database
railway connect postgresql

# Run commands in Railway environment
railway run npm install
railway run go test ./...

# Local development with Railway variables
railway run --local go run cmd/backend/main.go

# Deploy specific branch
railway up --branch production

# Rollback deployment
railway rollback

# Export environment variables
railway variables --json > variables.json
```

## Troubleshooting

### Common Issues

#### 1. Build Failures
```bash
# Check build logs
railway logs --deployment

# Verify Dockerfile
railway build --dockerfile Dockerfile

# Check Railway build context
railway up --verbose
```

#### 2. Database Connection Issues
```bash
# Test database connection
railway connect postgresql

# Check database variables
railway variables | grep -i db

# Verify migration status
railway run psql $DATABASE_URL -c "SELECT * FROM schema_migrations;"
```

#### 3. Environment Variable Issues
```bash
# List all variables
railway variables

# Test variable in Railway environment
railway run printenv | grep JWT_SECRET_KEY

# Update variables
railway variables set JWT_SECRET_KEY=new-value
```

#### 4. CORS Issues
Check browser console and update CORS configuration:
```go
w.Header().Set("Access-Control-Allow-Origin", "https://your-frontend.railway.app")
```

#### 5. Health Check Failures
```bash
# Test health endpoint
curl https://your-app.railway.app/health

# Check Railway health check configuration
railway service info
```

### Getting Help
- Railway Documentation: [docs.railway.app](https://docs.railway.app)
- Railway Discord: [discord.gg/railway](https://discord.gg/railway)
- Railway Status: [status.railway.app](https://status.railway.app)

## Cost Optimization

### 1. Resource Management
- Set appropriate memory and CPU limits
- Use autoscaling for variable workloads
- Monitor usage in Railway dashboard

### 2. Database Optimization
- Use connection pooling
- Implement query optimization
- Consider read replicas for high traffic

### 3. Monitoring Costs
- Check Railway usage dashboard
- Set up billing alerts
- Monitor resource utilization

Railway provides a generous free tier and competitive pricing for production workloads, making it an excellent choice for deploying the auction microservice.