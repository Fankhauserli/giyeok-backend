package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Fankhauserli/giyeok-backend/internal/auth"
	"github.com/Fankhauserli/giyeok-backend/internal/db"
	"github.com/Fankhauserli/giyeok-backend/internal/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	user, err := h.Queries.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		log.Printf("Error creating user: %v\n", err)
		h.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	h.RespondWithJSON(w, http.StatusCreated, user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := h.Queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		h.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		h.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := auth.GenerateJWT(uuid.UUID(user.ID.Bytes).String())
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Set to true in production
		Path:     "/",
	})

	h.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged in successfully"})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	h.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	uID, _ := uuid.Parse(userIDStr)
	userID := pgtype.UUID{Bytes: uID, Valid: true}

	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not fetch profile")
		return
	}

	h.RespondWithJSON(w, http.StatusOK, user)
}

type SettingsRequest struct {
	DailyNewLimit int     `json:"daily_new_limit"`
	RetentionGoal float64 `json:"retention_goal"`
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	uID, _ := uuid.Parse(userIDStr)
	userID := pgtype.UUID{Bytes: uID, Valid: true}

	var req SettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.DailyNewLimit < 0 || req.DailyNewLimit > 1000 {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid limit (0-1000)")
		return
	}
	
	if req.RetentionGoal < 0.5 || req.RetentionGoal > 0.99 {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid retention goal (0.5-0.99)")
		return
	}

	_, err := h.Queries.UpdateUserLimits(r.Context(), db.UpdateUserLimitsParams{
		ID:            userID,
		DailyNewLimit: int32(req.DailyNewLimit),
		RetentionGoal: req.RetentionGoal,
	})
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not update settings")
		return
	}

	h.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Settings updated"})
}
