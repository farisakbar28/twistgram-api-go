# TWISTGRAM

Disusun sebagai dokumen perencanaan produk & rekayasa perangkat lunak

untuk pengembangan perangkat lunak

Juni 2026

# Daftar Isi

	Daftar Isi	2

	1. Pendahuluan	5

	1.1 Latar Belakang	5

	1.2 Tujuan Dokumen	5

	1.3 Ruang Lingkup Produk	5

	1.4 Target Pengguna	5

	1.5 Definisi, Akronim, dan Istilah	6

	2. Ringkasan Eksekutif	7

	2.1 Visi Produk	7

	2.2 Strategi Pengembangan	7

	2.3 Tech Stack Utama	7

	2.4 Model Relasi Sosial	7

	3. Modul: Autentikasi & Keamanan Akun	8

	3.1 Fitur: Registrasi Akun	8

	3.2 Fitur: Login	8

	3.3 Fitur: Lupa Password	8

	3.4 Fitur: Lupa Username / Pemulihan Email	9

	3.5 Fitur Keamanan Tambahan	9

	3.6 Aturan Bisnis (Business Rules)	9

	4. Modul: Profil, Privasi & Relasi Sosial	10

	4.1 Fitur: Profil Pengguna	10

	4.2 Fitur: Privasi Akun	10

	4.3 Fitur: Follow System	10

	4.4 Fitur: Close Friends	10

	4.5 Fitur: Block & Report	11

	4.6 Perbandingan Block vs Unfollow vs Remove Follower	11

	4.7 Aturan Bisnis (Business Rules)	11

	5. Modul: Feed, Konten (Post) & Story	11

	5.1 Fitur: Beranda (Home Feed)	11

	5.2 Fitur: Membuat & Mengunggah Post	12

	5.3 Fitur: Interaksi Post	12

	5.4 Fitur: Manajemen Post	12

	5.5 Fitur: Story	13

	5.6 Aturan Bisnis (Business Rules)	13

	6. Modul: Pencarian & Discovery	14

	6.1 Fitur: Pencarian	14

	6.2 Aturan Bisnis	14

	7. Modul: Pesan Langsung (Direct Message / Chat)	15

	7.1 Fitur Inti	15

	7.2 Catatan Teknis	15

	7.3 Aturan Bisnis	15

	8. Modul: Notifikasi	16

	8.1 Fitur Inti	16

	8.2 Aturan Bisnis	16

	9. Modul: Moderasi & Keamanan Konten	17

	9.1 Fitur Inti	17

	9.2 Aturan Bisnis	17

	10. Desain Basis Data (Skema Tingkat ERD)	17

	10.1 Tabel: users	17

	10.2 Tabel: user_interests	18

	10.3 Tabel: follows	18

	10.4 Tabel: blocks	18

	10.5 Tabel: posts	19

	10.6 Tabel: post_media	19

	10.7 Tabel: post_tags (mention pengguna pada post)	19

	10.8 Tabel: hashtags & post_hashtags	20

	10.9 Tabel: likes (polymorphic untuk post & comment)	20

	10.10 Tabel: comments	20

	10.11 Tabel: saved_posts	21

	10.12 Tabel: stories	22

	10.13 Tabel: story_views	22

	10.14 Tabel: story_tags & story_highlights [ADV]	22

	10.15 Tabel: conversations & messages	22

	10.16 Tabel: notifications	23

	10.17 Tabel: reports	23

	10.18 Relasi Antar Entitas (Ringkasan)	24

	11. Daftar API Endpoint (REST)	24

	11.1 Authentication	24

	11.2 Users & Profile	25

	11.3 Follow & Relasi Sosial	25

	11.4 Posts & Feed	26

	11.5 Interaksi Post (Like, Comment, Save, Share)	26

	11.6 Story	26

	11.7 Pencarian	27

	11.8 Direct Message	27

	11.9 Notifikasi & Report	27

	12. Arsitektur Sistem	27

	12.1 Gambaran Umum Arsitektur	27

	12.2 Diagram Lapisan (Layered Architecture)	28

	12.3 Komponen Pendukung	28

	12.4 Pertimbangan Keamanan (Security Considerations)	28

	12.5 Pertimbangan Skalabilitas	29

	13. Roadmap & Fase Pengembangan	30

	13.1 Fase 1 — MVP (Estimasi 3-5 bulan, dikerjakan paralel dengan kuliah)	30

	13.2 Fase 2 — Advanced (Estimasi 3-4 bulan berikutnya, opsional/lanjutan)	30

	13.3 Matriks Prioritas Fitur	31

	14. Penutup	32

	14.1 Catatan Penggunaan Dokumen	32

	14.2 Rekomendasi Langkah Selanjutnya	32

	14.3 Riwayat Revisi	32

# 1. Pendahuluan

## 1.1 Latar Belakang

Twistgram adalah aplikasi media sosial berbasis foto, video, dan teks yang dirancang untuk memungkinkan pengguna terhubung, berbagi momen, dan menemukan konten yang relevan dengan minat masing-masing. Twistgram mengadopsi model interaksi sosial berbasis follow (seperti Instagram), dilengkapi dengan konsep Close Friends untuk berbagi konten secara lebih privat kepada lingkar pertemanan terdekat.

Dokumen ini disusun sebagai Software Requirements & Technical Design Document (SRS/TDD) yang menjadi acuan tunggal dalam proses perencanaan, pengembangan, dan evaluasi produk Twistgram. Dokumen ini ditulis dengan standar yang setara dengan dokumen perencanaan produk pada tim rekayasa perangkat lunak profesional.

## 1.2 Tujuan Dokumen

- Mendefinisikan seluruh ruang lingkup fungsional dan non-fungsional aplikasi Twistgram secara terstruktur.

- Menjadi acuan teknis bagi tim pengembang dalam merancang basis data, arsitektur sistem, dan API.

- Membagi fitur ke dalam fase pengembangan (MVP dan Advanced) agar pengembangan dapat dilakukan secara bertahap dan realistis.

- Menjadi dokumentasi yang dapat ditunjukkan sebagai bukti kompetensi perencanaan dan perancangan sistem perangkat lunak end-to-end.

## 1.3 Ruang Lingkup Produk

Twistgram mencakup modul-modul inti sebagai berikut:

