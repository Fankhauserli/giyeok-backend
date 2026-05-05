package db

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/schema.sql
var initSQL string

// InitializeSchema executes the embedded initialization SQL script.
// It uses CREATE TABLE IF NOT EXISTS to ensure idempotency.
func InitializeSchema(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Initializing database schema...")
	
	_, err := pool.Exec(ctx, initSQL)
	if err != nil {
		return fmt.Errorf("failed to execute init SQL: %w", err)
	}
	
	log.Println("Database schema initialized successfully.")
	return nil
}
