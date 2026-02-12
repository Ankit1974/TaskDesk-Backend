package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB(connString string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Unable to parse DATABASE_URL: %v", err))
	}

	Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Unable to create connection pool: %v", err))
	}

	// Ping to verify connection
	if err := Pool.Ping(ctx); err != nil {
		logger.Log.Fatal(fmt.Sprintf("Unable to connect to database: %v", err))
	}

	logger.Log.Info("Successfully connected to Supabase Database!")
}

func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}
