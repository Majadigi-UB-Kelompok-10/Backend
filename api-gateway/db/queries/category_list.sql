-- name: CreateCategory :one
INSERT INTO category_list (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetCategoryById :one
SELECT * FROM category_list
WHERE category_list_id = $1 LIMIT 1;

-- name: GetCategoryByName :one
SELECT * FROM category_list
WHERE name = $1 LIMIT 1;

-- name: ListCategories :many
SELECT * FROM category_list
ORDER BY name ASC;

-- name: UpdateCategory :one
UPDATE category_list
SET name = $2, description = $3
WHERE category_list_id = $1
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM category_list
WHERE category_list_id = $1;

-- name: GetAllCategories :many
SELECT * FROM category_list
ORDER BY created_at DESC;
