package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fankhauserli/giyeok-backend/internal/db"
	"github.com/Fankhauserli/giyeok-backend/internal/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) ListWords(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	words, err := h.Queries.ListWords(r.Context(), db.ListWordsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not fetch words")
		return
	}
	if words == nil {
		words = []db.Word{}
	}

	h.RespondWithJSON(w, http.StatusOK, words)
}

type WordImport struct {
	Korean       string `json:"korean"`
	English      string `json:"english"`
	PartOfSpeech string `json:"part_of_speech"`
	TopikLevel   int    `json:"topik_level"`
}

func (h *Handler) ImportWords(w http.ResponseWriter, r *http.Request) {
	var imports []WordImport
	if err := json.NewDecoder(r.Body).Decode(&imports); err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	for _, imp := range imports {
		_, err := h.Queries.CreateWord(r.Context(), db.CreateWordParams{
			Korean:       imp.Korean,
			English:      imp.English,
			PartOfSpeech: pgtype.Text{String: imp.PartOfSpeech, Valid: imp.PartOfSpeech != ""},
			TopikLevel:   pgtype.Int4{Int32: int32(imp.TopikLevel), Valid: imp.TopikLevel > 0},
		})
		if err != nil {
			// In a real app, we might want to continue and collect errors
			h.RespondWithError(w, http.StatusInternalServerError, "Could not import some words")
			return
		}
	}

	h.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Words imported successfully"})
}

func (h *Handler) ListLibrary(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	uID, _ := uuid.Parse(userIDStr)
	userID := pgtype.UUID{Bytes: uID, Valid: true}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	words, err := h.Queries.ListLibraryWords(r.Context(), db.ListLibraryWordsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not fetch library")
		return
	}
	if words == nil {
		words = []db.ListLibraryWordsRow{}
	}

	h.RespondWithJSON(w, http.StatusOK, words)
}

func (h *Handler) CreateSingleWord(w http.ResponseWriter, r *http.Request) {
	var imp WordImport
	if err := json.NewDecoder(r.Body).Decode(&imp); err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	word, err := h.Queries.CreateWord(r.Context(), db.CreateWordParams{
		Korean:       imp.Korean,
		English:      imp.English,
		PartOfSpeech: pgtype.Text{String: imp.PartOfSpeech, Valid: imp.PartOfSpeech != ""},
		TopikLevel:   pgtype.Int4{Int32: int32(imp.TopikLevel), Valid: imp.TopikLevel > 0},
	})
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not create word")
		return
	}

	h.RespondWithJSON(w, http.StatusCreated, word)
}
