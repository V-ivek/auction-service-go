package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := 3000
	
	// Serve static files from web directory
	fs := http.FileServer(http.Dir("./web/"))
	
	// Create a custom handler to add CORS headers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		fs.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("🎨 Frontend server starting on http://localhost:%d\n", port)
		fmt.Printf("📡 Backend API should be running on http://localhost:8081\n")
		fmt.Printf("🔌 WebSocket endpoint: ws://localhost:8081/ws\n")
		fmt.Println("📂 Serving files from ./web/ directory")
		fmt.Println("---")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Frontend server failed to start:", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down frontend server...")
}