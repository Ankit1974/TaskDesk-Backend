// Package db manages the PostgreSQL database connection pool.
// It uses pgx/v5 with connection pooling to efficiently handle concurrent
// database requests across the application.
//
// Usage: Access the pool from any package as db.Pool (e.g., db.Pool.QueryRow(...)).
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the global database connection pool.
// It manages multiple reusable connections to the Supabase PostgreSQL database.
// Must be initialized via InitDB() before use.
var Pool *pgxpool.Pool

// InitDB creates and validates a connection pool using the provided connection string.
// It will fatally exit if the database is unreachable (fail-fast on startup).
func InitDB(connString string) {
	// 10-second timeout for the initial connection attempt
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse the connection string into pool configuration
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Unable to parse DATABASE_URL: %v", err))
	}

	// Create the connection pool
	Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Unable to create connection pool: %v", err))
	}

	// Verify connectivity with a ping
	if err := Pool.Ping(ctx); err != nil {
		logger.Log.Fatal(fmt.Sprintf("Unable to connect to database: %v", err))
	}

	logger.Log.Info("Successfully connected to Supabase Database!")
}

// CloseDB gracefully shuts down the connection pool.
// Should be called via defer in main() to release all connections on server shutdown.
func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}
