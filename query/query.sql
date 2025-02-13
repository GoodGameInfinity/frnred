-- name: GetVanity :one
SELECT * FROM vanities
WHERE id = $1 LIMIT 1;

-- name: CreateVanity :one
INSERT INTO vanities (
  id, url
) VALUES (
  $1, $2
)
RETURNING *;

-- name: DeleteVanity :exec
DELETE FROM urls
WHERE id = $1;


-- name: GetUrl :one
SELECT * FROM urls
WHERE id = $1 LIMIT 1;


-- name: CreateUrl :one
INSERT INTO urls (
  id, url
) VALUES (
  $1, $2
)
RETURNING *;

-- name: DeleteUrl :exec
DELETE FROM vanities
WHERE id = $1;


-- name: CreateKey :one
INSERT INTO keys (
  id, hashed, admin
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: DeleteKey :exec
DELETE FROM keys
WHERE id = $1;

-- name: FindKey :one
SELECT * FROM keys WHERE hashed = $1;

-- name: CheckKey :one
SELECT 1 hashed FROM keys;

-- name: ListKeys :many
SELECT id, admin FROM keys;