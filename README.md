# Twistgram API Backend (Go)

Twistgram adalah platform media sosial berbasis foto, video, dan teks yang dirancang untuk memungkinkan pengguna terhubung, berbagi momen, dan menemukan konten yang relevan dengan minat masing-masing.

Repository ini memuat kode *Backend API* dari Twistgram yang dibangun menggunakan arsitektur **Layered Monolith** di atas ekosistem bahasa pemrograman **Go (Golang)**.

---

## 🛠 Tech Stack & Prasyarat
- **Bahasa**: Go 1.20+
- **Framework API**: Gin Web Framework (`github.com/gin-gonic/gin`)
- **Database ORM**: GORM (`gorm.io/gorm`)
- **Basis Data**: PostgreSQL
- **Kriptografi**: Bcrypt (`golang.org/x/crypto/bcrypt`) & JWT (`github.com/golang-jwt/jwt/v5`)
- **Live Reload (Dev)**: Air (`github.com/air-verse/air`)

## 📦 Panduan Instalasi & Konfigurasi

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
Salin berkas referensi `env` ke berkas asli yang akan dibaca sistem.
```bash
cp .env.example .env
```
Buka berkas `.env` yang baru dibuat dan isi informasi kredensial yang dibutuhkan, khususnya:
- `DATABASE_URL`: String koneksi PostgreSQL.
- `SUPABASE_JWT_SECRET`: Dibutuhkan oleh middleware API & generator token untuk HMAC-SHA256 signature.
- Konfigurasi `SMTP_*` jika ingin OTP dikirimkan via email (Bila kosong, OTP akan di-print ke Console *stdout*).

### 4. Database Migration & Schema Setup
Aplikasi memiliki kapabilitas `AutoMigrate` melalui GORM saat pertama kali *startup*. Namun, pastikan Anda mengeksekusi DDL script yang berada di dalam folder `migrations/` berurutan pada _SQL Editor_ DB Anda guna menyesuaikan _Triggers_, struktur khusus OTP, dan membuang struktur lama (sinkronisasi GoTrue) jika sebelumnya pernah menggunakan *Supabase Auth External*.

---

## 🚀 Instruksi Menjalankan Aplikasi & Test

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

## 📂 Peta Arsitektur Proyek
Struktur lapisan diadaptasi untuk skalabilitas yang baik (*separation of concerns*):

```text
twistgram-api-go/
├── cmd/api/                  # Titik masuk utama aplikasi (main.go), inisiasi DB & router.
├── docs/                     # Dokumentasi sistem (SRS & TDD).
├── internal/
│   ├── config/               # Struktur env variabel dan konfigurasi aplikasi.
│   ├── dto/                  # Data Transfer Objects (Payload requests & responses).
│   ├── handler/              # HTTP Controller/Handler (Menerima input dari jaringan).
│   ├── middleware/           # Proteksi rute (JWT Auth, Security Headers, Rate Limiter).
│   ├── model/                # Entity database & GORM Tags mapping.
│   ├── repository/           # Lapisan interaksi database (Query GORM & Transaction).
│   └── service/              # Core Business Logic (Verifikasi aturan/batas fitur).
├── migrations/               # Repositori skrip SQL (DDL PostgreSQL).
├── pkg/
│   ├── auth/                 # Utilitas enkripsi Bcrypt, OTP & penandatanganan JWT.
│   ├── mailer/               # Utilitas transporter SMTP.
│   └── response/             # Standardisasi cetak HTTP Response JSON.
└── Twistgram_Postman_Collection.json # Postman Collection lengkap yang siap untuk di-import.
```
