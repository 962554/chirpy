-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(), now(), now(), $1, $2
)
RETURNING *;

-- name: AllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps where id = $1;

-- name: GetUserChirps :many
SELECT * FROM chirps where user_id = $1 ORDER BY created_at ASC;

-- name: DeleteChirp :exec
DELETE FROM chirps where id = $1;

-- name: DeleteChirps :exec
DELETE FROM chirps;
