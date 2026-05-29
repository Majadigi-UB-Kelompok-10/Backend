# SERVICES DETAILS

### Replace (\n) with actual newlines for postgres text

1. Data Penerima & Info Program Bansos (SAPA BANSOS)

- title: "SAPA BANSOS"
- long_title: "Data Penerima & Info Program Bansos (SAPA BANSOS)"
- description: "Sapa Bansos, kepanjangan dari Sistem Aplikasi Pelayanan dan Aduan Bantuan Sosial. Dia merupakan sistem aplikasi untuk pengecekan bantuan sosial berdasarkan Nomor Induk Kependudukan (NIK).
  \n\n
  Sapa Bansos memudahkan warga Jatim update info program bansos lebih transparan dan akuntabel. Saat ini terdapat sekitar 11 program bansos yang tersalurkan ke masyarakat, di antaranya PKH+, bantuan kemiskinan ekstrem, KPM Jawara, dan lainnya.
  \n\n
  Warga Jatim juga bisa mengecek status penerimaan bantuan secara real-time lewat Sapa Bansos. Pengguna hanya perlu memasukkan NIK untuk mengecek data penerima bansos dan rekapitulasi penyaluran masing-masing program bansos. Jika ada penyaluran yang tak sesuai ketentuan, masyarakat bisa laporkan pengaduan lewat Sapa Bansos."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
    - semantic_label: "Logo Provinsi Jawa Timur"
- categories:
  - "Bantuan Sosial"
- endpoint:
  - endpoint_list_id: "${SAPABANSOS_UUID}"
    - slug_name: "/sapabansos"
    - page_url: "/sapabansos"
- integration:
  - integration_list_id:
    - service_list_id: "${BANSOS_UUID}"
    - endpoint_list_id: "${SAPABANSOS_UUID}"
    - title: "Data Penerima & Info Program Bansos"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${BANSOS_UUID}"
    - service_url: "https://sapabansos.dinsos.jatimprov.go.id/"
    - address: "Jl. Gayung Kebonsari No.56b, Gayungan, Kec. Gayungan, Kota Surabaya, Jawa Timur 60235"
    - operational_hour: {
      "Senin": "24 Jam",
      "Selasa": "24 Jam",
      "Rabu": "24 Jam",
      "Kamis": "24 Jam",
      "Jumat": "24 Jam",
      "Sabtu": "24 Jam",
      "Minggu": "24 Jam"
      }
    - social_media: {
      instagram: "https://www.instagram.com/dinsosjatim/",
      facebook: "https://www.facebook.com/dinsosjatim/",
      youtube: "https://www.youtube.com/channel/UC4lluqMbTSyA8oXTH7Hcc2A"
      }
- policies:
  - policy_list_id:
    - service_list_id: "${BANSOS_UUID}"
    - benefit: "Sapa Bansos memudahkan warga Jatim mengakses info program bansos, jadwal pencairan, dan daftar penerima dana secara real-time, transparan, dan akuntabel."
    - instruction: {
      "Persyaratan": [
      "Penerima bansos terdaftar dalam Data Terpadu Kesejahteraan Sosial (DTKS)",
      "Penerima bansos memenuhi kriteria sosial ekonomi tertentu seperti keluarga miskin atau rentan miskin, lanjut usia, dan penyandang disabilitas berat.",
      "Punya Nomor Induk Kependudukan (NIK) yang valid untuk validasi data penerima.",
      "Tidak berstatus sebagai Aparatur Sipil Negara (ASN), TNI, atau Polri."
      ],
      "Cara Cek Bansos lewat Sapa Bansos": [
      "Akses layanan Sapa Bansos melalui aplikasi Majadigi",
      "Pilih layanan Info Bansos",
      "Masukkan NIK untuk cek data penerima bansos",
      "Pilih informasi program untuk info detail rekapitulasi penyaluran tiap program bansos."
      ]
      }

2. Badan Pendapatan Daerah (BAPENDA) Jawa Timur

- title: "BAPENDA"
- long_title: "Badan Pendapatan Daerah (BAPENDA) Jawa Timur"
- description: "Badan Pendapatan Daerah (BAPENDA) Jawa Timur adalah pusat informasi dan layanan Pajak Kendaraan Bermotor (PKB), Nilai Jual Kendaraan Bermotor (NJKB), dan Pendapatan Asli Daerah (PAD). Melalui website resmi, pengguna bisa cek informasi pajak, simulasi tarif, hingga melihat detail NJKB hanya dalam genggaman tangan.
  \n\n
  Inovasi layanan seperti E-Samsat, layanan Samsat Drive Thru, Samsat Keliling, Samsat Bunda, dan Samsat OPOP, menghadirkan pengalaman baru dalam pembayaran pajak secara online yang lebih efisien dan nyaman. Penuhi kebutuhan informasi dan layanan seputar PKB, NJKB, Bea Balik Nama, dan PAD lewat satu pintu di aplikasi Majadigi."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/bapenda/logo_bapenda.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/bapenda/logo_bapenda.webp"
    - semantic_label: "Logo Bapenda Jatim"