- Autentikasi & Keamanan Akun

- Manajemen Profil & Privasi

- Relasi Sosial (Follow, Close Friends, Block)

- Manajemen Konten (Feed, Post, Story)

- Interaksi Konten (Like, Comment, Share, Save)

- Pencarian & Discovery

- Pesan Langsung (Direct Message / Chat)

- Notifikasi

- Moderasi & Keamanan (Report, Block)

## 1.4 Target Pengguna

- Pengguna umum berusia 16 tahun ke atas yang ingin berbagi dan menemukan konten visual (foto/video).

- Kreator konten yang ingin membangun audiens berdasarkan minat tertentu.

- Pengguna yang menginginkan kontrol privasi granular atas siapa yang dapat melihat kontennya.

## 1.5 Definisi, Akronim, dan Istilah

| **Istilah** | **Definisi** |
| --- | --- |
| MVP | Minimum Viable Product — fitur minimum yang wajib ada agar aplikasi dapat digunakan dan didemokan secara utuh. |
| Advanced | Fitur tambahan yang meningkatkan kelengkapan produk namun tidak wajib ada di rilis pertama. |
| Follower / Following | Follower adalah pengguna yang mengikuti akun lain; Following adalah akun yang diikuti. |
| Close Friends | Daftar khusus pengguna tepercaya yang dapat melihat story/konten privat tertentu. |
| OTP | One-Time Password — kode sekali pakai untuk verifikasi identitas. |
| RLS | Row Level Security — mekanisme keamanan tingkat baris data pada Supabase/PostgreSQL. |
| JWT | JSON Web Token — format token untuk autentikasi sesi pengguna. |
| Soft delete | Penghapusan data secara logis (ditandai terhapus) tanpa menghapus baris dari database secara permanen. |

# 2. Ringkasan Eksekutif

## 2.1 Visi Produk

Menjadi platform media sosial yang memungkinkan pengguna berbagi momen visual secara autentik, dengan kontrol privasi yang jelas dan sistem penemuan konten yang relevan berdasarkan minat maupun relasi sosial pengguna.

## 2.2 Strategi Pengembangan

Mengingat kompleksitas fitur media sosial skala penuh, pengembangan Twistgram dibagi menjadi dua fase utama:

**Fase**** 1 — MVP (Minimum Viable Product)**

Berisi seluruh fitur inti yang membuat Twistgram berfungsi sebagai media sosial yang utuh dan dapat didemokan: autentikasi, profil, follow system, feed, posting, story dasar, interaksi konten, pencarian, dan notifikasi dasar.

**Fase**** 2 — Advanced Features**

Berisi fitur penyempurnaan yang meningkatkan daya saing produk: musik pada konten/story, close friends, chat real-time lanjutan, rekomendasi berbasis algoritma, dan fitur keamanan lanjutan.

Pembagian ini ditandai secara konsisten di seluruh dokumen menggunakan label [MVP] dan [ADV] pada setiap fitur, sehingga tim pengembang dapat memprioritaskan pekerjaan secara jelas.

## 2.3 Tech Stack Utama

| **Layer** | **Teknologi** | **Keterangan** |
| --- | --- | --- |
| Backend API | Go (Golang) | REST API, dapat dikembangkan dengan framework seperti Fiber/Echo/Gin |
| Database | PostgreSQL (via Supabase) | Relational database utama |
| Auth | Supabase Auth | Manajemen sesi, JWT, OTP email/SMS |
| File Storage | Supabase Storage | Penyimpanan foto, video, audio story |
| Realtime | Supabase Realtime | Notifikasi live, chat, status online |
| Frontend Web | React.js | Aplikasi web utama (prioritas pertama) |
| Frontend Mobile | React Native | Aplikasi mobile (opsional, prioritas kedua) |
| API Testing | Postman | Dokumentasi & pengujian endpoint |
| Container | Docker | Containerization layanan Go untuk deployment konsisten |

## 2.4 Model Relasi Sosial

Twistgram menggunakan model Follow (satu arah, tanpa perlu persetujuan kecuali akun bersifat privat), dikombinasikan dengan konsep Close Friends sebagai sub-grup khusus dari following untuk berbagi story secara lebih privat. Model ini dipilih karena lebih sesuai dengan sifat konten Twistgram yang berorientasi pada discovery dan distribusi konten luas, bukan relasi dua arah yang ketat seperti pada model pertemanan (Friend system).

# 3. Modul: Autentikasi & Keamanan Akun

Modul ini menangani seluruh siklus identitas pengguna: pendaftaran, login, pemulihan akses, dan keamanan sesi. Modul ini menjadi fondasi keamanan seluruh aplikasi sehingga harus diselesaikan secara solid pada Fase MVP.

## 3.1 Fitur: Registrasi Akun

- **[MVP] **Input data: nama lengkap, email, nomor telepon (opsional), username, password.

- **[MVP] **Username otomatis disarankan sebagai placeholder dari nama (dapat diubah oleh pengguna sebelum submit).

- **[MVP] **Validasi password: minimum 8 karakter, mengandung minimal 1 huruf besar di awal, serta direkomendasikan kombinasi angka/simbol.

- **[MVP] **Validasi keunikan email dan username secara real-time (cek ketersediaan saat pengguna mengetik).

- **[MVP] **Verifikasi akun via kode OTP yang dikirim ke email (wajib) sebelum akun dapat digunakan secara penuh.

- **[ADV] **Verifikasi tambahan via nomor telepon (jika diisi), untuk mengaktifkan recovery via SMS.

- **[ADV] **Login/registrasi via OAuth (Google) sebagai alternatif registrasi manual.

Catatan desain: nomor telepon bersifat opsional pada saat registrasi, namun sistem akan menyimpan flag phone_verified secara terpisah dari email_verified. Hal ini penting karena memengaruhi jalur pemulihan akun (lihat 3.3 dan 3.4).

## 3.2 Fitur: Login

- **[MVP] **Login menggunakan kombinasi email ATAU username, dengan password.

- **[MVP] **Validasi kredensial dengan pesan error yang aman (tidak membocorkan apakah email/username terdaftar atau tidak — demi keamanan, gunakan pesan generik “kredensial tidak valid”).

