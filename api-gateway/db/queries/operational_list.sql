-- name: CreateOperational :one
INSERT INTO operational_list (service_list_id, service_url, address, operational_hour, social_media) 
VALUES ($1, $2, $3, $4, $5) 
RETURNING *;

-- name: GetOperationalByServiceId :one
SELECT * FROM operational_list 
WHERE service_list_id = $1 LIMIT 1;

-- name: UpdateOperational :one
UPDATE operational_list 
SET service_url = $2, address = $3, operational_hour = $4, social_media = $5 
WHERE service_list_id = $1 
RETURNING *;

-- name: DeleteOperational :exec
DELETE FROM operational_list 
WHERE service_list_id = $1;