- categories:
  - "Layanan Pajak"
- endpoint:
  - endpoint_list_id: "${PKB_UUID}"
    - slug_name: "/pajak-kendaraan"
    - page_url: "/pajak-kendaraan"
  - endpoint_list_id: "${NJKB_UUID}"
    - slug_name: "/nilai-jual-kendaraan-bermotor"
    - page_url: "/nilai-jual-kendaraan-bermotor"
- integration:
  - integration_list_id:
    - service_list_id: "${BAPENDA_UUID}"
    - endpoint_list_id: "${PKB_UUID}"
    - title: "Info Pajak Kendaraan Bermotor"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/bapenda/logo_bapenda.webp"
  - integration_list_id:
    - service_list_id: "${BAPENDA_UUID}"
    - endpoint_list_id: "${NJKB_UUID}"
    - title: "Info Nilai Jual Kendaraan Bermotor"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/bapenda/logo_bapenda.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${BAPENDA_UUID}"
    - service_url: "https://bapenda.jatimprov.go.id/"
    - address: "Jl. Manyar Kertoarjo No.1, Manyar Sabrangan, Kec. Mulyorejo, Surabaya, Jawa Timur 60116"
    - operational_hour: {
      Senin: "08:00 - 15:30",
      Selasa: "08:00 - 15:30",
      Rabu: "08:00 - 15:30",
      Kamis: "08:00 - 15:30",
      Jumat: "08:00 - 15:30"
      }
    - social_media: null
- policies:
  - policy_list_id:
    - service_list_id: "${BAPENDA_UUID}"
    - benefit: "Sebagai institusi yang berperan penting dalam pengelolaan Pendapatan Asli Daerah, BAPENDA bertujuan meningkatkan transparansi, akuntabilitas, dan kualitas pengelolaan keuangan di tingkat Provinsi dan Kabupaten/Kota di Jawa Timur"
    - instruction: {
      [
      "Cara cek informasi pajak dan nilai jual kendaraan": [
      "Kunjungi laman resmi Bapenda Jatim",
      "Cek info pajak kendaraan bermotor di menu Info, lalu pilih info PKB",
      "Masukkan plat nomor kendaraan",
      "Masukkan 5 digit terakhir nomor rangka",
      "Untuk mengetahui info nilai jual kendaraan, klik info besar PKB dan BBN di menu info. Isi data yang diminta, lalu klil submit"
      ],
      "Pembayaran Pajak Kendaraan Bermotor (PKB) tahunan bisa dilakukan di Kantor Bersama Samsat atau melalui E-Samsat. Aplikasi E-Samsat merupakan sistem pembayaran Pajak Kendaraan Bermotor (PKB), Sumbangan Wajib Dana Kecelakaan Lalu Lintas Jalan (SWDKLLJ), dan/atau Parkir Berlangganan tahunan.
      \n\n
      Pembayaran E-Samsat bisa melalui marketplace, e-wallet, serta Payment Poin Online Bank (PPOB), seperti Indomaret, Alfamart, Alfamidi, Kantor Pos, Agen Badan Usaha Milik Desa (BUMDes/Samsat Bunda), Samsat One Pesantren One Produk (OPOP), Samsat Kampus, dan sebagainya."
      ]
      }

3. Produk Hukum Jawa Timur (JDIH)

- title: "JDIH"
- long_title: "Produk Hukum Jawa Timur (JDIH)"
- description: "Jaringan Dokumentasi dan Informasi Hukum (JDIH) adalah layanan pusat data hukum yang memuat peraturan perundang-undangan, produk hukum pusat maupun daerah, serta dokumen penting lainnya.
  \n\n
  Sistem JDIH dirancang untuk mempermudah akses informasi hukum bagi anggota jaringan dan masyarakat di Jawa Timur, dengan visi menghadirkan layanan informasi hukum yang mudah, cepat, dan akurat menuju masyarakat sadar hukum.
  \n\n
  Misi JDIH:\n 1. Meningkatkan kualitas ragam pelayanan\n 2. Meningkatkan efisiensi dan efektifitas informasi hukum\n 3. Meningkatkan kerjasama kegiatan pendokumentasian produk hukum dalam satu jaringan\n 4. Menjadikan fasilitas yang tersedia untuk kerjasama dan pembentukan jaringan yang seutuhnya \n 5. Pemanfaatan dan pendayagunaan potensi masyarakat sebagai konstributor opini, analisa maupun informasi edukatif"
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/jdih/logo_jdih.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/jdih/logo_jdih_hero.webp"
    - semantic_label: "Ilustrasi JDIH Jawa Timur"
- categories:
  - "Produk Hukum"
- endpoint:
  - endpoint_list_id: "${PHJT_UUID}"
    - slug_name: "/produk-hukum-jawa-timur"
    - page_url: "/produk-hukum-jawa-timur"
