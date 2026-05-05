package db

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/schema.sql
var initSQL string

func InitializeSchema(ctx context.Context, pool *pgxpool.Pool) error {
	var dbName, user string
	err := pool.QueryRow(ctx, "SELECT current_database(), current_user").Scan(&dbName, &user)
	if err != nil {
		log.Printf("Could not query database identity: %v\n", err)
	} else {
		log.Printf("Connected to database: %s as user: %s\n", dbName, user)
	}

	log.Println("Starting database schema initialization...")

	// Split by semicolon and filter empty lines
	statements := strings.Split(initSQL, ";")

	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}

		log.Printf("Executing SQL statement: %s\n", strings.Split(trimmed, "\n")[0]) // Log first line of stmt
		_, err := pool.Exec(ctx, trimmed)
		if err != nil {
			// Some errors might be acceptable (e.g. extension already exists if not using IF NOT EXISTS)
			// but we use IF NOT EXISTS everywhere, so we should log and return error
			log.Printf("SQL execution error: %v\n", err)
			return fmt.Errorf("failed to execute statement [%s]: %w", trimmed, err)
		}
	}

	log.Println("Database schema initialization completed successfully.")
	return nil
}
