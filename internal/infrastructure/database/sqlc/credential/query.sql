-- name: CreateCredential :one
INSERT INTO user_credentials (
    id,
    user_id,
    password_hash,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetCredentialByUserID :one
SELECT * FROM user_credentials WHERE user_id = $1;

-- name: UpdateCredentialPassword :one
UPDATE user_credentials SET
    password_hash = $2,
    updated_at = $3
WHERE user_id = $1
RETURNING *;

-- name: DeleteCredentialByUserID :exec
DELETE FROM user_credentials WHERE user_id = $1;

-- name: ExistsCredentialByUserID :one
SELECT EXISTS(SELECT 1 FROM user_credentials WHERE user_id = $1);