- **[MVP] **Sesi login menggunakan JWT (access token + refresh token) yang dikelola oleh Supabase Auth.

- **[ADV] **Manajemen sesi aktif: pengguna dapat melihat daftar device yang sedang login dan melakukan logout jarak jauh dari device tertentu.

- **[ADV] **Two-Factor Authentication (2FA) via OTP saat login dari device baru.

## 3.3 Fitur: Lupa Password

Alur (flow) pemulihan password:

- Pengguna memasukkan email ATAU nomor telepon terdaftar.

- Sistem mengirimkan kode OTP (6 digit, berlaku 5-10 menit) ke email/nomor telepon yang dipilih.

- Pengguna memasukkan kode OTP untuk verifikasi.

- Setelah kode valid, sistem menampilkan form input password baru.

- Password baru divalidasi dengan aturan yang sama seperti saat registrasi (lihat 3.1).

- Seluruh sesi aktif sebelumnya otomatis di-invalidate setelah password berhasil diganti (praktik keamanan standar).

- **[MVP] **Reset password via OTP email.

- **[ADV] **Reset password via OTP SMS (jika nomor telepon terverifikasi).

## 3.4 Fitur: Lupa Username / Pemulihan Email

Bagian ini memerlukan penyesuaian logika dibanding permintaan awal, karena terdapat inkonsistensi yang perlu diluruskan: pengguna yang lupa email tidak mungkin menerima kode verifikasi di email yang sama yang ia lupakan. Maka alur yang digunakan adalah sebagai berikut:

**Skenario**** A — ****Lupa**** Username (email ****masih**** ****diketahui****)**

- Pengguna memasukkan email terdaftar.

- Sistem mengirimkan kode OTP ke email tersebut.

- Setelah verifikasi berhasil, sistem menampilkan username yang terdaftar dan memberi opsi mengubahnya.

**Skenario**** B — ****Lupa**** Email (****nomor**** ****telepon**** ****terverifikasi**** ****tersedia****)**

- Pengguna memasukkan username dan nomor telepon yang terdaftar.

- Sistem mengirimkan kode OTP via SMS ke nomor tersebut.

- Setelah verifikasi berhasil, pengguna dapat melihat email yang tersamarkan (contoh: a***@gmail.com) dan mengganti email.

**Skenario**** C — ****Lupa**** Email DAN ****tidak**** ****ada**** ****nomor**** ****telepon**** ****terverifikasi**

- Sistem mengarahkan pengguna ke proses verifikasi manual (contoh: form bantuan/support) karena tidak ada jalur otomatis yang aman untuk diverifikasi.

- **[MVP] **Skenario A dan B.

- **[ADV] **Skenario C (form bantuan manual / customer support flow).

## 3.5 Fitur Keamanan Tambahan

- **[ADV] **Two-Factor Authentication (2FA).

- **[ADV] **Log aktivitas login (waktu, lokasi perkiraan, device).

- **[ADV] **Notifikasi email saat ada login dari device/lokasi baru.

## 3.6 Aturan Bisnis (Business Rules)

| **Kode** | **Aturan** |
| --- | --- |
| AUTH-01 | Password minimum 8 karakter, huruf besar di awal kata pertama. |
| AUTH-02 | Akun yang belum verifikasi email tidak dapat mengakses fitur posting/follow, hanya dapat melihat onboarding. |
| AUTH-03 | Kode OTP berlaku maksimum 10 menit dan hanya dapat digunakan satu kali. |
| AUTH-04 | Maksimum 5 kali percobaan login gagal dalam 15 menit sebelum akun dikenakan cooldown sementara (rate limiting). |
| AUTH-05 | Penggantian password/email otomatis menghentikan seluruh sesi aktif lainnya. |

# 4. Modul: Profil, Privasi & Relasi Sosial

## 4.1 Fitur: Profil Pengguna

- **[MVP] **Lihat profil: foto profil, nama, username, bio singkat, jumlah following/followers, jumlah post.

- **[MVP] **Edit profil: ubah nama, username, bio, foto profil, link eksternal.

- **[MVP] **Pilih minat/kategori konten (digunakan untuk personalisasi feed, lihat Modul 5).

- **[ADV] **Foto sampul/header profil.

- **[ADV] **Badge verifikasi (centang biru) untuk akun tertentu — opsional, dapat berupa fitur admin-only di awal.

## 4.2 Fitur: Privasi Akun

- **[MVP] **Toggle akun Publik vs Privat.

- Akun publik: siapa pun dapat melihat post dan langsung follow tanpa approval.

- Akun privat: calon follower harus mengirim permintaan follow yang harus disetujui pemilik akun.

- **[MVP] **Daftar permintaan follow masuk (untuk akun privat), dengan aksi Approve / Decline.

## 4.3 Fitur: Follow System

- **[MVP] **Follow / Unfollow pengguna lain.

- **[MVP] **Lihat daftar Following dan Followers.

- **[MVP] **Remove follower (pemilik akun dapat menghapus seseorang dari daftar followernya).

## 4.4 Fitur: Close Friends

- **[ADV] **Tambah/hapus pengguna dari daftar Close Friends (subset dari following).

- **[ADV] **Opsi “bagikan ke Close Friends saja” saat upload Story.

- **[ADV] **Indikator visual (lingkaran hijau) pada story yang dibagikan ke Close Friends.

- **[ADV] **Pengguna dapat menghapus dirinya sendiri dari daftar Close Friends milik orang lain (opsional, fitur lanjutan).

## 4.5 Fitur: Block & Report

- **[MVP] **Block pengguna: mencegah pengguna yang diblokir melihat profil, konten, dan mengirim pesan.

- **[MVP] **Unblock pengguna.

- **[MVP] **Report pengguna atau konten (kategori: spam, konten tidak pantas, pelecehan, akun palsu, lainnya).

- **[ADV] **Dashboard moderasi sederhana (admin-only) untuk meninjau report yang masuk.

## 4.6 Perbandingan Block vs Unfollow vs Remove Follower

| **Aksi** | **Efek**** pada Following/Follower** | **Efek**** pada ****Visibilitas** |
| --- | --- | --- |
| Unfollow | Saya berhenti follow orang lain | Saya tidak lagi melihat post mereka di feed |
| Remove Follower | Orang lain dihapus dari follower saya | Mereka tidak lagi melihat post privat saya |
| Block | Relasi follow kedua arah dihapus otomatis | Kedua pihak tidak dapat melihat profil/konten satu sama lain, tidak dapat mengirim pesan |

