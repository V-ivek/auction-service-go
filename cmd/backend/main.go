package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Adapters
	"auction-microservice/internal/adapters/clock"
	wsAdapter "auction-microservice/internal/adapters/websocket"

	// Domain
	"auction-microservice/internal/domain"

	// Handlers
	"auction-microservice/internal/handlers"

	// Services - import for mocks
	"auction-microservice/internal/services"

	// Package imports
	"auction-microservice/pkg/config"
	"auction-microservice/pkg/metrics"
	"auction-microservice/pkg/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Override database config for local testing without database
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.JWT.SecretKey = "development-secret-key-for-testing"
	cfg.Metrics.Port = 9091 // Use different port to avoid conflicts
	cfg.Server.Port = 8081  // Backend on port 8081

	// Initialize logger
	var logger *zap.Logger
	var err error
	if cfg.Logging.Format == "console" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting auction microservice BACKEND SERVER")

	// Initialize metrics
	var m *metrics.Metrics
	if cfg.Metrics.Enabled {
		m = metrics.New()
		logger.Info("metrics enabled", zap.Int("port", cfg.Metrics.Port))
	}

	// Initialize mock repositories for local testing
	userRepo := services.NewMockUserRepository()
	listingRepo := services.NewMockListingRepository()
	auctionRepo := services.NewMockAuctionRepository()
	bidRepo := services.NewMockBidRepository()
	subscriptionRepo := services.NewMockSubscriptionRepository()

	logger.Info("using in-memory repositories for local testing")

	// Add some sample data
	setupSampleData(userRepo, listingRepo, auctionRepo, logger)

	// Initialize adapters
	realClock := clock.NewRealClock()
	timeoutManager := clock.NewTimeoutManager(realClock, logger)

	// Initialize WebSocket hub and notifier
	hub := wsAdapter.NewHub(logger, subscriptionRepo)
	notifier := wsAdapter.NewNotifier(hub)

	// Initialize services
	listingService := services.NewListingService(listingRepo, userRepo, logger)
	auctionService := services.NewAuctionService(
		auctionRepo, listingRepo, bidRepo, notifier, timeoutManager, nil, logger,
	)
	biddingService := services.NewBiddingService(
		bidRepo, auctionRepo, listingRepo, notifier, timeoutManager, nil, logger,
	)
	subscriptionService := services.NewSubscriptionService(
		subscriptionRepo, listingRepo, userRepo, logger,
	)

	// Initialize middleware
	auth := middleware.NewJWTMiddleware(cfg, logger)

	// Initialize handlers
	httpHandlers := handlers.NewHTTPHandlers(listingService, auctionService, biddingService, logger)
	wsHandler := handlers.NewWebSocketHandler(
		hub, biddingService, subscriptionService, auctionService, auth, logger,
	)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("/health", httpHandlers.HealthCheck)

	// API endpoints
	mux.HandleFunc("/api/listings", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS for frontend
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		if r.Method == http.MethodGet {
			httpHandlers.GetListings(w, r)
		} else if r.Method == http.MethodPost {
			auth.Authenticate(http.HandlerFunc(httpHandlers.CreateListing)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/listings/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		httpHandlers.GetListing(w, r)
	})

	mux.HandleFunc("/api/auctions/active", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		httpHandlers.GetActiveAuctions(w, r)
	})

	mux.HandleFunc("/api/auctions/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		if r.Method == http.MethodGet {
			httpHandlers.GetAuction(w, r)
		} else if r.Method == http.MethodPost {
			auth.Authenticate(http.HandlerFunc(httpHandlers.CreateAuction)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Apply middleware chain (except for WebSocket)
	var handler http.Handler = mux
	handler = middleware.LoggingMiddleware(logger)(handler)
	if cfg.Metrics.Enabled && m != nil {
		handler = middleware.MetricsMiddleware(m)(handler)
	}

	// Create custom handler that bypasses middleware for WebSocket
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			// Handle WebSocket directly without middleware, with CORS
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			wsHandler.HandleWebSocket(w, r)
			return
		}
		// Use normal handler chain for other requests
		handler.ServeHTTP(w, r)
	})

	// Start metrics server if enabled
	if cfg.Metrics.Enabled {
		go func() {
			metricsServer := http.NewServeMux()
			metricsServer.Handle("/metrics", promhttp.Handler())

			logger.Info("starting metrics server", zap.Int("port", cfg.Metrics.Port))
			if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Metrics.Port), metricsServer); err != nil {
				logger.Error("metrics server failed", zap.Error(err))
			}
		}()
	}

	// Start WebSocket hub
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Start metrics updater
	if cfg.Metrics.Enabled && m != nil {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					m.SetActiveConnections(float64(hub.GetActiveConnections()))
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Start HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      customHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting BACKEND server",
			zap.String("host", cfg.Server.Host),
			zap.Int("port", cfg.Server.Port),
			zap.String("mode", "BACKEND-API"))

		logger.Info("backend endpoints available",
			zap.String("health", fmt.Sprintf("http://%s:%d/health", cfg.Server.Host, cfg.Server.Port)),
			zap.String("api", fmt.Sprintf("http://%s:%d/api", cfg.Server.Host, cfg.Server.Port)),
			zap.String("websocket", fmt.Sprintf("ws://%s:%d/ws", cfg.Server.Host, cfg.Server.Port)),
			zap.String("metrics", fmt.Sprintf("http://%s:%d/metrics", cfg.Server.Host, cfg.Metrics.Port)))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("backend server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down backend server...")

	// Cancel context to stop background goroutines
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("backend server shutdown failed", zap.Error(err))
	} else {
		logger.Info("backend server shutdown complete")
	}
}

func setupSampleData(userRepo *services.MockUserRepository, listingRepo *services.MockListingRepository, auctionRepo *services.MockAuctionRepository, logger *zap.Logger) {
	ctx := context.Background()

	// Create sample users
	alice := domain.NewUser("alice", "alice@example.com")
	bob := domain.NewUser("bob", "bob@example.com")
	charlie := domain.NewUser("charlie", "charlie@example.com")

	userRepo.Create(ctx, alice)
	userRepo.Create(ctx, bob)
	userRepo.Create(ctx, charlie)

	// Create sample listings
	watch := domain.NewListing("Vintage Watch", "A beautiful vintage watch from the 1960s", 10000, 15000, alice.ID)
	car := domain.NewListing("Classic Car", "Well-maintained 1967 Mustang", 2500000, 3000000, bob.ID)
	painting := domain.NewListing("Art Painting", "Original oil painting by local artist", 50000, 75000, charlie.ID)

	listingRepo.Create(ctx, watch)
	listingRepo.Create(ctx, car)
	listingRepo.Create(ctx, painting)

	// Create sample auctions
	watchAuction := domain.NewAuction(watch.ID, 24*time.Hour)
	watchAuction.Start()
	auctionRepo.Create(ctx, watchAuction)

	carAuction := domain.NewAuction(car.ID, 48*time.Hour)
	// Leave car auction in pending state
	auctionRepo.Create(ctx, carAuction)

	logger.Info("sample data created",
		zap.Int("users", 3),
		zap.Int("listings", 3),
		zap.Int("auctions", 2),
		zap.String("active_auction", watchAuction.ID.String()))
}