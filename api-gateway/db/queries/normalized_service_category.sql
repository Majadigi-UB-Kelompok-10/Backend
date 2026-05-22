-- name: GetAllNormalizedServiceCategory :many
SELECT
    sl.service_list_id,
    sl.title,
    sl.long_title,
    sl.description,
    sl.icon_url,
    array_agg(cl.category_list_id::text) FILTER (WHERE cl.category_list_id IS NOT NULL) AS category_ids,
    sl.created_at
FROM service_list sl
LEFT JOIN services_has_categories shc ON sl.service_list_id = shc.service_list_id
LEFT JOIN category_list cl ON shc.category_list_id = cl.category_list_id
GROUP BY sl.service_list_id, sl.title, sl.long_title, sl.description, sl.icon_url, sl.created_at
ORDER BY sl.created_at DESC;

-- name: GetNormalizedServiceCategoryById :one
SELECT
    sl.service_list_id,
    sl.title,
    sl.long_title,
    sl.description,
    sl.icon_url,
    array_agg(cl.category_list_id::text) FILTER (WHERE cl.category_list_id IS NOT NULL) AS category_ids,
    sl.created_at
FROM service_list sl
LEFT JOIN services_has_categories shc ON sl.service_list_id = shc.service_list_id
LEFT JOIN category_list cl ON shc.category_list_id = cl.category_list_id
WHERE sl.service_list_id = $1
GROUP BY sl.service_list_id, sl.title, sl.long_title, sl.description, sl.icon_url, sl.created_at;