- integration:
  - integration_list_id:
    - service_list_id: "${JDIH_UUID}"
    - endpoint_list_id: "${PHJT_UUID}"
    - title: "Produk Hukum Jawa Timur - JDIH"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/jdih/logo_jdih.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${JDIH_UUID}"
    - service_url: "https://jdih.jatimprov.go.id/"
    - address: "Jl. Pahlawan No.110, Alun-alun Contong, Bubutan, Surabaya 60174."
    - operational_hour: {
      Senin: "08:00 - 16:00",
      Selasa: "08:00 - 16:00",
      Rabu: "08:00 - 16:00",
      Kamis: "08:00 - 16:00",
      Jumat: "08:00 - 16:00"
      }
    - social_media: {
      "instagram": "https://www.instagram.com/jdihjatim/",
      "facebook": "https://www.facebook.com/JDIH-Jatimprov-607856333021401/",
      "youtube": "https://www.youtube.com/channel/UCqVqELJZYASq8_1RckTKCnw"
      }
- policies:
  - policy_list_id:
    - service_list_id: "${JDIH_UUID}"
    - benefit: [
      "Akses dokumen hukum lebih cepat, akurat, dan efisien.",
      "JDIH menjadi rujukan utama bagi produk hukum pusat dan daerah",
      "Dukung kolaborasi antar lembaga dalam jaringan dokumentasi hukum",
      "Kontribusi nasional: Berkontribusi dalam keterbukaan informasi hukum di tingkat daerah dan nasional."
      ]
    - instruction: {
      "Kebijakan Privasi": "JDIH Jatim berkomitmen menjaga kerahasiaan dan melindungi data privasi pengguna mulai dari pengumpulan data, penggunaan, dan menjaga informasi pribadi Anda saat anda menggunakan layanan JDIH Provinsi Jawa Timur.",
      "Pengumpulan Data": [
      "Nama, email, alamat, dan informasi lain yang diisi secara sukarela saat menggunakan layanan JDIH.",
      "Data teknis seperti IP address, jenis browser, halaman yang dikunjungi, dan waktu akses melalui cookie atau teknologi serupa."
      ],
      "Penggunaan Informasi": [
      "Peningkatan layanan JDIH Jatim",
      "Memproses permintaan dan pertanyaan pengguna",
      "Meningkatkan pengalaman pengguna di situs laman",
      "Mengirim pembaruan dan info layanan JDIH",
      "Peningkatan layanan melalui analisis dan penelitian",
      "Memenuhi kewajiban hukum dan peraturan yang berlaku."
      ],
      "Perlindungan Informasi": "Kami menggunakan enkripsi, firewall, dan prosedur keamanan lain untuk menjaga kerahasiaan dan integritas data.",
      "Pembagian Informasi": "Kami bisa membagikan data pribadi ke pihak ketiga yang membantu operasional JDIH, seperti layanan hosting. Mereka wajib menjaga privasi sama seperti kami. Kami tidak akan membagikan, menjual atau menyewakan data Anda tanpa izin.",
      "Perubahan Kebijakan Privasi": "Kebijakan privasi ini bisa berubah sewaktu-waktu. Kami akan mengumumkan perubahan di situs JDIH Jatim, jadi silakan cek halaman ini secara berkala."
      }

4. Klinik Hoaks

- title: "Klinik Hoaks"
- long_title: "Klinik Hoaks"
- description: "Klinik Hoaks adalah platform layanan publik yang dibangun oleh Dinas Komunikasi dan Informatika Provinsi Jawa Timur untuk membantu masyarakat memverifikasi kebenaran informasi beredar, terutama soal berita hoaks. Pengguna bisa memastikan apakah informasi yang diterima itu fakta, disinformasi, ujaran kebencian, atau justru hoaks.
  \n\n
  Selain pengecekan fakta, Klinik Hoaks juga menyediakan fitur pelaporan dan konsultasi dengan tim ahli jika ditemukan informasi mencurigakan atau meragukan. Pengguna juga bisa mengecek status permohonan klarifikasi secara real-time hanya dengan memasukkan nomor tiket. Hadirnya Klinik Hoaks sebagai layanan digital diharapkan dapat:\n 1. Mengurangi penyebaran berita hoaks yang menimbulkan keresahan warga Jawa Timur.\n 2. Mendorong transparansi informasi publik.\n 3. Menjadikan warga Jawa Timur menjadi lebih bijak dalam menerima dan menyebarkan informasi.\n\n
  Mari wujudkan ruang digital yang sehat, aman, dan bebas hoaks."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/klinik/logo_klinik_hoak_hero.webp"
    - semantic_label: "Ilustrasi Klinik Hoaks"
- categories:
  - "Cek Hoax"
- endpoint:
  - endpoint_list_id: "${KH_UUID}"
    - slug_name: "/klinik-hoaks"
    - page_url: "/klinik-hoaks"
- integration:
  - integration_list_id:
    - service_list_id: "${KLINIK_UUID}"
    - endpoint_list_id: "${KH_UUID}"
    - title: "Klinik Hoaks"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${KLINIK_UUID}"
    - service_url: "https://klinikhoaks.jatimprov.go.id/"
    - address: "Jl. A. Yani 242 - 244, Gayungan, Surabaya."
    - operational_hour: {
      Senin: "24 Jam",
      Selasa: "24 Jam",
      Rabu: "24 Jam",
      Kamis: "24 Jam",
      Jumat: "24 Jam",
      Sabtu: "24 Jam",
      Minggu: "24 jam"
      }
    - social_media: {
      instagram: "https://www.instagram.com/kominfojatim",
      youtube: "https://www.youtube.com/channel/UCEe1ees-scoEkTQv3he9PJw"
      }
