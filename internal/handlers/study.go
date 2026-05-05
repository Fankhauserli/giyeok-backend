package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/Fankhauserli/giyeok-backend/internal/db"
	"github.com/Fankhauserli/giyeok-backend/internal/fsrs"
	"github.com/Fankhauserli/giyeok-backend/internal/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) ListDueReviews(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	uID, _ := uuid.Parse(userIDStr)
	userID := pgtype.UUID{Bytes: uID, Valid: true}

	levelStr := r.URL.Query().Get("level")
	level := 0
	if l, err := strconv.Atoi(levelStr); err == nil {
		level = l
	}

	// Fetch due reviews
	dueReviews, err := h.Queries.ListDueReviews(r.Context(), db.ListDueReviewsParams{
		UserID:     userID,
		Due:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Limit:      50,
		TopikLevel: pgtype.Int4{Int32: int32(level), Valid: true},
	})
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not fetch due reviews")
		return
	}
	if dueReviews == nil {
		dueReviews = []db.ListDueReviewsRow{}
	}

	// Fetch user's custom limit
	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not fetch user settings")
		return
	}
	dailyNewLimit := user.DailyNewLimit
	
	// Start of today
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	newStartedToday, _ := h.Queries.CountNewWordsStartedToday(r.Context(), db.CountNewWordsStartedTodayParams{
		UserID:    userID,
		CreatedAt: pgtype.Timestamptz{Time: today, Valid: true},
	})

	canTakeNew := int32(dailyNewLimit) - int32(newStartedToday)
	if canTakeNew > 0 && len(dueReviews) < 20 {
		limit := canTakeNew
		if int32(20-len(dueReviews)) < limit {
			limit = int32(20 - len(dueReviews))
		}

		newWords, err := h.Queries.GetNewWordsForUser(r.Context(), db.GetNewWordsForUserParams{
			UserID:     userID,
			Limit:      limit,
			TopikLevel: pgtype.Int4{Int32: int32(level), Valid: true},
		})
		if err == nil {
			for _, nw := range newWords {
				dueReviews = append(dueReviews, db.ListDueReviewsRow{
					ID:           nw.ID,
					Korean:       nw.Korean,
					English:      nw.English,
					PartOfSpeech: nw.PartOfSpeech,
					TopikLevel:   nw.TopikLevel,
					State:        0, // New
					Due:          pgtype.Timestamptz{Time: time.Now(), Valid: true},
					CreatedAt:    nw.CreatedAt,
				})
			}
		}
	}

	// Shuffle the final list to prevent predictable patterns
	for i := range dueReviews {
		j := rand.Intn(i + 1)
		dueReviews[i], dueReviews[j] = dueReviews[j], dueReviews[i]
	}

	h.RespondWithJSON(w, http.StatusOK, dueReviews)
}

type ReviewRequest struct {
	Rating int `json:"rating"` // 1-4
}

func (h *Handler) GetProficiency(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	uID, _ := uuid.Parse(userIDStr)
	userID := pgtype.UUID{Bytes: uID, Valid: true}

	proficiency, err := h.Queries.GetLevelProficiency(r.Context(), userID)
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not fetch proficiency")
		return
	}
	if proficiency == nil {
		proficiency = []db.GetLevelProficiencyRow{}
	}

	// Fetch weekly activity
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	activity, _ := h.Queries.GetWeeklyActivity(r.Context(), db.GetWeeklyActivityParams{
		UserID:    userID,
		UpdatedAt: pgtype.Timestamptz{Time: oneWeekAgo, Valid: true},
	})

	// Calculate streak
	streak := 0
	today := now.Truncate(24 * time.Hour)
	current := today

	// activity is ordered by date DESC
	for _, a := range activity {
		studyDate := a.Time.Truncate(24 * time.Hour)
		if studyDate.Equal(current) {
			streak++
			current = current.AddDate(0, 0, -1)
		} else if studyDate.Before(current) {
			break
		}
	}

	// Fetch user's custom limit
	user, err := h.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not fetch user settings")
		return
	}

	h.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"levels": proficiency,
		"weekly_days": len(activity),
		"streak": streak,
		"daily_new_limit": user.DailyNewLimit,
		"retention_goal": user.RetentionGoal,
	})
}

func (h *Handler) ReviewWord(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	uID, _ := uuid.Parse(userIDStr)
	userID := pgtype.UUID{Bytes: uID, Valid: true}

	wordIDStr := r.PathValue("word_id")
	wID, err := uuid.Parse(wordIDStr)
	if err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid word ID")
		return
	}
	wordID := pgtype.UUID{Bytes: wID, Valid: true}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Fetch current review state or create new if not exists
	dbReview, err := h.Queries.GetReview(r.Context(), db.GetReviewParams{
		UserID: userID,
		WordID: wordID,
	})

	card := fsrs.Card{}
	if err == nil {
		card.State = int(dbReview.State)
		card.Due = dbReview.Due.Time
		card.Stability = dbReview.Stability
		card.Difficulty = dbReview.Difficulty
		card.ElapsedDays = int(dbReview.ElapsedDays)
		card.ScheduledDays = int(dbReview.ScheduledDays)
		card.Reps = int(dbReview.Reps)
		card.Lapses = int(dbReview.Lapses)
		if dbReview.LastReview.Valid {
			card.LastReview = dbReview.LastReview.Time
		}
	} else {
		// New card
		card.State = 0
		card.Due = time.Now()
	}

	// Fetch user's retention goal
	user, _ := h.Queries.GetUserByID(r.Context(), userID)
	retentionGoal := 0.90
	if user.RetentionGoal > 0 {
		retentionGoal = user.RetentionGoal
	}

	card.Review(fsrs.Rating(req.Rating), time.Now(), retentionGoal)

	// Upsert review back to DB
	_, err = h.Queries.UpsertReview(r.Context(), db.UpsertReviewParams{
		UserID:        userID,
		WordID:        wordID,
		State:         int32(card.State),
		Due:           pgtype.Timestamptz{Time: card.Due, Valid: true},
		Stability:     card.Stability,
		Difficulty:    card.Difficulty,
		ElapsedDays:   int32(card.ElapsedDays),
		ScheduledDays: int32(card.ScheduledDays),
		Reps:          int32(card.Reps),
		Lapses:        int32(card.Lapses),
		LastReview:    pgtype.Timestamptz{Time: card.LastReview, Valid: !card.LastReview.IsZero()},
	})

	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Could not save review")
		return
	}

	h.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Review recorded"})
}
