-- name: CreateEndpoint :one
INSERT INTO endpoint_list (slug_name, page_url)
VALUES ($1, $2)
RETURNING *;

-- name: GetEndpointById :one
SELECT * FROM endpoint_list
WHERE endpoint_list_id = $1 LIMIT 1;

-- name: GetEndpointBySlug :one
SELECT * FROM endpoint_list
WHERE slug_name = $1 LIMIT 1;

-- name: ListEndpoints :many
SELECT * FROM endpoint_list
ORDER BY created_at DESC;

-- name: UpdateEndpoint :one
UPDATE endpoint_list
SET slug_name = $2, page_url = $3
WHERE endpoint_list_id = $1
RETURNING *;

-- name: DeleteEndpoint :exec
DELETE FROM endpoint_list
WHERE endpoint_list_id = $1;

-- name: GetAllEndpoints :many
SELECT * FROM endpoint_list
ORDER BY created_at DESC;
