-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(), 
    NOW() AT TIME ZONE 'UTC', 
    NOW() AT TIME ZONE 'UTC', 
    $1
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;
