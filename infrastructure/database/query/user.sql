-- name: FindUserByName :one
SELECT id, name, tier, created_at, updated_at
FROM users
WHERE name = $1
LIMIT 1;

-- name: FindUserByID :one
SELECT id, name, tier, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (name, tier, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING id, name, tier, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET tier = $2, updated_at = $3
WHERE id = $1
RETURNING id, name, tier, created_at, updated_at;

-- name: FindUserByNameForUpdate :one
SELECT id, name, tier, created_at, updated_at
FROM users
WHERE name = $1
FOR UPDATE
LIMIT 1;
