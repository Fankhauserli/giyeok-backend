-- name: ListWords :many
SELECT * FROM words
ORDER BY topik_level, korean
LIMIT $1 OFFSET $2;

-- name: CreateWord :one
INSERT INTO words (
    korean, english, part_of_speech, topik_level
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (korean, english) DO UPDATE SET
    part_of_speech = COALESCE(NULLIF(EXCLUDED.part_of_speech, ''), words.part_of_speech),
    topik_level = COALESCE(NULLIF(EXCLUDED.topik_level, 0), words.topik_level)
RETURNING *;

-- name: GetWordByID :one
SELECT * FROM words
WHERE id = $1 LIMIT 1;
