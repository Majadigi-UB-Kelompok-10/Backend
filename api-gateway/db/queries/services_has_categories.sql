-- name: AddCategoryToService :exec
INSERT INTO services_has_categories (service_list_id, category_list_id)
VALUES ($1, $2);

-- name: RemoveCategoryFromService :exec
DELETE FROM services_has_categories
WHERE service_list_id = $1 AND category_list_id = $2;

-- name: ListCategoriesByServiceId :many
SELECT c.* FROM category_list c
JOIN services_has_categories sc ON c.category_list_id = sc.category_list_id
WHERE sc.service_list_id = $1;

-- name: ListServicesByCategoryId :many
SELECT s.* FROM service_list s
JOIN services_has_categories sc ON s.service_list_id = sc.service_list_id
WHERE sc.category_list_id = $1;

-- name: GetAllServicesHasCategories :many
SELECT * FROM services_has_categories
ORDER BY created_at DESC;

-- name: ListServicesByCategories :many
SELECT DISTINCT s.*
FROM service_list s
JOIN services_has_categories sc ON s.service_list_id = sc.service_list_id
WHERE sc.category_list_id = ANY($1::uuid[])
ORDER BY s.title;
