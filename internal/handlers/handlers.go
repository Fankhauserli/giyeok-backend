package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Fankhauserli/giyeok-backend/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Queries *db.Queries
	Pool    *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		Queries: db.New(pool),
		Pool:    pool,
	}
}

func (h *Handler) RespondWithError(w http.ResponseWriter, code int, message string) {
	h.RespondWithJSON(w, code, map[string]string{"error": message})
}

func (h *Handler) RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(payload)
}
