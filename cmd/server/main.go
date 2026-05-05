package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Fankhauserli/giyeok-backend/internal/db"
	"github.com/Fankhauserli/giyeok-backend/internal/handlers"
	"github.com/Fankhauserli/giyeok-backend/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/giyeok?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	// Initialize database schema if it doesn't exist
	if err := db.InitializeSchema(ctx, pool); err != nil {
		log.Fatalf("Unable to initialize database schema: %v\n", err)
	}

	h := handlers.NewHandler(pool)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Auth routes (unprotected)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/logout", h.Logout)

	// Protected routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/words", h.ListWords)
	protectedMux.HandleFunc("POST /api/words", h.CreateSingleWord)
	protectedMux.HandleFunc("POST /api/words/import", h.ImportWords)
	protectedMux.HandleFunc("GET /api/library", h.ListLibrary)
	protectedMux.HandleFunc("GET /api/study/due", h.ListDueReviews)
	protectedMux.HandleFunc("POST /api/study/review/{word_id}", h.ReviewWord)
	protectedMux.HandleFunc("GET /api/study/proficiency", h.GetProficiency)
	protectedMux.HandleFunc("GET /api/user/profile", h.GetProfile)
	protectedMux.HandleFunc("POST /api/user/settings", h.UpdateSettings)

	// Apply middleware to protected routes
	authHandler := middleware.AuthMiddleware(protectedMux)
	mux.Handle("/api/words/", authHandler)
	mux.Handle("/api/words", authHandler)
	mux.Handle("/api/study/", authHandler)
	mux.Handle("/api/user/", authHandler)
	mux.Handle("/api/library", authHandler)
	mux.Handle("/api/library/", authHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v\n", err)
	}
}
