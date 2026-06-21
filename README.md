# Twistgram API

Backend API untuk aplikasi media sosial Twistgram, dibangun dengan Go (Gin) + GORM + PostgreSQL (Supabase).

## Tech Stack

- **Runtime:** Go 1.26
- **Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL via Supabase
- **Auth:** Supabase Auth (JWT validation)
- **Storage:** Supabase Storage

## Struktur Folder

```
├── cmd/api/            # Entrypoint aplikasi
├── internal/
│   ├── config/         # Konfigurasi & koneksi database
│   ├── handler/        # HTTP handler (controller)
│   ├── service/        # Business logic
│   ├── repository/     # Akses data via GORM
│   ├── model/          # Model/GORM struct
│   ├── middleware/     # Middleware (auth, CORS, dll)
│   └── dto/            # Request/response DTO
├── pkg/response/       # Format response konsisten
├── migrations/         # SQL migration files
├── .env.example        # Template environment variables
└── go.mod / go.sum
```

## Cara Install & Run

### 1. Prerequisites

- Go 1.26+
- PostgreSQL (via Supabase)

### 2. Clone & Setup

```bash
git clone <repo-url>
cd twistgram-api-go
```

### 3. Environment Variables

Salin `.env.example` ke `.env` dan isi kredensial:

```bash
cp .env.example .env
```

Isi variabel berikut di `.env`:

| Variable              | Description                          |
|-----------------------|--------------------------------------|
| `DATABASE_URL`        | Connection string ke Supabase Postgres (session pooler) |
| `SUPABASE_URL`        | URL project Supabase                 |
| `SUPABASE_JWT_SECRET` | JWT secret untuk validasi token      |
| `PORT`                | Port server (default: 8080)          |

### 4. Run

```bash
go run cmd/api/main.go
```

Server akan berjalan di `http://localhost:8080`.

### 5. Verifikasi

```bash
curl http://localhost:8080/health
```

Response sukses:
```json
{
  "success": true,
  "message": "Success",
  "data": {
    "status": "ok",
    "database": "connected",
    "timestamp": "2026-06-21T21:30:00+08:00"
  }
}
```

## API Documentation

Dokumentasi API lengkap tersedia di Postman Collection (lihat folder `migrations/` untuk referensi, koleksi Postman akan ditambahkan di fase akhir).

## Development

### Menambahkan Dependencies

```bash
go get <package-name>
go mod tidy
```

### Build

```bash
go build -o bin/api cmd/api/main.go
```

## License

Proyek ini dikembangkan sebagai portofolio.
