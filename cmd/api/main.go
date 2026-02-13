// Package main is the entry point for the TaskDesk Backend API server.
// It initializes all core dependencies in the correct order and starts the HTTP server.
package main

import (
	"fmt"
	"log"

	"github.com/Ankit1974/TaskDeskBackend/internal/api/router"
	"github.com/Ankit1974/TaskDeskBackend/internal/config"
	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
)

func main() {
	// 1. Load Config — reads .env file and environment variables via Viper.
	//    Must run first since all other initializations depend on config values.
	cfg := config.LoadConfig()

	// 2. Initialize Logger — sets up Zap structured logging (development vs production mode).
	//    defer Sync() flushes any buffered log entries on shutdown.
	logger.InitLogger(cfg.Env)
	defer logger.Log.Sync()

	logger.Log.Info("Starting TaskDesk Backend...")

	// 3. Initialize Database — creates a PostgreSQL connection pool to the Supabase DB.
	//    defer CloseDB() ensures connections are released on shutdown.
	db.InitDB(cfg.DatabaseURL)
	defer db.CloseDB()

	// 4. Setup Router — registers all API routes, middleware chains, and handler functions.
	r := router.SetupRouter()

	// 5. Run Server — starts listening on the configured port (default: 8080).
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Log.Info(fmt.Sprintf("Server is running on %s", addr))
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
