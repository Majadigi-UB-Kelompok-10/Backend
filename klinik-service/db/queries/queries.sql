-- name: CreateReport :one
INSERT INTO hoax_reports (
    ticket_number, reporter_name, reporter_email, reporter_phone, 
    content, proof_link, proof_image_url
) VALUES (
    sqlc.arg('ticket_number'), 
    sqlc.arg('reporter_name'), 
    sqlc.arg('reporter_email'), 
    sqlc.arg('reporter_phone'), 
    sqlc.arg('content'), 
    sqlc.narg('proof_link'),      
    sqlc.narg('proof_image_url') 
) RETURNING id, ticket_number, created_at;

-- name: GetReportTrackingByTicket :one
SELECT 
    r.id AS report_id,
    r.ticket_number,
    r.reporter_name,
    r.status AS report_status,
    r.created_at AS reported_at,
    n.id AS news_id,
    n.title AS news_title,
    n.image_url AS news_image,
    c.name AS category_name
FROM hoax_reports r
LEFT JOIN hoax_news n ON n.report_id = r.id
LEFT JOIN hoax_categories c ON n.category_id = c.id
WHERE r.ticket_number = UPPER(REPLACE(sqlc.arg('ticket_number')::text, ' ', ''))
LIMIT 1;

-- name: GetDashboardStats :many
SELECT 
    c.id AS category_id,
    c.name AS category_name,
    c.slug AS category_slug,
    c.icon_url,
    COUNT(n.id) AS total_news
FROM hoax_categories c
LEFT JOIN hoax_news n ON c.id = n.category_id
GROUP BY c.id, c.name, c.slug, c.icon_url
ORDER BY c.name ASC;

-- name: GetPublicNews :many
SELECT 
    n.id, n.title, n.slug, n.image_url, n.published_at, 
    c.name AS category_name, c.slug AS category_slug
FROM hoax_news n
JOIN hoax_categories c ON n.category_id = c.id
ORDER BY n.published_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountPublicNews :one
SELECT COUNT(id) FROM hoax_news;

-- name: GetNewsDetail :one
SELECT 
    n.id, n.title, n.description, n.reference_link, n.image_url, n.published_at, 
    c.name AS category_name, c.slug AS category_slug
FROM hoax_news n
JOIN hoax_categories c ON n.category_id = c.id
WHERE n.id = sqlc.arg('news_id')
LIMIT 1;

-- name: SearchPublicNews :many
SELECT 
    n.id, n.title, n.slug, n.image_url, n.published_at, 
    c.name AS category_name, c.slug AS category_slug
FROM hoax_news n
JOIN hoax_categories c ON n.category_id = c.id
WHERE 
    n.search_vector @@ plainto_tsquery('simple', sqlc.arg('keyword')::text)
    OR n.title ILIKE '%' || sqlc.arg('keyword')::text || '%'
ORDER BY 
    ts_rank(n.search_vector, plainto_tsquery('simple', sqlc.arg('keyword')::text)) DESC,
    n.published_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountSearchPublicNews :one
SELECT COUNT(n.id)
FROM hoax_news n
WHERE 
    n.search_vector @@ plainto_tsquery('simple', sqlc.arg('keyword')::text)
    OR n.title ILIKE '%' || sqlc.arg('keyword')::text || '%';

-- name: GetAllReportsAdmin :many
SELECT 
    id, ticket_number, reporter_name, reporter_email, 
    content, proof_link, proof_image_url, status, created_at
FROM hoax_reports
WHERE (sqlc.narg('status_filter')::report_status IS NULL OR status = sqlc.narg('status_filter')::report_status)
ORDER BY created_at ASC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountAllReportsAdmin :one
SELECT COUNT(id) 
FROM hoax_reports
WHERE (sqlc.narg('status_filter')::report_status IS NULL OR status = sqlc.narg('status_filter')::report_status);

-- name: UpdateReportStatus :exec
UPDATE hoax_reports 
SET status = sqlc.arg('status')
WHERE id = sqlc.arg('report_id');

-- name: CreateNewsClarification :one
INSERT INTO hoax_news (
    report_id, category_id, title, slug, description, reference_link, image_url
) VALUES (
    sqlc.narg('report_id'), 
    sqlc.arg('category_id'), 
    sqlc.arg('title'), 
    sqlc.arg('slug'),
    sqlc.arg('description'), 
    sqlc.narg('reference_link'), 
    sqlc.arg('image_url')
) RETURNING id, title, published_at;

-- name: GetAllNewsAdmin :many
SELECT 
    n.id, n.title, c.name AS category_name, n.published_at, n.created_at,
    r.ticket_number
FROM hoax_news n
JOIN hoax_categories c ON n.category_id = c.id
LEFT JOIN hoax_reports r ON n.report_id = r.id
ORDER BY n.created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountAllNewsAdmin :one
SELECT COUNT(id) FROM hoax_news;

-- name: DeleteNews :exec
DELETE FROM hoax_news WHERE id = sqlc.arg('news_id');

-- name: GetAllCategories :many
SELECT id, name, slug, icon_url 
FROM hoax_categories 
ORDER BY name ASC;

-- name: CreateCategory :one
INSERT INTO hoax_categories (name, slug, icon_url) 
VALUES (sqlc.arg('name'), sqlc.arg('slug'), sqlc.narg('icon_url'))
RETURNING id;

-- name: GetNewsDetailBySlug :one
SELECT h.id, h.title, h.slug, h.description, h.reference_link, h.image_url, h.published_at, 
       c.name AS category_name, c.slug AS category_slug
FROM hoax_news h
JOIN hoax_categories c ON h.category_id = c.id
WHERE h.slug = sqlc.arg('slug')
LIMIT 1;