- policies:
  - policy_list_id:
    - service_list_id: "${KLINIK_UUID}"
    - benefit: [
      "Membantu masyarakat memverifikasi informasi yang meragukan agar terhindar dari hoaks.",
      "Melindungi publik dari dampak negatif informasi palsu yang menyesatkan.",
      "Meningkatkan kesadaran dan literasi digital masyarakat melalui klarifikasi informasi.",
      "Mendukung terciptanya ruang digital yang sehat dan bebas hoaks.",
      "Menyediakan akses transparan terhadap hasil verifikasi informasi."
      ]
    - instruction: {
      "Prosedur menggunakan situs Klinik Hoaks:": [
      "Akses laman https://klinikhoaks.jatimprov.go.id/",
      "Masukkan kata kunci informasi atau berita yang dicari",
      "Sistem akan menampilkan hasil temuan sesuai kata kunci yang dicari, termasuk status dan penjelasannya."
      ],
      "Jika informasi yang dicari tidak ditemukan, pengguna bisa ajukan permohonan klarifikasi sebagai berikut:": [
      "Klik menu di laman Klinik Hoaks",
      "Pilih menu permohonan klarifikasi",
      "Isi formulir data dengan informasi yang diminta, lalu klik tombol kirim."
      ]
      }

5. Rumah Sakit Dr. Saiful Anwar (RSSA)

- title: "RSSA"
- long_title: "Rumah Sakit Umum Daerah (RSUD) Dr. Saiful Anwar"
- description: "RSUD Dr. Saiful Anwar awalnya dikenal sebagai Rumah Sakit Celaket yang berfungsi sebagai rumah sakit militer milik KNIL. Pada masa pendudukan Jepang, rumah sakit ini tetap digunakan untuk keperluan militer. Setelah kemerdekaan Republik Indonesia, statusnya berubah menjadi rumah sakit umum.
  \n\n
  Sementara itu, Rumah Sakit Sukun yang berada di bawah naungan Kotapraja Malang difungsikan sebagai rumah sakit darurat. Pada tahun 1947, di tengah situasi pasca Perang Dunia II yang penuh gejolak, Rumah Sakit Celaket kembali berfungsi ganda sebagai rumah sakit militer sekaligus rumah sakit umum demi memenuhi kebutuhan strategis saat itu.
  \n\n
  Seiring waktu, Rumah Sakit Celaket terus berkembang. Tepatnya tanggal 12 November 1979, Gubernur Jawa Timur meresmikan Rumah Sakit Celaket menjadi Rumah Sakit Umum Daerah (RSUD) Dr. Saiful Anwar. Rumah sakit ini juga berperan sebagai tempat praktek Fakultas Kedokteran Universitas Brawijaya dan berbagai institusi pendidikan kesehatan lainnya.
  \n\n
  Pada tahun 2007, RSUD Dr. Saiful Anwar ditetapkan sebagai rumah sakit kelas A, status tertinggi untuk rumah sakit di Indonesia, serta diubah menjadi Badan Layanan Umum pada tahun 2008. Hingga kini, RSUD Dr. Saiful Anwar terus menjadi rumah sakit rujukan utama di wilayah Malang dan sekitarnya dengan layanan medis yang komprehensif dan berkualitas tinggi."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/rsud-saiful/rsud_saiful_anwar.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/rsud-saiful/rsud_saiful_anwar_hero.webp"
    - semantic_label: "Ilustrasi RSUD Saiful Anwar"
- categories:
  - "Layanan Kesehatan"
- endpoint:
  - endpoint_list_id: "${KKR_UUID}"
    - slug_name: "/rssa"
    - page_url: "/rssa"
- integration:
  - integration_list_id:
    - service_list_id: "${RSSA_UUID}"
    - endpoint_list_id: "${KKR_UUID}"
    - title: "Rumah Sakit Dr. Saiful Anwar (RSSA)"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/rsud-saiful/rsud_saiful_anwar.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${RSSA_UUID}"
    - service_url: "https://rsusaifulanwar.jatimprov.go.id/v2/"
    - address: "Jl. Jaksa Agung Suprapto No.2, Klojen, Kec. Klojen, Kota Malang, Jawa Timur 65112"
    - operational_hour: {
      Senin: "07:00 - 13:00",
      Selasa: "07:00 - 13:00",
      Rabu: "07:00 - 13:00",
      Kamis: "07:00 - 13:00",
      Jumat: "07:00 - 11:30"
      }
    - social_media: {
      instagram: "https://www.instagram.com/rssaifulanwar/",
      facebook: "https://web.facebook.com/p/RSUD-DrSaiful-Anwar-100064865763707/?locale=id_ID&_rdc=1&_rdr#",
      youtube: "https://www.youtube.com/channel/UC2BH1k2njD7EynOQBdOEWIg"
      }
