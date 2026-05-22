-- One-time script: run after initial Railway deployment to register
-- all service URLs in the api-gateway endpoints table.
-- Run with: psql "$DATABASE_URL" -f scripts/seed-endpoints.sql

INSERT INTO endpoint_list (slug_name, page_url) VALUES
  ('user',          'http://user-service.railway.internal:8080/api/v1'),
  ('transjatim',    'http://transjatim-service.railway.internal:8080/api/v1'),
  ('bapenda',       'http://bapenda-service.railway.internal:8080/api/v1'),
  ('klinik',        'http://klinik-service.railway.internal:8080/api/v1'),
  ('rssa',          'http://rssa-service.railway.internal:8080/api/v1'),
  ('sidita',        'http://sidita-service.railway.internal:8080/api/v1'),
  ('siskaperbapo',  'http://siskaperbapo-service.railway.internal:8080/api/v1'),
  ('sinaker',       'http://sinaker-service.railway.internal:8080/api/v1'),
  ('jdih',          'http://jdih-service.railway.internal:8080/api/v1'),
  ('bansos',        'http://bansos-service.railway.internal:8080/api/v1')
ON CONFLICT (slug_name) DO UPDATE SET page_url = EXCLUDED.page_url;
