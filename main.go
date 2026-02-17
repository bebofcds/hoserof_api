/*
================================================================================
HOSEROF_API - Main Application Entry Point
================================================================================

Description:
This file is the entry point for this project. It handles
service initialization, HTTP routing, JWT setup, and graceful shutdown.

Responsibilities:
1. Load environment variables from a `.env` file.
2. Initialize external services:
   - Firebase (database)
   - Supabase (storage bucket)
   - JWT authentication
3. Configure HTTP routes via `routes.SetupRouter`.
4. Start the HTTP server with proper read/write/idle timeouts.
5. Handle graceful shutdown on OS signals (SIGINT, SIGTERM).

Environment Variables:
- JWT_KEY          : Secret key for JWT authentication (required)
- PORT             : Server port (optional, defaults to 8080)
- Firebase & Supabase credentials must be set in environment or `.env`.

Developer Notes:
- All services are wrapped in the `config.Services` struct and injected into
  routes and middleware for dependency management.
- To add a new route, modify `routes.SetupRouter` and inject `services`.
- Adjust server timeouts as needed in the `http.Server` struct.
- Any long-running goroutines should observe `ctx.Done()` to support
  graceful shutdown.
- `godotenv` is used for local development; in production, env variables
  should be managed by your deployment platform.

================================================================================
*/

package main

import (
	"HOSEROF_API/config"
	"HOSEROF_API/routes"
	"HOSEROF_API/services"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file (if exists)
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
	// Create a context that listens for system interrupts
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// Ensure context is canceled on exit
	defer stop()

	// Initialize Firebase service
	firebase, err := config.NewFirebaseInstance(ctx)
	if err != nil {
		log.Fatalf("failed to initialize firebase: %v", err)
	}
	// Ensure Firebase resources are cleaned up
	defer firebase.Close()

	// Initialize Supabase service
	supabase, err := config.NewSupabaseInstance()
	if err != nil {
		log.Fatalf("failed to initialize supabase: %v", err)
	}

	// Read JWT secret from environment variables
	jwtSecret := os.Getenv("JWT_KEY")
	if jwtSecret == "" {
		// Terminate if secret is missing
		log.Fatal("JWT_KEY is required")
	}

	// Wrap all services in a single struct for dependency injection
	services := &config.Services{
		Firebase:  firebase,
		Supabase:  supabase,
		JWTSecret: []byte(jwtSecret),
		JWT:       services.Init([]byte(jwtSecret)),
	}

	router := routes.SetupRouter(services)

	// Determine server port (default to 8080 if not specified)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure HTTP server with timeouts
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 8 * time.Second,
	}
	go func() {
		log.Printf("Server running on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("Server exited Successfully")
}
