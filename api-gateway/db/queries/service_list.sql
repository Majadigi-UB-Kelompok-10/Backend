-- name: CreateService :one
INSERT INTO service_list (title, long_title, description, icon_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetServiceById :one
SELECT * FROM service_list
WHERE service_list_id = $1 LIMIT 1;

-- name: ListServices :many
SELECT * FROM service_list
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: SearchServices :many
SELECT * FROM service_list
WHERE title ILIKE $1
ORDER BY title ASC;

-- name: UpdateService :one
UPDATE service_list
SET title = $2, long_title = $3, description = $4, icon_url = $5
WHERE service_list_id = $1
RETURNING *;

-- name: DeleteService :exec
DELETE FROM service_list
WHERE service_list_id = $1;

-- name: GetAllServices :many
SELECT * FROM service_list
ORDER BY created_at DESC;

-- name: CountServices :one
SELECT COUNT(*) FROM service_list;
