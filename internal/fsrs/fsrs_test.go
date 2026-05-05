package fsrs

import (
	"testing"
	"time"
)

func TestNewCard(t *testing.T) {
	card := NewCard()
	if card.State != 0 {
		t.Errorf("Expected state 0, got %d", card.State)
	}
	if card.Stability != 0 {
		t.Errorf("Expected stability 0, got %f", card.Stability)
	}
}

func TestCard_Review(t *testing.T) {
	card := NewCard()
	now := time.Now()

	// Initial review (Good)
	card.Review(Good, now)

	if card.State != 1 {
		t.Errorf("Expected state 1 after initial review, got %d", card.State)
	}
	if card.Stability <= 0 {
		t.Errorf("Expected positive stability, got %f", card.Stability)
	}
	if card.Reps != 1 {
		t.Errorf("Expected 1 rep, got %d", card.Reps)
	}

	lastStability := card.Stability
	
	// Second review (Good)
	nextReview := now.Add(24 * time.Hour)
	card.Review(Good, nextReview)

	if card.State != 2 {
		t.Errorf("Expected state 2 (Review), got %d", card.State)
	}
	if card.Stability <= lastStability {
		t.Errorf("Expected stability to increase, got %f -> %f", lastStability, card.Stability)
	}
	if card.Reps != 2 {
		t.Errorf("Expected 2 reps, got %d", card.Reps)
	}
}

func TestCard_Again(t *testing.T) {
	card := NewCard()
	now := time.Now()

	card.Review(Again, now)
	if card.Stability != 0.1 {
		t.Errorf("Expected low stability for 'Again', got %f", card.Stability)
	}

	card.Review(Again, now.Add(1*time.Hour))
	if card.Lapses != 1 {
		t.Errorf("Expected 1 lapse, got %d", card.Lapses)
	}
	if card.State != 3 {
		t.Errorf("Expected state 3 (Relearning), got %d", card.State)
	}
}
