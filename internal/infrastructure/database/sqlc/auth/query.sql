-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id,
    user_id,
    token_hash,
    expires_at,
    revoked_at,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1;

-- name: GetRefreshTokensByUserID :many
SELECT * FROM refresh_tokens 
WHERE user_id = $1 
  AND revoked_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET
    revoked_at = NOW()
WHERE token_hash = $1
RETURNING *;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET
    revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens 
WHERE expires_at < NOW() 
   OR revoked_at IS NOT NULL;

-- name: IsRefreshTokenValid :one
SELECT EXISTS(
    SELECT 1 FROM refresh_tokens 
    WHERE token_hash = $1 
      AND revoked_at IS NULL 
      AND expires_at > NOW()
);
