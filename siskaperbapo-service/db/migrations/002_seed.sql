--
-- PostgreSQL database dump
--

\restrict ISeaDcJZnIIqjcHnbOD1ek9UM28DK5lOiNRI6QisRwpDjM1JtGpmufpOuDdmT8k

-- Dumped from database version 14.20 (Homebrew)
-- Dumped by pg_dump version 14.20 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: bahan_pokok; Type: TABLE DATA; Schema: public; Owner: dzaky
--

INSERT INTO public.bahan_pokok VALUES (1, 'Bawang Merah / Kg', 'bawang-merah', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/bawang-merah.webp');
INSERT INTO public.bahan_pokok VALUES (2, 'Beras Medium / Kg', 'beras-medium', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/beras-premium.webp');
INSERT INTO public.bahan_pokok VALUES (3, 'Bawang Putih / Kg', 'bawang-putih', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/bawang-putih.webp');
INSERT INTO public.bahan_pokok VALUES (4, 'Cabai Rawit / Kg', 'cabai-rawit', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/cabai-rawit.webp');
INSERT INTO public.bahan_pokok VALUES (5, 'Cabai Merah / Kg', 'cabai-merah', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/cabe-merah.webp');
INSERT INTO public.bahan_pokok VALUES (6, 'Tepung Terigu / Kg', 'tepung-terigu', 'kg', 'https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/siskaperbapo/tepung-terigu.webp');

--
-- Data for Name: master_area; Type: TABLE DATA; Schema: public; Owner: dzaky
--

INSERT INTO public.master_area VALUES (1, 'Bangkalan', 'bangkalan');
INSERT INTO public.master_area VALUES (2, 'Banyuwangi', 'banyuwangi');
INSERT INTO public.master_area VALUES (3, 'Bojonegoro', 'bojonegoro');
INSERT INTO public.master_area VALUES (4, 'Bondowoso', 'bondowoso');
INSERT INTO public.master_area VALUES (5, 'Gresik', 'gresik');
INSERT INTO public.master_area VALUES (6, 'Jember', 'jember');
INSERT INTO public.master_area VALUES (7, 'Jombang', 'jombang');
INSERT INTO public.master_area VALUES (8, 'Lamongan', 'lamongan');
INSERT INTO public.master_area VALUES (9, 'Lumajang', 'lumajang');
INSERT INTO public.master_area VALUES (10, 'Magetan', 'magetan');
INSERT INTO public.master_area VALUES (11, 'Nganjuk', 'nganjuk');
INSERT INTO public.master_area VALUES (12, 'Ngawi', 'ngawi');
INSERT INTO public.master_area VALUES (13, 'Pacitan', 'pacitan');
INSERT INTO public.master_area VALUES (14, 'Pamekasan', 'pamekasan');
INSERT INTO public.master_area VALUES (15, 'Ponorogo', 'ponorogo');
INSERT INTO public.master_area VALUES (16, 'Sampang', 'sampang');
INSERT INTO public.master_area VALUES (17, 'Sidoarjo', 'sidoarjo');
INSERT INTO public.master_area VALUES (18, 'Situbondo', 'situbondo');
INSERT INTO public.master_area VALUES (19, 'Sumenep', 'sumenep');
INSERT INTO public.master_area VALUES (20, 'Surabaya', 'surabaya');
INSERT INTO public.master_area VALUES (21, 'Trenggalek', 'trenggalek');
INSERT INTO public.master_area VALUES (22, 'Tuban', 'tuban');
INSERT INTO public.master_area VALUES (23, 'Tulungagung', 'tulungagung');
INSERT INTO public.master_area VALUES (24, 'Batu', 'batu');
INSERT INTO public.master_area VALUES (25, 'Blitar', 'blitar');
INSERT INTO public.master_area VALUES (26, 'Kabupaten Blitar', 'kab-blitar');
INSERT INTO public.master_area VALUES (27, 'Kediri', 'kediri');
INSERT INTO public.master_area VALUES (28, 'Kabupaten Kediri', 'kab-kediri');
INSERT INTO public.master_area VALUES (29, 'Madiun', 'madiun');
INSERT INTO public.master_area VALUES (30, 'Kabupaten Madiun', 'kab-madiun');
INSERT INTO public.master_area VALUES (31, 'Malang', 'malang');
INSERT INTO public.master_area VALUES (32, 'Kabupaten Malang', 'kab-malang');
INSERT INTO public.master_area VALUES (33, 'Mojokerto', 'mojokerto');
INSERT INTO public.master_area VALUES (34, 'Kabupaten Mojokerto', 'kab-mojokerto');
INSERT INTO public.master_area VALUES (35, 'Pasuruan', 'pasuruan');
INSERT INTO public.master_area VALUES (36, 'Kabupaten Pasuruan', 'kab-pasuruan');
INSERT INTO public.master_area VALUES (37, 'Probolinggo', 'probolinggo');
INSERT INTO public.master_area VALUES (38, 'Kabupaten Probolinggo', 'kab-probolinggo');


--
-- Data for Name: harga_harian; Type: TABLE DATA; Schema: public; Owner: dzaky
--

INSERT INTO public.harga_harian VALUES (1, 1, 20, 17000, '2026-03-27');
INSERT INTO public.harga_harian VALUES (2, 1, 29, 17000, '2026-03-27');
INSERT INTO public.harga_harian VALUES (4, 2, 7, 15000, '2026-03-27');
INSERT INTO public.harga_harian VALUES (3, 1, 7, 15000, '2026-03-27');
INSERT INTO public.harga_harian VALUES (6, 1, 7, 15000, '2026-03-26');
INSERT INTO public.harga_harian VALUES (7, 1, 20, 15000, '2026-03-26');
INSERT INTO public.harga_harian VALUES (8, 1, 20, 19000, '2026-03-29');
INSERT INTO public.harga_harian VALUES (9, 1, 20, 21000, '2026-03-28');
INSERT INTO public.harga_harian VALUES (11, 1, 20, 21000, '2026-03-25');
INSERT INTO public.harga_harian VALUES (12, 3, 20, 21000, '2026-03-25');
INSERT INTO public.harga_harian VALUES (13, 4, 20, 21000, '2026-03-25');
INSERT INTO public.harga_harian VALUES (15, 5, 20, 21000, '2026-03-25');
INSERT INTO public.harga_harian VALUES (16, 6, 20, 21000, '2026-03-25');


--
-- Name: bahan_pokok_id_seq; Type: SEQUENCE SET; Schema: public; Owner: dzaky
--

SELECT pg_catalog.setval('public.bahan_pokok_id_seq', 2, true);


--
-- Name: harga_harian_id_seq; Type: SEQUENCE SET; Schema: public; Owner: dzaky
--

SELECT pg_catalog.setval('public.harga_harian_id_seq', 11, true);


--
-- Name: master_area_id_seq; Type: SEQUENCE SET; Schema: public; Owner: dzaky
--

SELECT pg_catalog.setval('public.master_area_id_seq', 38, true);


--
-- PostgreSQL database dump complete
--

\unrestrict ISeaDcJZnIIqjcHnbOD1ek9UM28DK5lOiNRI6QisRwpDjM1JtGpmufpOuDdmT8k

