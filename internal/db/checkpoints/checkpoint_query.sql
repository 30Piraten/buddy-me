-- name: CreateCheckpoint :one
INSERT INTO checkpoints (roadmap_id, title, description, position, type, status, estimated_time, reward_points, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
RETURNING *; 

-- name: GetCheckpoint :one
SELECT * FROM checkpoints WHERE id = $1; 

-- name: ListCheckpoints :many
SELECT * FROM checkpoints WHERE roadmap_id = $1 ORDER BY position ASC;

-- name: UpdateCheckpoint :one
UPDATE checkpoints
SET 
    title = $2,
    description = $3,
    position = $4,
    type = $5,
    status = $6,
    estimated_time = $7,
    reward_points = $8
WHERE id = $1
RETURNING *;

-- name: DeleteCheckpoint :one
DELETE FROM checkpoints
WHERE id = $1 
RETURNING *;

-- name: ListUserCheckpoints :many
SELECT c.*
FROM checkpoints c
JOIN user_checkpoints uc ON c.id = uc.checkpoint_id 
WHERE uc.user_id = sqlc.arg(user_id)
  AND (sqlc.arg(roadmap_id) IS NULL OR c.roadmap_id = sqlc.arg(roadmap_id)::uuid)
  AND (sqlc.arg(status) IS NULL OR c.status = sqlc.arg(status))
ORDER BY c.position ASC;