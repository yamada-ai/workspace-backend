-- name: CreateSession :one
INSERT INTO sessions (user_id, work_name, start_time, planned_end, icon_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at;

-- name: FindSessionByID :one
SELECT id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at
FROM sessions
WHERE id = $1
LIMIT 1;

-- name: FindActiveSessionByUserID :one
SELECT id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at
FROM sessions
WHERE user_id = $1 AND actual_end IS NULL
ORDER BY start_time DESC
LIMIT 1;

-- name: UpdateSessionPlannedEnd :one
UPDATE sessions
SET planned_end = $2, updated_at = $3
WHERE id = $1
RETURNING id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at;

-- name: UpdateSessionWorkName :one
UPDATE sessions
SET work_name = $2, updated_at = $3
WHERE id = $1
RETURNING id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at;

-- name: CompleteSession :one
UPDATE sessions
SET actual_end = $2, updated_at = $3
WHERE id = $1
RETURNING id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at;

-- name: ListUserSessions :many
SELECT id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at
FROM sessions
WHERE user_id = $1
ORDER BY start_time DESC
LIMIT $2 OFFSET $3;

-- name: GetActiveSessions :many
SELECT
  s.id,
  s.user_id,
  s.work_name,
  s.start_time,
  s.planned_end,
  s.actual_end,
  s.icon_id,
  s.created_at,
  s.updated_at,
  u.name as user_name,
  u.tier as user_tier
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.actual_end IS NULL
ORDER BY s.start_time DESC;

-- name: ListUserSessionsForDate :many
SELECT id, user_id, work_name, start_time, planned_end, actual_end, icon_id, created_at, updated_at
FROM sessions
WHERE user_id = $1
  AND start_time >= $2
  AND start_time < $3
ORDER BY start_time DESC;