## 4.7 Aturan Bisnis (Business Rules)

| **K****ode** | **Aturan** |
| --- | --- |
| SOC-01 | Akun privat: post tidak muncul di hasil pencarian/explore bagi non-follower. |
| SOC-02 | Block bersifat mutual: jika A block B, maka B juga tidak dapat melihat/menghubungi A. |
| SOC-03 | Close Friends maksimal hanya dapat memuat akun yang sudah berstatus following. |
| SOC-04 | Permintaan follow ke akun privat otomatis expired/dapat dibatalkan oleh pengirim sebelum disetujui. |
| SOC-05 | Edit username pada saat edit profil hanya bisa dilakukan sekali dalam sebulan. |

# 5. Modul: Feed, Konten (Post) & Story

## 5.1 Fitur: Beranda (Home Feed)

- **[MVP] **Menampilkan konten dari akun yang di-follow, diurutkan secara kronologis (terbaru di atas) pada versi awal MVP.

- **[MVP] **Jika pengguna belum follow siapa pun atau ingin variasi konten, sistem menampilkan konten dari kategori minat yang dipilih saat onboarding.

- **[ADV] **Algoritma ranking sederhana: kombinasi kedekatan relasi (following), tingkat engagement (like/comment), dan kesesuaian minat — menggantikan urutan kronologis murni.

- **[ADV] **Tab “For You”/Explore yang terpisah dari feed following, khusus menampilkan konten berdasarkan minat & trending.

## 5.2 Fitur: Membuat & Mengunggah Post

- **[MVP] **Upload konten berupa gambar atau video (1 media per post pada versi awal).

- **[MVP] **Menulis caption.

- **[MVP] **Tag pengguna lain pada konten (menandai posisi pada gambar, atau sekadar mention pada caption).

- **[MVP] **Hashtag pada caption (searchable, lihat Modul 7).

- **[ADV] **Multi-media dalam satu post (carousel, hingga 10 item).

- **[ADV] **Menambahkan musik/audio pada konten (memerlukan library musik berlisensi atau audio milik pengguna sendiri — lihat catatan teknis di bawah).

- **[ADV] **Location tagging (geotag).

- **[ADV] **Filter dasar pada gambar (brightness, contrast, crop) sebelum upload.

Catatan teknis — fitur musik: penambahan musik berlisensi (seperti pada Instagram/TikTok) memerlukan kerja sama lisensi musik dengan label/agregator (contoh: layanan seperti Epidemic Sound API) yang secara realistis di luar scope proyek ini. Rekomendasi: implementasikan sebagai upload audio milik pengguna sendiri (royalty-free) sebagai pengganti, dan dokumentasikan keterbatasan ini secara transparan sebagai “trade-off teknis yang disadari” — hal ini justru menjadi nilai tambah karena menunjukkan pemahaman tentang batasan lisensi konten pihak ketiga.

## 5.3 Fitur: Interaksi Post

- **[MVP] **Like / Unlike post.

- **[MVP] **Comment pada post.

- **[MVP] **Reply pada comment (nested/threaded, minimal 1 level kedalaman).

- **[MVP] **Like pada comment.

- **[MVP] **Share post (ke Direct Message internal, atau generate link eksternal).

- **[MVP] **Save / Unsave post ke koleksi tersimpan pribadi.

- **[ADV] **Pin comment tertentu (oleh pemilik post).

- **[ADV] **Hide/turn-off comment section untuk post tertentu.

- **[ADV] **Koleksi tersimpan (collections) — pengguna dapat mengelompokkan saved post ke folder bernama, bukan satu daftar besar.

## 5.4 Fitur: Manajemen Post

- **[MVP] **Delete post (soft delete — lihat catatan teknis).

- **[MVP] **Archive / Unarchive post (post diarsipkan tidak tampil di profil/feed publik namun tidak terhapus).

- **[MVP] **Edit caption setelah post diunggah (media tidak dapat diedit, hanya caption/tag).

Catatan teknis — delete: gunakan soft delete (kolom deleted_at) bukan penghapusan baris permanen, agar data terkait (komentar, like) tidak menimbulkan orphan record dan memudahkan fitur “restore” di masa depan.

## 5.5 Fitur: Story

- **[MVP] **Upload story berupa foto, video, atau teks.

- **[MVP] **Story otomatis kedaluwarsa (hilang) setelah 24 jam.

- **[MVP] **Melihat story dari akun yang di-follow, ditampilkan sebagai daftar lingkaran horizontal di atas Beranda.

- **[MVP] **Reply terhadap story — masuk sebagai Direct Message (DM), BUKAN komentar publik.

- **[MVP] **Daftar siapa saja yang telah melihat story milik sendiri (story viewers).

- **[ADV] **Tag pengguna lain pada story.

- **[ADV] **Menambahkan musik/audio pada story (dengan catatan keterbatasan lisensi yang sama seperti pada 5.2).

- **[ADV] **Story Highlight — menyimpan story tertentu secara permanen di profil, dikelompokkan dalam koleksi bernama.

- **[ADV] **Bagikan story khusus ke Close Friends (terhubung dengan Modul 4.4).

- **[ADV] **Sticker interaktif: polling, pertanyaan (Q&A), kuis.

## 5.6 Aturan Bisnis (Business Rules)

| **Ko****de** | **Aturan** |
| --- | --- |
| CNT-01 | Story otomatis non-aktif (hidden dari viewer) tepat 24 jam setelah waktu upload, namun tetap tersimpan di backend selama 7 hari untuk keperluan recovery sebelum dihapus permanen. |
| CNT-02 | Reply terhadap story selalu menghasilkan entri Direct Message baru/lanjutan, tidak pernah berupa komentar publik. |
| CNT-03 | Post yang diarsipkan tetap dapat dilihat oleh pemiliknya sendiri, namun hilang dari profil publik dan feed follower. |
| CNT-04 | Hanya pemilik post yang dapat menghapus, mengarsipkan, atau mengedit caption post tersebut. |
| CNT-05 | Tag pengguna pada post/story memerlukan notifikasi ke pengguna yang ditandai, dan pengguna tersebut berhak menghapus tag dari dirinya sendiri. |

