-- name: CreateImage :one
INSERT INTO image_list (service_list_id, image_url, semantic_label) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: GetImageById :one
SELECT * FROM image_list 
WHERE image_list_id = $1 LIMIT 1;

-- name: ListImagesByServiceId :many
SELECT * FROM image_list 
WHERE service_list_id = $1 
ORDER BY created_at DESC;

-- name: UpdateImage :one
UPDATE image_list 
SET image_url = $2, semantic_label = $3 
WHERE image_list_id = $1 
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM image_list 
WHERE image_list_id = $1;