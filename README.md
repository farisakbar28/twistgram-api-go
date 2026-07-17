# Twistgram API Backend (Go)

Twistgram adalah platform media sosial berbasis foto, video, dan teks yang dirancang untuk memungkinkan pengguna terhubung, berbagi momen, dan menemukan konten yang relevan dengan minat masing-masing.

Repository ini memuat kode *Backend API* dari Twistgram yang dibangun menggunakan arsitektur **Layered Monolith** di atas ekosistem bahasa pemrograman **Go (Golang)**.

---

## Tech Stack & Prasyarat
- **Bahasa**: Go 1.25+
- **Framework API**: Gin Web Framework (`github.com/gin-gonic/gin`)
- **Database ORM**: GORM (`gorm.io/gorm`)
- **Basis Data**: PostgreSQL
- **Kriptografi**: Bcrypt (`golang.org/x/crypto/bcrypt`) & JWT (`github.com/golang-jwt/jwt/v5`)
- **Rate Limiting**: Token bucket per IP (`golang.org/x/time/rate`)
- **Live Reload (Dev)**: Air (`github.com/air-verse/air`)

## Panduan Instalasi & Konfigurasi

### 1. Kloning Repository
```bash
git clone https://github.com/farisakbar28/twistgram-api-go.git
cd twistgram-api-go
```

### 2. Instalasi Dependencies
```bash
go mod tidy
```

### 3. Konfigurasi Environment Variables
Salin berkas referensi `.env.example` ke berkas `.env` yang akan dibaca sistem.
```bash
cp .env.example .env
```
Buka berkas `.env` yang baru dibuat dan isi informasi kredensial yang dibutuhkan:
- `DATABASE_URL`: String koneksi PostgreSQL.
- `SUPABASE_JWT_SECRET`: Dibutuhkan oleh middleware API & generator token untuk HMAC-SHA256 signature.
- Konfigurasi `SMTP_*` jika ingin OTP dikirimkan via email (Bila kosong, OTP akan di-print ke Console *stdout*).

### 4. Database Migration & Schema Setup
Aplikasi memiliki kapabilitas `AutoMigrate` melalui GORM saat pertama kali *startup*. Selain itu, eksekusi DDL script yang berada di dalam folder `migrations/` berurutan pada _SQL Editor_ DB Anda:

1. `001_schema.sql` — Skema utama (users, posts, follows, dll)
2. `002_auth_sync.sql` — **LEGACY** (hanya untuk migrasi dari Supabase GoTrue, skip jika self-hosted auth)
3. `003_add_user_external_link.sql` — Kolom `external_link` pada users
4. `004_self_hosted_auth.sql` — Tabel `auth_otps` dan kolom `password_hash`
5. `005_drop_supabase_trigger.sql` — Hapus trigger Supabase Auth (setelah self-hosted)
6. `006_performance_indexes.sql` — Index performa dan tabel `search_history`

---

## Instruksi Menjalankan Aplikasi & Test

Kami menyediakan *helper script* (`run.sh`) untuk mempermudah operasional CLI:

**1. Menjalankan Server (Development dengan Live-Reload)**
*Membutuhkan `Air` terinstal secara global di mesin Anda.*
```bash
./run.sh start
```

**2. Menjalankan Server (Standar / Production)**
```bash
./run.sh run
# ATAU untuk proses build manual:
./run.sh build
./twistgram-api
```
Server secara default akan mendengarkan _port_ `8080` (`http://localhost:8080`).

**3. Eksekusi Pengujian Otomatis (Unit Test)**
```bash
./run.sh test
```

---

## Daftar API Endpoint