- policies:
  - policy_list_id: - service_list_id: "${RSSA_UUID}" - benefit: [
    "Memudahkan pasien mendapatkan layanan rawat jalan dengan jadwal operasional yang jelas.",
    "Tersedia dua opsi pendaftaran: offline langsung di RS atau online melalui aplikasi, sehingga pasien dapat memilih metode sesuai kebutuhan.",
    "Mengurangi antrean fisik di rumah sakit karena adanya sistem pengambilan nomor antrean secara online.",
    "Akses informasi yang mudah melalui website resmi RSSA"
    ] - instruction: {
    "Persyaratan Umum": [
    "Memiliki identitas diri (KTP/SIM/Paspor) atau kartu pasien RSSA.",
    "Untuk pasien BPJS Kesehatan, wajib membawa kartu BPJS dan surat rujukan sesuai ketentuan.",
    "Untuk pasien umum, cukup membawa identitas dan bersedia membayar sesuai tarif yang berlaku.",
    "Pasien harus datang sesuai jadwal dan jam operasional yang ditentukan."
    ],
    "Sistem dan Mekanisme Pendaftaran Online": [
    "Unduh aplikasi Antrian Poliklinik RSSA atau "Mobile JKN" melalui Play Store (Android) atau Apple Store (iOS)",
    "Lakukan registrasi akun dan login",
    "Pilih layanan poliklinik, tanggal kunjungan, dan jam yang diinginkan",
    "Simpan bukti pendaftaran digital (biasanya berupa QR code atau nomor antrean)",
    "Datang ke rumah sakit sesuai jadwal yang telah dipilih untuk verifikasi dan pelayanan"
    ],
    "Prosedur saat Hari Kunjungan untuk Pasien Online": [
    "Datang sesuai waktu yang dipilih di aplikasi",
    "Menuju loket khusus pendaftaran online untuk scan QR code/bukti pendaftaran",
    "Langsung diarahkan menuju poliklinik terkait sesuai jadwal"
    ]
    }

6. Sistem Informasi Daya Tarik Wisata Jawa Timur (SIDITA)

- title: "SIDITA"
- long_title: "Sistem Informasi Daya Tarik Wisata Jawa Timur (SIDITA)"
- description: "Sistem Informasi Daya Tarik Wisata (SIDITA), platform berbasis web untuk media promosi dan informasi destinasi, event, serta akomodasi hotel yang tersebar di Jawa Timur."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sidita/logo_sidita.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sidita/logo_sidita.webp"
    - semantic_label: "Logo Sidita"
- categories:
  - "Eksplorasi Pariwisata"
- endpoint:
  - endpoint_list_id: "${SIDI_UUID}"
    - slug_name: "/sidita"
    - page_url: "/sidita"
- integration:
  - integration_list_id:
    - service_list_id: "${SIDITA_UUID}"
    - endpoint_list_id: "${SIDI_UUID}"
    - title: "SIDITA"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sidita/logo_sidita.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${SIDITA_UUID}"
    - service_url: "https://sidita.disbudpar.jatimprov.go.id/"
    - address: "Jalan Wisata Menanggal, Dukuh Menanggal, Kec. Gayungan Kota Surabaya, Provinsi Jawa Timur 60234"
    - operational_hour: {
      Senin: "00:00 - 23:59",
      Selasa: "00:00 - 23:59",
      Rabu: "00:00 - 23:59",
      Kamis: "00:00 - 23:59",
      Jumat: "00:00 - 23:59",
      }
    - social_media: null
- policies:
  - policy_list_id:
    - service_list_id: "${SIDITA_UUID}"
    - benefit: {
      "Aplikasi SIDITA (Sistem Informasi Daya Tarik Wisata) merupakan platform yang menyediakan layanan informasi terkait data kepariwisataan, khususnya di wilayah Jawa Timur. Manfaat yang diperoleh pengguna dari aplikasi ini antara lain:": [
      "Data dan informasi valid",
      "Fitur maps dan direction ke destinasi tujuan",
      "Data diperbarui secara real time"
      ]
      }
    - instruction: {
      "Pengunjung perlu menyiapkan 3 hal ini untuk menikmati layanan 360 East Java Virtual Tour, seperti:": [
      "Perangkat elektronik, berupa handphone atau laptop",
      "Koneksi internet stabil",
      "Browser yang update"
      ]
      "Sistem": {
      "Layanan SIDITA dilengkapi dengan 2 fitur, yaitu:": [
      "SIDITA berbasis website untuk memudahkan pengunjung menikmati layanannya tanpa perlu instal aplikasi.",
      "Titik koordinat wisata sebagai panduan perjalanan ke lokasi tujuan"
      ]
      }
      }

7. Sistem Informasi Ketenagakerjaan (SINAKER)

