package fsrs

import (
	"math"
	"time"
)

// FSRS parameters and basic logic
// Reference: https://github.com/open-spaced-repetition/fsrs4anki/wiki/The-Algorithm

type Card struct {
	Due           time.Time
	Stability     float64
	Difficulty    float64
	ElapsedDays   int
	ScheduledDays int
	Reps          int
	Lapses        int
	State         int // 0: New, 1: Learning, 2: Review, 3: Relearning
	LastReview    time.Time
}

type Rating int

const (
	Again Rating = 1
	Hard  Rating = 2
	Good  Rating = 3
	Easy  Rating = 4
)

// Default weights for FSRS
var DefaultWeights = []float64{0.4, 0.6, 2.4, 5.8, 4.93, 0.94, 0.86, 0.01, 1.49, 0.14, 0.94, 2.18, 0.05, 0.34, 1.26, 0.26, 2.05}

func NewCard() Card {
	return Card{
		Due:        time.Now(),
		Stability:  0,
		Difficulty: 0,
		State:      0,
	}
}

func (c *Card) Review(rating Rating, now time.Time, retentionGoal float64) {
	if c.State == 0 { // New
		c.init(rating)
		c.State = 1 // Learning
	} else {
		// Simplified FSRS-like update for brevity in this prototype
		// In a real app, use the full formula
		interval := now.Sub(c.LastReview).Hours() / 24
		c.ElapsedDays = int(interval)
		
		// Very simplified updates
		switch rating {
		case Again:
			c.Difficulty = math.Min(10, c.Difficulty+0.5)
			c.Stability = 0.0007 // ~1 minute (1/1440 days)
			c.Lapses++
			c.State = 3 // Relearning
		case Hard:
			c.Difficulty = math.Min(10, c.Difficulty+0.2)
			c.Stability = math.Max(1.0, c.Stability * 1.2)
			c.State = 2 // Review
		case Good:
			c.Stability = math.Max(1.0, c.Stability * 2.5)
			c.State = 2 // Review
		case Easy:
			c.Difficulty = math.Max(1, c.Difficulty-0.2)
			c.Stability = math.Max(1.0, c.Stability * 3.5)
			c.State = 2 // Review
		}
		
		c.Reps++
	}
	
	c.LastReview = now
	
	// If stability is very low (less than 1 day), use minutes
	if c.Stability < 1.0 {
		minutes := int(c.Stability * 1440)
		if minutes < 1 { minutes = 1 }
		c.Due = now.Add(time.Duration(minutes) * time.Minute)
		c.ScheduledDays = 0
	} else {
		// Calculate interval based on stability and retention goal
		// Formula: I = S * ln(R) / ln(0.9)
		// We default to 0.9 as the 'standard' stability base
		multiplier := math.Log(retentionGoal) / math.Log(0.90)
		interval := c.Stability * multiplier
		
		c.ScheduledDays = int(math.Max(1, interval))
		c.Due = now.Add(time.Duration(c.ScheduledDays) * 24 * time.Hour)
	}
}

func (c *Card) init(rating Rating) {
	// Initial stability based on first rating
	switch rating {
	case Again:
		c.Stability = 0.0007 // ~1 minute
		c.Difficulty = 5.0
	case Hard:
		c.Stability = 0.0035 // ~5 minutes
		c.Difficulty = 5.0
	case Good:
		c.Stability = 0.007 // ~10 minutes
		c.Difficulty = 5.0
	case Easy:
		c.Stability = 4.0 // 4 days
		c.Difficulty = 3.0
	}
	c.Reps = 1
}
