-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    email, password_hash
) VALUES (
    $1, $2
)
RETURNING *;

-- name: UpdateUserLimits :one
UPDATE users
SET daily_new_limit = $2,
    retention_goal = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