# 6. Modul: Pencarian & Discovery

## 6.1 Fitur: Pencarian

- **[MVP] **Pencarian pengguna berdasarkan username atau nama.

- **[MVP] **Pencarian berdasarkan hashtag.

- **[ADV] **Riwayat pencarian terbaru (recent searches), dapat dihapus oleh pengguna.

- **[ADV] **Halaman Explore/Discover terpisah — grid konten trending/relevan berdasarkan minat, terlepas dari following.

## 6.2 Aturan Bisnis

| **Kode** | **Aturan** |
| --- | --- |
| SRCH-01 | Hasil pencarian pengguna tidak menampilkan akun yang sudah melakukan block terhadap pencari (atau sebaliknya). |
| SRCH-02 | Post dari akun privat tidak muncul di hasil pencarian hashtag/explore bagi non-follower. |

# 7. Modul: Pesan Langsung (Direct Message / Chat)

## 7.1 Fitur Inti

- **[MVP] **Kirim & terima pesan teks 1-ke-1 antar pengguna yang saling tidak memblokir.

- **[MVP] **Riwayat percakapan tersimpan dan dapat di-scroll (pagination).

- **[MVP] **Kirim media (gambar) dalam chat.

- **[MVP] **Reply story masuk sebagai pesan dalam thread chat (terhubung dengan Modul 5.5).

- **[ADV] **Status online / last seen.

- **[ADV] **Read receipt (tanda pesan telah dibaca).

- **[ADV] **Typing indicator (“sedang mengetik...”).

- **[ADV] **Group chat (lebih dari 2 partisipan).

- **[ADV] **Unsend/hapus pesan (untuk semua partisipan).

- **[ADV] **Kirim video/voice note dalam chat.

## 7.2 Catatan Teknis

Fitur real-time (status online, typing indicator, pesan masuk instan) direkomendasikan menggunakan Supabase Realtime (berbasis PostgreSQL logical replication/WebSocket), sehingga tidak perlu membangun infrastruktur WebSocket terpisah di sisi Go pada fase awal. Jika di kemudian hari diperlukan kontrol lebih granular, layanan chat dapat dipisah menjadi service Go tersendiri yang berkomunikasi via WebSocket native.

## 7.3 Aturan Bisnis

| **Kode** | **Aturan** |
| --- | --- |
| MSG-01 | Pengguna yang saling block tidak dapat mengirim/menerima pesan satu sama lain. |
| MSG-02 | Untuk akun privat, permintaan pesan dari non-follower masuk ke folder “Permintaan Pesan” terpisah, bukan langsung ke kotak masuk utama. |

# 8. Modul: Notifikasi

## 8.1 Fitur Inti

- **[MVP] **Notifikasi dalam aplikasi (in-app) untuk: like, comment, follow baru, permintaan follow, mention/tag, reply story.

- **[MVP] **Pusat notifikasi (daftar riwayat notifikasi) dengan status terbaca/belum terbaca.

- **[ADV] **Push notification (mobile, via Firebase Cloud Messaging atau Expo Notifications untuk React Native).

- **[ADV] **Pengaturan preferensi notifikasi (pengguna dapat menonaktifkan jenis notifikasi tertentu).

## 8.2 Aturan Bisnis

| **Kode** | **Aturan** |
| --- | --- |
| NTF-01 | Notifikasi tidak dikirim untuk aksi yang dilakukan terhadap diri sendiri (contoh: like pada post sendiri). |
| NTF-02 | Notifikasi dari akun yang diblokir tidak ditampilkan. |

# 9. Modul: Moderasi & Keamanan Konten

## 9.1 Fitur Inti

- **[MVP] **Report konten/pengguna dengan kategori alasan.

- **[ADV] **Dashboard admin sederhana untuk meninjau report dan mengambil tindakan (hapus konten, suspend akun).

- **[ADV] **Filter otomatis kata kasar pada caption/komentar (basic keyword filtering).

## 9.2 Aturan Bisnis

| **Ko****de** | **Aturan** |
| --- | --- |
| MOD-01 | Setiap report tersimpan dengan status: pending, reviewed, action_taken, dismissed. |
| MOD-02 | Akun yang menerima lebih dari N report tervalidasi dalam periode tertentu otomatis ditandai untuk peninjauan manual (ambang batas N ditentukan saat implementasi). |

# 10. Desain Basis Data (Skema Tingkat ERD)

Skema berikut menggunakan pendekatan relational (PostgreSQL/Supabase). Tipe data ditulis dalam notasi umum PostgreSQL. Seluruh tabel menggunakan primary key bertipe UUID untuk konsistensi dengan Supabase Auth, dan menerapkan soft delete (kolom deleted_at) pada entitas yang relevan.

## 10.1 Tabel: users

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) | Mengikuti id dari Supabase Auth |
| name | VARCHAR | Nama lengkap |
| username | VARCHAR (UNIQUE) | Username, unik |
| email | VARCHAR (UNIQUE) | Email, unik |
| phone | VARCHAR, NULLABLE | Nomor telepon, opsional |
| phone_verified | BOOLEAN | Status verifikasi nomor telepon |
| email_verified | BOOLEAN | Status verifikasi email |
| bio | TEXT, NULLABLE | Bio profil |
| avatar_url | VARCHAR, NULLABLE | URL foto profil di Supabase Storage |
| is_private | BOOLEAN, DEFAULT false | Status akun privat/publik |
| created_at | TIMESTAMP |  |
| updated_at | TIMESTAMP |  |

## 10.2 Tabel: user_interests

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| user_id | UUID (FK -> users.id) |  |
| interest_category | VARCHAR | Contoh: musik, olahraga, kuliner, dst. |

## 10.3 Tabel: follows

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| follower_id | UUID (FK -> users.id) | Pengguna yang melakukan follow |
| following_id | UUID (FK -> users.id) | Pengguna yang di-follow |
| status | ENUM | accepted, pending (untuk akun privat) |
| is_close_friend | BOOLEAN, DEFAULT false | Ditandai oleh following_id terhadap follower_id |
| created_at | TIMESTAMP |  |

Catatan: kombinasi (follower_id, following_id) harus UNIQUE untuk mencegah duplikasi relasi follow.

