-- name: SaveRefreshToken :one
INSERT 
INTO refresh_tokens 
(token, user_id, expires_at, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3, 
    NOW() AT TIME ZONE 'UTC',
    NOW() AT TIME ZONE 'UTC'
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW() AT TIME ZONE 'UTC', 
    updated_at = NOW() AT TIME ZONE 'UTC'
WHERE token = $1;