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
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Initialize Logger
	logger.InitLogger(cfg.Env)
	defer logger.Log.Sync()

	logger.Log.Info("Starting TaskDesk Backend...")

	// 3. Initialize Database
	db.InitDB(cfg.DatabaseURL)
	defer db.CloseDB()

	// 4. Setup Router
	r := router.SetupRouter()

	// 4. Run Server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Log.Info(fmt.Sprintf("Server is running on %s", addr))
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