## 10.4 Tabel: blocks

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| blocker_id | UUID (FK -> users.id) |  |
| blocked_id | UUID (FK -> users.id) |  |
| created_at | TIMESTAMP |  |

## 10.5 Tabel: posts

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| user_id | UUID (FK -> users.id) |  |
| caption | TEXT, NULLABLE |  |
| location | VARCHAR, NULLABLE | [ADV] Geotag |
| is_archived | BOOLEAN, DEFAULT false |  |
| comments_disabled | BOOLEAN, DEFAULT false | [ADV] |
| created_at | TIMESTAMP |  |
| deleted_at | TIMESTAMP, NULLABLE | Soft delete |

## 10.6 Tabel: post_media

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| post_id | UUID (FK -> posts.id) |  |
| media_url | VARCHAR | URL di Supabase Storage |
| media_type | ENUM | image, video |
| order_index | INTEGER | [ADV] Urutan pada carousel |
| music_track_url | VARCHAR, NULLABLE | [ADV] Audio tambahan |

## 10.7 Tabel: post_tags (mention pengguna pada post)

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| post_id | UUID (FK -> posts.id) |  |
| tagged_user_id | UUID (FK -> users.id) |  |
| position_x | FLOAT, NULLABLE | Koordinat tag pada gambar (opsional) |
| position_y | FLOAT, NULLABLE |  |

## 10.8 Tabel: hashtags & post_hashtags

| **Tabel** | **Kolom ****Penting** | **Keterangan** |
| --- | --- | --- |
| hashtags | id (PK), tag (UNIQUE) | Daftar hashtag unik |
| post_hashtags | post_id (FK), hashtag_id (FK) | Tabel relasi many-to-many |

## 10.9 Tabel: likes (polymorphic untuk post & comment)

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| user_id | UUID (FK -> users.id) |  |
| likeable_type | ENUM | post, comment |
| likeable_id | UUID | FK dinamis ke posts.id atau comments.id |
| created_at | TIMESTAMP |  |

## 10.10 Tabel: comments

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| post_id | UUID (FK -> posts.id) |  |
| user_id | UUID (FK -> users.id) |  |
| parent_comment_id | UUID, NULLABLE (FK -> comments.id) | Untuk reply/nested comment |
| content | TEXT |  |
| is_pinned | BOOLEAN, DEFAULT false | [ADV] |
| created_at | TIMESTAMP |  |
| deleted_at | TIMESTAMP, NULLABLE |  |

## 10.11 Tabel: saved_posts

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| user_id | UUID (FK -> users.id) |  |
| post_id | UUID (FK -> posts.id) |  |
| collection_name | VARCHAR, DEFAULT 'All' | [ADV] Nama koleksi tersimpan |
| created_at | TIMESTAMP |  |

## 10.12 Tabel: stories

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| user_id | UUID (FK -> users.id) |  |
| media_url | VARCHAR, NULLABLE | Untuk story foto/video |
| media_type | ENUM | image, video, text |
| text_content | TEXT, NULLABLE | Untuk story berbasis teks |
| music_track_url | VARCHAR, NULLABLE | [ADV] |
| visibility | ENUM, DEFAULT 'all_followers' | all_followers, close_friends [ADV] |
| expires_at | TIMESTAMP | created_at + 24 jam |
| created_at | TIMESTAMP |  |

## 10.13 Tabel: story_views

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| story_id | UUID (FK -> stories.id) |  |
| viewer_id | UUID (FK -> users.id) |  |
| viewed_at | TIMESTAMP |  |

## 10.14 Tabel: story_tags & story_highlights [ADV]

| **Tabel** | **Kolom ****Penting** | **Keterangan** |
| --- | --- | --- |
| story_tags | story_id (FK), tagged_user_id (FK) | Mention pengguna pada story |
| highlights | id (PK), user_id (FK), title | Koleksi highlight permanen |
| highlight_stories | highlight_id (FK), story_id (FK) | Relasi many-to-many highlight & story |

## 10.15 Tabel: conversations & messages

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| conversations.id | UUID (PK) |  |
| conversations.is_group | BOOLEAN, DEFAULT false | [ADV] |
| conversation_participants.conversation_id | UUID (FK) |  |
| conversation_participants.user_id | UUID (FK) |  |
| messages.id | UUID (PK) |  |
| messages.conversation_id | UUID (FK) |  |
| messages.sender_id | UUID (FK -> users.id) |  |
| messages.content | TEXT, NULLABLE | Untuk pesan teks |
| messages.media_url | VARCHAR, NULLABLE | Untuk pesan media |
| messages.reply_to_story_id | UUID, NULLABLE (FK -> stories.id) | Penghubung Story Reply |
| messages.is_read | BOOLEAN, DEFAULT false | [ADV] Read receipt |
| messages.created_at | TIMESTAMP |  |

## 10.16 Tabel: notifications

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| recipient_id | UUID (FK -> users.id) |  |
| actor_id | UUID (FK -> users.id) | Pengguna yang memicu notifikasi |
| type | ENUM | like, comment, follow, follow_request, mention, story_reply |
| reference_id | UUID, NULLABLE | ID post/comment/story terkait |
| is_read | BOOLEAN, DEFAULT false |  |
| created_at | TIMESTAMP |  |

## 10.17 Tabel: reports

| **Kolom** | **Tipe** | **Keterangan** |
| --- | --- | --- |
| id | UUID (PK) |  |
| reporter_id | UUID (FK -> users.id) |  |
| target_type | ENUM | user, post, comment |
| target_id | UUID |  |
| reason | ENUM | spam, inappropriate, harassment, fake_account, other |
| status | ENUM, DEFAULT 'pending' | pending, reviewed, action_taken, dismissed |
| created_at | TIMESTAMP |  |

## 10.18 Relasi Antar Entitas (Ringkasan)

Ringkasan relasi utama antar tabel (notasi: 1-N berarti satu ke banyak, N-N berarti banyak ke banyak):

- users 1-N posts, users 1-N stories, users 1-N comments

- users N-N users (melalui follows) — relasi following/follower

- users N-N users (melalui blocks) — relasi block

- posts 1-N post_media, posts 1-N comments, posts N-N hashtags (melalui post_hashtags)

