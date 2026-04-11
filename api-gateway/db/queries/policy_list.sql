-- name: CreatePolicy :one
INSERT INTO policy_list (service_list_id, benefit, instruction) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: GetPolicyByServiceId :one
SELECT * FROM policy_list 
WHERE service_list_id = $1 LIMIT 1;

-- name: UpdatePolicy :one
UPDATE policy_list 
SET benefit = $2, instruction = $3 
WHERE service_list_id = $1 
RETURNING *;

-- name: DeletePolicy :exec
DELETE FROM policy_list 
WHERE service_list_id = $1;