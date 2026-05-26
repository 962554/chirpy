-- name: AddToken :exec
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
$1, now(), now(), $2, now() + interval '60 day'
)
RETURNING *;

-- name: RevokeToken :exec
UPDATE refresh_tokens SET updated_at = now(), revoked_at = now() WHERE token = $1;

-- name: GetToken :one
SELECT * from refresh_tokens WHERE token = $1;

-- name: DeleteTokens :exec
DELETE FROM refresh_tokens;