- posts N-N users (melalui likes dan saved_posts)

- stories 1-N story_views, stories N-N users (melalui story_tags)

- conversations N-N users (melalui conversation_participants), conversations 1-N messages

# 11. Daftar API Endpoint (REST)

Seluruh endpoint menggunakan prefix /api/v1. Autentikasi menggunakan Bearer Token (JWT) pada header Authorization, kecuali endpoint yang ditandai (Publik). Daftar ini bersifat ringkasan tingkat resource; detail request/response body sebaiknya didokumentasikan lebih lanjut menggunakan Postman Collection.

## 11.1 Authentication

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| POST | /auth/register | Registrasi akun baru (Publik) |
| POST | /auth/verify-otp | Verifikasi OTP registrasi (Publik) |
| POST | /auth/login | Login dengan email/username + password (Publik) |
| POST | /auth/refresh-token | Perbarui access token menggunakan refresh token |
| POST | /auth/logout | Logout sesi saat ini |
| POST | /auth/forgot-password | Request OTP untuk reset password (Publik) |
| POST | /auth/reset-password | Set password baru setelah OTP terverifikasi (Publik) |
| POST | /auth/recover-username | Skenario A pada bagian 3.4 (Publik) |
| POST | /auth/recover-email | Skenario B pada bagian 3.4 (Publik) |

## 11.2 Users & Profile

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| GET | /users/me | Profil pengguna saat ini |
| PATCH | /users/me | Update profil (nama, bio, avatar, dll.) |
| GET | /users/:username | Lihat profil pengguna lain |
| PATCH | /users/me/privacy | Toggle akun privat/publik |
| GET | /users/me/interests | Lihat minat yang dipilih |
| PUT | /users/me/interests | Set/update minat pengguna |

## 11.3 Follow & Relasi Sosial

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| POST | /users/:id/follow | Follow pengguna / kirim permintaan follow (jika privat) |
| DELETE | /users/:id/follow | Unfollow pengguna |
| GET | /users/:id/followers | Daftar follower |
| GET | /users/:id/following | Daftar following |
| DELETE | /users/me/followers/:id | Remove follower |
| GET | /follow-requests | Daftar permintaan follow masuk |
| POST | /follow-requests/:id/approve | Setujui permintaan follow |
| POST | /follow-requests/:id/decline | Tolak permintaan follow |
| POST | /users/:id/close-friends | [ADV] Tambah ke Close Friends |
| DELETE | /users/:id/close-friends | [ADV] Hapus dari Close Friends |
| POST | /users/:id/block | Block pengguna |
| DELETE | /users/:id/block | Unblock pengguna |

## 11.4 Posts & Feed

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| GET | /feed | Beranda — daftar post (following + minat) |
| GET | /explore | [ADV] Konten discovery berbasis minat/trending |
| POST | /posts | Buat post baru (upload media + caption) |
| GET | /posts/:id | Detail post |
| PATCH | /posts/:id | Edit caption post |
| DELETE | /posts/:id | Hapus post (soft delete) |
| POST | /posts/:id/archive | Arsipkan post |
| POST | /posts/:id/unarchive | Kembalikan dari arsip |
| GET | /users/:id/posts | Daftar post milik pengguna tertentu |

## 11.5 Interaksi Post (Like, Comment, Save, Share)

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| POST | /posts/:id/like | Like post |
| DELETE | /posts/:id/like | Unlike post |
| GET | /posts/:id/comments | Daftar komentar (dengan nested reply) |
| POST | /posts/:id/comments | Tambah komentar |
| DELETE | /comments/:id | Hapus komentar |
| POST | /comments/:id/like | Like komentar |
| POST | /posts/:id/save | Simpan post |
| DELETE | /posts/:id/save | Hapus dari tersimpan |
| GET | /users/me/saved | Daftar post tersimpan |
| POST | /posts/:id/share | Catat aksi share (untuk analitik) & generate link |

## 11.6 Story

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| POST | /stories | Upload story baru |
| GET | /stories/feed | Daftar story dari following (aktif, belum expired) |
| GET | /stories/:id | Detail story |
| POST | /stories/:id/view | Catat bahwa story telah dilihat |
| GET | /stories/:id/viewers | Daftar penonton story (untuk pemilik) |
| DELETE | /stories/:id | Hapus story |
| POST | /stories/:id/highlight | [ADV] Simpan ke Highlight |

## 11.7 Pencarian

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| GET | /search/users?q= | Cari pengguna berdasarkan username/nama |
| GET | /search/hashtags?q= | Cari hashtag |
| GET | /hashtags/:tag/posts | Daftar post dengan hashtag tertentu |

## 11.8 Direct Message

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| GET | /conversations | Daftar percakapan pengguna |
| POST | /conversations | Mulai percakapan baru |
| GET | /conversations/:id/messages | Riwayat pesan (paginated) |
| POST | /conversations/:id/messages | Kirim pesan baru |
| DELETE | /messages/:id | [ADV] Unsend pesan |

## 11.9 Notifikasi & Report

| **Method** | **Endpoint** | **Keterangan** |
| --- | --- | --- |
| GET | /notifications | Daftar notifikasi pengguna |
| POST | /notifications/:id/read | Tandai sebagai sudah dibaca |
| POST | /reports | Buat report baru terhadap konten/pengguna |

# 12. Arsitektur Sistem

## 12.1 Gambaran Umum Arsitektur

Twistgram menggunakan arsitektur layered monolith pada Fase MVP — satu service Go yang menangani seluruh business logic, terhubung ke Supabase sebagai backend-as-a-service untuk database, auth, storage, dan realtime. Pendekatan ini dipilih karena lebih realistis untuk dikerjakan individu/tim kecil, namun tetap disusun dengan pemisahan layer yang rapi (handler, service, repository) agar mudah dipecah menjadi microservices di masa depan jika diperlukan.

## 12.2 Diagram Lapisan (Layered Architecture)

- Client Layer: React.js (web, prioritas utama), React Native (mobile, osional),

- API Layer (Go): HTTP Handler → menerima request, validasi input dasar

- Service Layer (Go): Business logic — aturan-aturan pada bagian 3.6, 4.7, 5.6, dst diimplementasikan di sini

- Repository Layer (Go): Query ke database, abstraksi akses data

