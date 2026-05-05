-- name: GetReview :one
SELECT * FROM user_reviews
WHERE user_id = $1 AND word_id = $2;

-- name: ListDueReviews :many
SELECT w.id, w.korean, w.english, w.part_of_speech, w.topik_level, w.created_at, r.state, r.due, r.stability, r.difficulty, r.elapsed_days, r.scheduled_days, r.reps, r.lapses, r.last_review
FROM words w
JOIN user_reviews r ON w.id = r.word_id
WHERE r.user_id = $1 AND r.due <= $2
AND (w.topik_level = $4 OR $4 = 0)
ORDER BY r.due ASC
LIMIT $3;

-- name: UpsertReview :one
INSERT INTO user_reviews (
    user_id, word_id, state, due, stability, difficulty, elapsed_days, scheduled_days, reps, lapses, last_review, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW()
)
ON CONFLICT (user_id, word_id) DO UPDATE SET
    state = EXCLUDED.state,
    due = EXCLUDED.due,
    stability = EXCLUDED.stability,
    difficulty = EXCLUDED.difficulty,
    elapsed_days = EXCLUDED.elapsed_days,
    scheduled_days = EXCLUDED.scheduled_days,
    reps = EXCLUDED.reps,
    lapses = EXCLUDED.lapses,
    last_review = EXCLUDED.last_review,
    updated_at = NOW()
RETURNING *;

-- name: GetNewWordsForUser :many
SELECT * FROM words
WHERE id NOT IN (
    SELECT word_id FROM user_reviews WHERE user_id = $1
)
AND (topik_level = $3 OR $3 = 0)
ORDER BY RANDOM()
LIMIT $2;

-- name: GetLevelProficiency :many
SELECT 
    topik_level,
    COUNT(*) as total,
    COUNT(r.id) as reviewed
FROM words
LEFT JOIN user_reviews r ON words.id = r.word_id AND r.user_id = $1
GROUP BY topik_level
ORDER BY topik_level;

-- name: CountNewWordsStartedToday :one
SELECT count(*) FROM user_reviews
WHERE user_id = $1 
AND created_at >= $2;

-- name: GetWeeklyActivity :many
SELECT date_trunc('day', updated_at)::date as study_date
FROM user_reviews
WHERE user_id = $1 AND updated_at >= $2
GROUP BY study_date
ORDER BY study_date DESC;

-- name: ListLibraryWords :many
SELECT 
    w.id, w.korean, w.english, w.part_of_speech, w.topik_level,
    COALESCE(r.state, -1)::int as state,
    r.due, r.reps
FROM words w
LEFT JOIN user_reviews r ON w.id = r.word_id AND r.user_id = $1
ORDER BY 
    state DESC, 
    w.topik_level ASC, 
    w.korean ASC
LIMIT $2 OFFSET $3;