### Authentication (Public)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| POST | `/api/v1/auth/register` | Registrasi akun baru |
| POST | `/api/v1/auth/verify-otp` | Verifikasi OTP registrasi |
| POST | `/api/v1/auth/login` | Login email/username + password |
| POST | `/api/v1/auth/refresh-token` | Perbarui access token |
| POST | `/api/v1/auth/logout` | Logout |
| POST | `/api/v1/auth/forgot-password` | Request OTP reset password |
| POST | `/api/v1/auth/reset-password` | Set password baru |
| POST | `/api/v1/auth/recover-username` | Pemulihan username |
| POST | `/api/v1/auth/recover-email` | Pemulihan email (via OTP) |
| POST | `/api/v1/auth/recover-email/complete` | Selesaikan pemulihan email |
| POST | `/api/v1/auth/resend-otp` | Kirim ulang OTP |
| GET | `/api/v1/auth/check-availability` | Cek ketersediaan username/email |

### Users & Profile (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/api/v1/users/me` | Profil pengguna saat ini |
| PATCH | `/api/v1/users/me` | Update profil |
| DELETE | `/api/v1/users/me` | Hapus akun |
| PATCH | `/api/v1/users/me/privacy` | Toggle privat/publik |
| GET | `/api/v1/users/me/interests` | Lihat minat |
| PUT | `/api/v1/users/me/interests` | Set minat |
| GET | `/api/v1/users/:identifier` | Profil pengguna lain |

### Follow & Social (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| POST | `/api/v1/users/:identifier/follow` | Follow / kirim permintaan |
| DELETE | `/api/v1/users/:identifier/follow` | Unfollow |
| GET | `/api/v1/users/:identifier/followers` | Daftar follower |
| GET | `/api/v1/users/:identifier/following` | Daftar following |
| DELETE | `/api/v1/users/:identifier/followers` | Remove follower |
| GET | `/api/v1/users/me/follow-requests` | Daftar permintaan follow |
| POST | `/api/v1/users/:identifier/follow-requests/approve` | Setujui follow request |
| POST | `/api/v1/users/:identifier/follow-requests/decline` | Tolak follow request |
| POST | `/api/v1/users/:identifier/close-friends` | Tambah close friend |
| DELETE | `/api/v1/users/:identifier/close-friends` | Hapus close friend |
| GET | `/api/v1/users/me/close-friends` | Daftar close friends |
| POST | `/api/v1/users/:identifier/block` | Block pengguna |
| DELETE | `/api/v1/users/:identifier/block` | Unblock pengguna |
| GET | `/api/v1/users/me/blocked` | Daftar blocked users |
| POST | `/api/v1/reports` | Buat report |

### Posts & Feed (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/api/v1/feed` | Beranda (following + fallback global) |
| POST | `/api/v1/posts` | Buat post baru |
| PATCH | `/api/v1/posts/:id` | Edit caption |
| DELETE | `/api/v1/posts/:id` | Hapus post (soft delete) |
| POST | `/api/v1/posts/:id/archive` | Arsipkan post |
| POST | `/api/v1/posts/:id/unarchive` | Kembalikan dari arsip |
| GET | `/api/v1/users/me/posts` | Daftar post milik saya |
| DELETE | `/api/v1/posts/:id/tags/:taggedUserId` | Hapus tag |

### Interactions (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| POST | `/api/v1/posts/:id/like` | Like post |
| DELETE | `/api/v1/posts/:id/like` | Unlike post |
| GET | `/api/v1/posts/:id/comments` | Daftar komentar |
| POST | `/api/v1/posts/:id/comments` | Tambah komentar |
| DELETE | `/api/v1/posts/:id/comments/:comment_id` | Hapus komentar |
| POST | `/api/v1/posts/:id/comments/:comment_id/like` | Like komentar |
| GET | `/api/v1/users/me/saved` | Daftar tersimpan |
| POST | `/api/v1/posts/:id/save` | Simpan post |
| DELETE | `/api/v1/posts/:id/save` | Hapus dari tersimpan |
| POST | `/api/v1/posts/:id/share` | Share post |

### Stories (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| POST | `/api/v1/stories` | Upload story |
| GET | `/api/v1/stories/feed` | Daftar story following |
| GET | `/api/v1/stories/:id` | Detail story |
| POST | `/api/v1/stories/:id/views` | Catat view |
| GET | `/api/v1/stories/:id/viewers` | Daftar penonton |
| DELETE | `/api/v1/stories/:id` | Hapus story |

