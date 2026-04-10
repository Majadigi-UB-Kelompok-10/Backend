--
-- PostgreSQL database dump
--

-- \restrict Oa3ZepIBStJz7h4tgBqKjTFFpGQuRE0OwbUDhNWQq8bLyATqjc1Obt7JX8rSvZ5

-- Dumped from database version 17.6
-- Dumped by pg_dump version 17.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: service_list; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO "public"."service_list" ("service_list_id", "title", "description", "icon_url", "created_at") VALUES
	('1d81c635-4f53-4060-b849-a62e80aab1fa', 'Harga Bahan Pokok (SISKAPERBAPO)', 'SISKAPERBAPO, singkatan dari Sistem Informasi Ketersediaan dan Perkembangan Harga Bahan Pokok. Dia adalah portal berbasis online yang menyajikan info tren harga dan ketersediaan bahan pokok harian dari seluruh area di Jawa Timur. Diantaranya beras, minyak goreng, sayur mayur, ikan segar, produk olahan, perlengkapan rumah tangga, dan komoditas lain.

SISKAPERBAPO menyajikan informasi update harga di tingkat konsumen dan produsen, terutama di sentra produksi. Sehingga, masyarakat dan pelaku usaha lebih mudah untuk memantau fluktuasi harga harian secara real-time, di mana pun dan kapan pun. Tak hanya itu, fitur analisis tren harga pasar juga bisa dijadikan dasar perencanaan distribusi, pengambilan keputusan bisnis, hingga pengendalian inflasi. Sebagai wujud komitmen terhadap transparansi dan akuntabilitas publik, SISKAPERBAPO hadir untuk memperkuat ekosistem perdagangan dan mendukung ketahanan pangan di Jawa Timur.', 'siskaperbapo/logo-provinsi-jawa-timur.webp', '2026-03-22 09:59:21.573827+00');


--
-- Data for Name: image_list; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO "public"."image_list" ("image_list_id", "service_list_id", "image_url", "semantic_label", "created_at") VALUES
	('02a37660-2649-488c-b97f-35caea0cf04d', '1d81c635-4f53-4060-b849-a62e80aab1fa', 'siskaperbapo/logo-provinsi-jawa-timur.webp', 'Logo Provinsi Jawa Timur', '2026-03-24 12:34:33.672666+00');

--
-- Data for Name: operational_list; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO "public"."operational_list" ("operational_list_id", "service_list_id", "service_url", "address", "operational_hour", "social_media", "created_at") VALUES
	('22091e47-3a61-4a23-92a7-92bf8e565e9f', '1d81c635-4f53-4060-b849-a62e80aab1fa', 'https://siskaperbapo.jatimprov.go.id/', 'Jl. Siwalankerto Utara II/42 Surabaya', '{"rabu": "24 Jam", "jumat": "24 Jam", "kamis": "24 Jam", "sabtu": "24 Jam", "senin": "24 Jam", "minggu": "24 Jam", "selasa": "24 Jam"}', '{"youtube": "https://www.youtube.com/@disperindagprovjatim", "facebook": "https://www.facebook.com/disperindagprovinsijatim/", "instagram": "https://instagram.com/disperindagprovjatim"}', '2026-03-25 02:00:48.127829+00');


--
-- Data for Name: policy_list; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO "public"."policy_list" ("policy_list_id", "service_list_id", "benefit", "instruction", "created_at") VALUES
	('cba13b3d-d65f-4d04-b90c-bf35ff2f2528', '1d81c635-4f53-4060-b849-a62e80aab1fa', '{"Manfaat": {"1": "Akses informasi harga bahan pokok secara harian dan transparan.", "2": "Pemantauan ketersediaan bahan pokok dengan mudah, kapan saja.", "3": "Mendukung pengendalian inflasi dan menjaga stabilitas harga bahan pokok."}}', '{"Sistem": "Website milik Disperindag Jatim yang menyediakan data harian harga dan ketersediaan bahan pokok secara detail.", "Cek Harga Sembako Lewat SISKAPERBAPO": {"1": "Akses situs web resmi SISKAPERBAPO", "2": "Telusuri data harga bahan pokok berdasarkan kategori bahan pokok/", "3": "Pilih lokasi atau wilayah kabupaten/kota yang ingin dipantau.", "4": "Lihat grafik tren harga harian dan informasi ketersediaan barang.", "5": "Gunakan fitur perbandingan harga produsen dan konsumen untuk analisis."}}', '2026-03-25 02:58:48.169506+00');

--
-- PostgreSQL database dump complete
--

-- \unrestrict Oa3ZepIBStJz7h4tgBqKjTFFpGQuRE0OwbUDhNWQq8bLyATqjc1Obt7JX8rSvZ5

-- RESET ALL;