- title: "SINAKER"
- long_title: "Sistem Informasi Ketenagakerjaan (SINAKER)"
- description: "Sistem Informasi Ketenagakerjaan (SiNaker) adalah platform yang dibangun untuk mengelola, mengolah, dan menyediakan informasi seputar ketenagakerjaan.
  \n\n
  Platform yang dikembangkan oleh Dinas Tenaga Kerja dan Transmigrasi (Disnakertrans) Provinsi Jawa Timur ini punya sistem yang lebih modern, dan terintegrasi dengan berbagai layanan ketenagakerjaan seperti pendaftaran kerja, info loker, pelatihan vokasi, dan konsultasi.
  \n\n
  SiNaker Jatim memudahkan akses layanan ketenagakerjaan bagi pekerja, pencari kerja, dan pelaku usaha. Dengan sistem yang terintegrasi, pengelolaan data lebih efisien dan akurat. Kehadiran SiNaker mendorong layanan yang terbuka dan mendukung peningkatan kesejahteraan pekerja.
  \n\n
  Sinaker memberikan informasi pelatihan dan lowongan kerja yang terintegrasi dengan semua Unit Pelaksana Teknis Balai Latihan Kerja (UPT BLK) di Provinsi Jawa Timur."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sinaker/logo_disnakertrans.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sinaker/logo_disnakertrans.webp"
    - semantic_label: "logo Disnakertrans"
- categories:
  - "Informasi Ketenagakerjaan"
- endpoint:
  - endpoint_list_id: "${DPK_UUID}"
    - slug_name: "/daftar-pelatihan-kerja"
    - page_url: "/daftar-pelatihan-kerja"
  - endpoint_list_id: "${BLK_UUID}"
    - slug_name: "/balai-latihan-kerja"
    - page_url: "/balai-latihan-kerja"
  - endpoint_list_id: "${CPP_UUID}"
    - slug_name: "/cek-pendaftaran-pelatihan-kerja"
    - page_url: "/cek-pendaftaran-pelatihan-kerja"
- integration:
  - integration_list_id:
    - service_list_id: "${SINAKER_UUID}"
    - endpoint_list_id: "${DPK_UUID}"
    - title: "Daftar Pelatihan Kerja"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sinaker/logo_disnakertrans.webp"
  - integration_list_id:
    - service_list_id: "${SINAKER_UUID}"
    - endpoint_list_id: "${BLK_UUID}"
    - title: "Balai Latihan Kerja"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sinaker/logo_disnakertrans.webp"
  - integration_list_id:
    - service_list_id: "${SINAKER_UUID}"
    - endpoint_list_id: "${CPP_UUID}"
    - title: "Cek Pendaftaran Pelatihan"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/sinaker/logo_disnakertrans.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${SINAKER_UUID}"
    - service_url: "https://disnakertrans.jatimprov.go.id/sinaker/"
    - address: "Jl. Dukuh Menanggal Sel. No.124-126, Dukuh Menanggal, Kec. Gayungan, Surabaya, Jawa Timur 60234"
    - operational_hour: {
      Senin: "07.00 - 15.30",
      Selasa: "07.00 - 15.30",
      Rabu: "07.00 - 15.30",
      Kamis: "07.00 - 15.30",
      Jumat: "07.00 - 14.30"
      }
    - social_media: {
      instagram: "https://www.instagram.com/naker_jatim/",
      facebook: "https://www.facebook.com/disnakerprovjatim",
      youtube: "https://www.youtube.com/channel/UCaynEG6fWEcUtqW7s76tCFA"
      }
- policies:
  - policy_list_id:
    - service_list_id: "${SINAKER_UUID}"
    - benefit: {
      "Melalui SiNaker, pengguna bisa:": [
      "Temukan lowongan kerja lebih mudah sesuai minat dan keahlian",
      "Upgrade skill lewat fitur pelatihan dan sertifikasi uji kompetensi",
      "Urus perizinan dan pengawasan ketenagakerjaan bagi perusahaan",
      "Layanan pengaduan untuk perlindungan tenaga kerja."
      ]
      }
    - instruction: [
      "Kunjungi laman Disnakertrans Jatim untuk akses Pelatihan dan Profil Kelembagaan UPT BLK",
      "Lihat daftar UPT BLK di menu Balai Latihan Kerja",
      "Pilih UPT BLK Kabupaten / Kota yang dituju, lalu klik tombol daftar pelatihan",
      "Lengkapi formulir data diri pendaftaran dan informasi lain yang dibutuhkan hingga proses selesai",
      "Cek pendaftaran dengan memasukkan NIK dan nomor HP yang didaftarkan",
      "Pantau informasi jadwal dan pengumuman secara berkala di website Disnakertrans Jatim"
      ]

8. Harga Bahan Pokok (SISKAPERBAPO)