### Highlights (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/api/v1/highlights` | Daftar highlights saya |
| POST | `/api/v1/highlights` | Buat highlight baru |
| PATCH | `/api/v1/highlights/:id` | Edit judul highlight |
| DELETE | `/api/v1/highlights/:id` | Hapus highlight |
| POST | `/api/v1/highlights/:id/stories` | Tambah story ke highlight |
| DELETE | `/api/v1/highlights/:id/stories/:story_id` | Hapus story dari highlight |

### Search & History (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/api/v1/search?q=` | Cari user & hashtag |
| GET | `/api/v1/hashtags/:tag/posts` | Post dengan hashtag |
| GET | `/api/v1/search/history` | Riwayat pencarian |
| POST | `/api/v1/search/history?q=&type=` | Simpan pencarian |
| DELETE | `/api/v1/search/history/:id` | Hapus satu riwayat |
| DELETE | `/api/v1/search/history` | Hapus semua riwayat |

### Direct Messages (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/api/v1/conversations` | Daftar percakapan |
| POST | `/api/v1/conversations` | Mulai percakapan |
| GET | `/api/v1/conversations/requests` | Permintaan pesan (private accounts) |
| GET | `/api/v1/conversations/:id/messages` | Riwayat pesan |
| POST | `/api/v1/conversations/:id/messages` | Kirim pesan |

### Notifications (Protected)
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/api/v1/notifications` | Daftar notifikasi |
| POST | `/api/v1/notifications/:id/read` | Tandai sudah dibaca |
| POST | `/api/v1/notifications/read-all` | Tandai semua sudah dibaca |
| POST | `/api/v1/notifications` | Buat notifikasi manual |

### System
| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/health` | Health check (public) |

---

## Peta Arsitektur Proyek

```text
twistgram-api-go/
├── cmd/api/                  # Titik masuk utama (main.go)
├── docs/                     # Dokumentasi (SRS & TDD)
├── internal/
│   ├── config/               # Konfigurasi aplikasi
│   ├── constants/            # Konstanta domain (status, tipe, dll)
│   ├── dto/                  # Data Transfer Objects
│   ├── handler/              # HTTP Controllers
│   ├── middleware/           # JWT Auth, Security Headers, Rate Limiter, Request ID
│   ├── model/                # Entity database & GORM Tags
│   ├── repository/           # Database queries & transactions
│   └── service/              # Core Business Logic
├── migrations/               # SQL DDL scripts (001-006)
├── pkg/
│   ├── auth/                 # Bcrypt, OTP, JWT utilities
│   ├── mailer/               # SMTP email sender
│   └── response/             # Standard HTTP response JSON
└── Twistgram_Postman_Collection.json
```

---

## Business Rules Implemented

| Code | Rule | Status |
|------|------|--------|
| AUTH-01 | Password min 8 chars, uppercase at start | Implemented |
| AUTH-02 | Unverified accounts limited to onboarding | Implemented |
| AUTH-03 | OTP valid 10 min, single use | Implemented |
| AUTH-04 | Rate limiting on auth endpoints | Implemented |
| AUTH-05 | Password change invalidates sessions | Partial (stateless JWT) |
| SOC-01 | Private posts hidden from search/explore | Implemented |
| SOC-02 | Block is mutual | Implemented |
| SOC-03 | Close friends limited to followers | Implemented |
| SOC-04 | Follow requests can be declined | Implemented |
| SOC-05 | Username change once per month | Implemented |
| CNT-01 | Story expires after 24h | Implemented |
| CNT-02 | Story reply becomes DM | Implemented |
| CNT-03 | Archived posts visible to owner only | Implemented |
| CNT-04 | Only owner can modify post | Implemented |
| CNT-05 | Tag notification to tagged user | Implemented |
| SRCH-01 | Blocked users hidden from search | Implemented |
| SRCH-02 | Private posts hidden in hashtag search | Implemented |
| MSG-01 | Blocked users cannot message | Implemented |
| MSG-02 | Private account DM requests separated | Implemented |
| NTF-01 | No self-notifications | Implemented |
| NTF-02 | Blocked user notifications hidden | Implemented |