- Data Layer: PostgreSQL (Supabase), Supabase Storage (media), Supabase Realtime (chat & notifikasi live)

## 12.3 Komponen Pendukung

| **Komponen** | **Fungsi** | **Teknologi**** yang ****Disarankan** |
| --- | --- | --- |
| Reverse Proxy / Load Balancer | Routing request, SSL termination | Nginx atau Caddy (opsional, untuk deployment lanjutan) |
| Containerization | Konsistensi environment dev/staging/prod | Docker + Docker Compose |
| CI/CD | Build & test otomatis saat push | GitHub Actions |
| Media Processing | Compress/resize gambar & video sebelum disimpan | Library Go (image/draw) atau ffmpeg untuk video |
| Rate Limiting | Mencegah abuse pada endpoint sensitif (login, OTP) | Middleware Go custom / Redis |
| Caching | [ADV] Mempercepat feed/explore yang sering diakses | Redis |

## 12.4 Pertimbangan Keamanan (Security Considerations)

- Row Level Security (RLS) Supabase diaktifkan pada seluruh tabel agar akses data tetap aman meski terjadi kesalahan logic pada API layer.

- Validasi input pada seluruh endpoint untuk mencegah SQL Injection (mitigasi otomatis jika menggunakan parameterized query/ORM seperti pgx atau sqlc).

- Rate limiting pada endpoint login, register, dan OTP untuk mencegah brute-force.

- Password di-hash menggunakan bcrypt/argon2 (ditangani otomatis oleh Supabase Auth).

- Validasi tipe & ukuran file saat upload media untuk mencegah upload file berbahaya.

- CORS dikonfigurasi secara ketat hanya mengizinkan origin aplikasi resmi.

## 12.5 Pertimbangan Skalabilitas

- Pagination wajib diterapkan pada seluruh endpoint yang mengembalikan list (feed, comments, followers, messages) menggunakan cursor-based pagination untuk performa yang konsisten.

- Index database pada kolom yang sering di-query: users.username, follows(follower_id, following_id), posts.user_id, hashtags.tag.

- Media (gambar/video) disimpan di object storage (Supabase Storage), bukan di database, untuk menjaga performa query.

# 13. Roadmap & Fase Pengembangan

## 13.1 Fase 1 — MVP

Tujuan fase ini: menghasilkan aplikasi yang dapat di-demo end-to-end dengan seluruh flow inti berfungsi nyata (bukan mockup).

**Sprint 1 — ****Fondasi**** (Auth ****&**** ****Profil****)**

- Setup project: Go backend + Supabase + React Native skeleton

- Modul Autentikasi lengkap (register, login, forgot password, recovery)

- Modul Profil & Privasi dasar

**Sprint 2 — ****Relasi**** Sosial ****&**** Feed**

- Follow system + approval untuk akun privat

- Block & Report dasar

- Beranda (feed kronologis + fallback minat)

**Sprint 3 — ****Konten**** (Post)**

- Upload post (gambar/video) + caption + tag + hashtag

- Like, comment (dengan reply 1 level), save, delete, archive

**Sprint 4 — Story ****&**** ****Pencarian**

- Upload & lihat story + auto-expire 24 jam + story viewers

- Story reply via DM

- Pencarian user & hashtag

**Sprint 5 — Chat, ****Notifikasi****, Polish**

- Direct Message dasar (teks + gambar)

- Notifikasi in-app

- Bug fixing, polish UI/UX, deployment ke staging

## 13.2 Fase 2 — Advanced

**Sprint 6 — ****Personalisasi**** ****&**** ****Algoritma**

- Feed ranking algorithm (engagement + relevansi minat)

- Tab Explore/For You

**Sprint 7 — Close Friends ****&**** Story ****Lanjutan**

- Close Friends + story visibility khusus

- Story Highlight

- Musik pada post/story (dengan audio milik pengguna sebagai pengganti lisensi pihak ketiga)

**Sprint 8 — Chat ****&**** Realtime ****Lanjutan**

- Read receipt, typing indicator, status online

- Group chat

**Sprint 9 — ****Keamanan**** ****&**** Observability**

- 2FA, session management

- Dashboard moderasi admin

- Logging & monitoring dasar

## 13.3 Matriks Prioritas Fitur

| **Modul** | **Fase**** MVP** | **Fase**** Advanced** |
| --- | --- | --- |
| Autentikasi | Register, login, forgot password/username (Skenario A&B) | 2FA, OAuth, session management, Skenario C |
| Profil & Relasi | Edit profil, privasi, follow/block/report | Close Friends, badge verifikasi |
| Konten | Post single-media, caption, tag, hashtag, CRUD | Carousel, musik, filter gambar, geotag |
| Story | Upload, view, expire 24 jam, reply via DM | Highlight, tag, musik, sticker interaktif |
| Chat | Teks & gambar 1-ke-1 | Read receipt, typing, group chat, unsend |
| Notifikasi | In-app | Push notification, preferensi notifikasi |
| Discovery | Search user & hashtag | Explore/For You, algoritma ranking |

# 14. Penutup

## 14.1 Catatan Penggunaan Dokumen

Dokumen ini bersifat hidup (living document) — disarankan untuk diperbarui seiring berjalannya pengembangan, terutama saat ditemukan kebutuhan baru atau perubahan keputusan teknis. Setiap perubahan signifikan pada scope sebaiknya dicatat pada riwayat revisi.

## 14.2 Rekomendasi Langkah Selanjutnya

- Buat struktur folder mengikuti layered architecture pada bagian 12.2.

- Setup project Supabase (database, auth, storage) sesuai skema pada bagian 10.

- Buat Postman Collection berdasarkan daftar endpoint pada bagian 11, sebagai kontrak antara frontend dan backend sebelum coding dimulai.

- Mulai implementasi mengikuti urutan Sprint 1-5 pada roadmap Fase MVP.

- Dokumentasikan progres secara berkala (README per fitur, architecture decision records).

## 14.3 Riwayat Revisi

| **Versi** | **Tanggal** | **Keterangan** |
| --- | --- | --- |
| 1.0 | Juni 2026 | Dokumen awal — hasil analisis kebutuhan & gap analysis terhadap aplikasi media sosial existing |

Halaman 5 dari 6