- title: "SISKAPERBAPO"
- long_title: "Sistem Informasi Ketersediaan dan Perkembangan Harga Bahan Pokok (SISKAPERBAPO)"
- description: "SISKAPERBAPO, singkatan dari Sistem Informasi Ketersediaan dan Perkembangan Harga Bahan Pokok. Dia adalah portal berbasis online yang menyajikan info tren harga dan ketersediaan bahan pokok harian dari seluruh area di Jawa Timur. Diantaranya beras, minyak goreng, sayur mayur, ikan segar, produk olahan, perlengkapan rumah tangga, dan komoditas lain.
  \n\n
  SISKAPERBAPO menyajikan informasi update harga di tingkat konsumen dan produsen, terutama di sentra produksi. Sehingga, masyarakat dan pelaku usaha lebih mudah untuk memantau fluktuasi harga harian secara real-time, di mana pun dan kapan pun. Tak hanya itu, fitur analisis tren harga pasar juga bisa dijadikan dasar perencanaan distribusi, pengambilan keputusan bisnis, hingga pengendalian inflasi. Sebagai wujud komitmen terhadap transparansi dan akuntabilitas publik, SISKAPERBAPO hadir untuk memperkuat ekosistem perdagangan dan mendukung ketahanan pangan di Jawa Timur."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
    - semantic_label: ""
- categories:
  - "Harga Kebutuhan Pokok"
- endpoint:
  - endpoint_list_id: "${HBP_UUID}"
    - slug_name: "/siskaperbapo"
    - page_url: "/siskaperbapo"
- integration:
  - integration_list_id:
    - service_list_id: "${SISKAPERBAPO_UUID}"
    - endpoint_list_id: "${HBP_UUID}"
    - title: "Harga Bahan Pokok"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${SISKAPERBAPO_UUID}"
    - service_url: "https://siskaperbapo.jatimprov.go.id/"
    - address: "Jl. Siwalankerto Utara II/42 Surabaya"
    - operational_hour: {
      Senin: "24 Jam",
      Selasa: "24 Jam",
      Rabu: "24 Jam",
      Kamis: "24 Jam",
      Jumat: "24 Jam",
      Sabtu: "24 Jam",
      Minggu: "24 Jam"
      }
    - social_media: {
      instagram: "https://instagram.com/disperindagprovjatim",
      facebook: "https://www.facebook.com/disperindagprovinsijatim/",
      youtube: "https://www.youtube.com/@disperindagprovjatim"
      }
- policies:
  - policy_list_id:
    - service_list_id: "${SISKAPERBAPO_UUID}"
    - benefit: [
      "Akses informasi harga bahan pokok secara harian dan transparan.",
      "Pemantauan ketersediaan bahan pokok dengan mudah, kapan saja.",
      "Mendukung pengendalian inflasi dan menjaga stabilitas harga bahan pokok."
      ]
    - instruction: {
      "Sistem": "Website milik Disperindag Jatim yang menyediakan data harian harga dan ketersediaan bahan pokok secara detail.",
      "Cek Harga Sembako Lewat SISKAPERBAPO": [
      "Akses situs web resmi SISKAPERBAPO",
      "Telusuri data harga bahan pokok berdasarkan kategori bahan pokok.",
      "Pilih lokasi atau wilayah kabupaten/kota yang ingin dipantau.",
      "Lihat grafik tren harga harian dan informasi ketersediaan barang.",
      "Gunakan fitur perbandingan harga produsen dan konsumen untuk analisis."
      ]
      }

9. TransJatim (Transportasi Publik)

- title: "Transjatim"
- long_title: "Transjatim Ajaib 2.0"
- description: "Trans Jatim AJAIB (Aplikasi Jatim Informasi Bus) 2.0 adalah aplikasi untuk kemudahan akses layanan bus Trans Jatim. Pengguna bisa memantau pergerakan bus dan jumlah penumpang di setiap rute dari layar smartphone. Fitur baru seperti Protect Me & Rambu Suara hadir untuk memberikan peringatan suara saat melewati jalur rawan kecelakaan atau persimpan.
  \n\n
  Inovasi dari Trans Jatim AJAIB tidak hanya meningkatkan kenyamanan, tetapi juga memberi perlindungan ekstra bagi pengguna, khususnya penyandang disabilitas.Tersedia gratis di Play Store dan App Store, Trans Jatim AJAIB 2.0 menjadi solusi transportasi publik yang aman, efisien, dan inklusif bagi seluruh warga Jawa Timur."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/transjatim/logo_transjatim.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/transjatim/logo_transjatim.webp"
    - semantic_label: "Logo TransJatim"
- categories:
  - "Transportasi Publik"
- endpoint:
  - endpoint_list_id: "${TTJ_UUID}"
    - slug_name: "/transjatim"
    - page_url: "/transjatim"
- integration:
  - integration_list_id:
    - service_list_id: "${TRANSJATIM_UUID}"
    - endpoint_list_id: "${TTJ_UUID}"
    - title: "Tiket TransJatim"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/transjatim/logo_transjatim.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${TRANSJATIM_UUID}"
    - service_url: "https://dishub.jatimprov.go.id/"
    - address: "Jl. Johar No.17, Alun-alun Contong, Kec. Bubutan, Surabaya, Jawa Timur 60174"
    - operational_hour: {
      Senin: "05.00 - 21.00",
      Selasa: "05.00 - 21.00",
      Rabu: "05.00 - 21.00",
      Kamis: "05.00 - 21.00",
      Jumat: "05.00 - 21.00",
      Sabtu: "05.00 - 21.00",
      Minggu: "05:00 - 21:00"
      }
    - social_media: {
      instagram: "https://www.instagram.com/officialtransjatim/",
      facebook: "https://www.facebook.com/dishubjawatimur",
      youtube: "https://www.youtube.com/channel/UCguXvUk1_z620b8Hr9O9CRQ"
      }
