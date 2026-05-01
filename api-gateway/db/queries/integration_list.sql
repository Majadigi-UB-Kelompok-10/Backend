-- name: CreateIntegration :one
INSERT INTO integration_list (service_list_id, endpoint_list_id, title, icon_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetIntegrationById :one
SELECT * FROM integration_list
WHERE integration_list_id = $1 LIMIT 1;

-- name: ListIntegrationsByServiceId :many
SELECT * FROM integration_list
WHERE service_list_id = $1
ORDER BY created_at DESC;

-- name: UpdateIntegration :one
UPDATE integration_list
SET endpoint_list_id = $2, title = $3, icon_url = $4
WHERE integration_list_id = $1
RETURNING *;

-- name: DeleteIntegration :exec
DELETE FROM integration_list
WHERE integration_list_id = $1;

-- name: GetAllIntegrations :many
SELECT * FROM integration_list
ORDER BY created_at DESC;