- policies:
  - policy_list_id:
    - service_list_id: "${TRANSJATIM_UUID}"
    - benefit: "Memberikan kemudahaan kepada Pengguna Transjatim untuk mengakses informasi terupdate terkait layanan Transjatim"
    - instruction: {
      "Panduan naik Bus Trans Jatim": [
      "Unduh aplikasi Transjatim Ajaib di Play Store dan App Store.",
      "Tentukan tujuan perjalanan Anda, lalu cek jadwal, rute, dan halte atau shelter terdekat.",
      "Beli tiket melalui aplikasi (e-money) atau secara langsung (uang tunai) saat di lokasi. Anda juga bisa membeli tiket Trans Jatim luxury langsung dari aplikasi. Untuk pembelian langsung di tempat, sebaiknya siapkan uang pas.",
      "Pastikan Anda naik di halte terdekat dan sesuai dengan rute Trans Jatim.",
      "Silakan duduk di tempat yang tersedia. Prioritaskan tempat duduk untuk lansia, ibu hamil, dan penyandang disabilitas. Pastikan Anda juga memantau titik pemberhentian bus supaya tidak kelewatan."
      ]
      }

10. Nomor Darurat Jawa Timur (Nomor Penting)

- title: "NODA"
- long_title: "Nomor Darurat Jawa Timur"
- description: "Nomor darurat merupakan layanan cepat tanggap dari pemerintah atau instansi terkait untuk memberikan bantuan kepada masyarakat. Nomor ini dapat dihubungi saat warga menghadapi situasi mendesak, berbahaya, atau yang mengancam nyawa—seperti kecelakaan, kebakaran, bencana alam, gangguan keamanan, hingga kondisi medis gawat darurat.
\n\n
Sejak 2015, pemerintah Indonesia menerapkan Program Layanan Call Center 112 di berbagai daerah di Indonesia. Nomor darurat sengaja dibuat singkat agar mudah diingat dan bisa diakses dengan cepat. Selain Call Center 112, masing-masing wilayah di Indonesia juga memiliki nomor darurat khusus yang bisa mempercepat penanganan. Di Jawa Timur misalnya, tiap instansi menyediakan nomor darurat khusus yang bisa diakses 24 jam dan bebas pulsa.
\n\n
Agar berjalan efektif, warga dihimbau tidak melakukan panggilan iseng. Informasi yang jelas dan tepat saat melapor akan membantu petugas memberikan respon cepat dan tepat sasaran."
- icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- images:
  - image_list_id:
    - image_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
    - semantic_label: "Logo Provinsi Jawa Timur"
- categories:
  - "Layanan Darurat"
- endpoint:
  - endpoint_list_id: "${NODA_UUID}"
    - slug_name: "/nodajatim"
    - page_url: "/nodajatim"
- integration:
  - integration_list_id:
    - service_list_id: "${NOMOR_DARURAT_UUID}"
    - endpoint_list_id: "${NODA_UUID}"
    - title: "Kontak Darurat Jawa Timur"
    - icon_url: "https://nhsdrdhzkogczngslvvh.supabase.co/storage/v1/object/public/image-asset/shared/logo-provinsi-jawa-timur.webp"
- operational:
  - operational_list_id:
    - service_list_id: "${NOMOR_DARURAT_UUID}"
    - service_url: ""
    - address: "Seluruh kota"
    - operational_hour: {
      Senin: "24 Jam",
      Selasa: "24 Jam",
      Rabu: "24 Jam",
      Kamis: "24 Jam",
      Jumat: "24 Jam",
      Sabtu: "24 Jam",
      Minggu: "24 Jam"
      }
    - social_media: {}
- policies:
  - policy_list_id:
    - service_list_id: "${NOMOR_DARURAT_UUID}"
    - benefit: "Layanan nomor darurat memberikan akses cepat dan mudah untuk mencari pertolongan saat seseorang mengalami atau mengetahui situasi darurat seperti kebakaran, banjir, kecelakaan lalu lintas, kriminalitas, dan lainnya. Dengan begitu, layanan ini diharapkan mampu mempercepat penanganan keadaan darurat dan meminimalisir dampak buruk yang muncul akibat situasi darurat. Layanan nomor darurat beroperasi 24 jam sehari, dan 7 hari seminggu. Sehingga masyarakat bisa mengaksesnya kapanpun dan dari manapun."
    - instruction: "Kontak darurat biasanya lebih pendek atau sedikit dengan tujuan agar mudah diingat. Pastikan Anda menyimpan daftar kontak darurat di ponsel Anda atau tempat yang mudah dijangkau. Terakhir, pastikan Anda memberikan informasi secara jelas mengenai kejadian dan lokasinya agar petugas bisa mengeksekusinya lebih cepat."