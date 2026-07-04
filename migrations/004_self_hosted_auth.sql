-- ============================================================
-- Twistgram - Migration: Self-Hosted Auth Support
-- ============================================================
-- File ini menambahkan struktur database yang diperlukan untuk
-- modul Autentikasi internal (tanpa Supabase GoTrue).
-- ============================================================

-- Tambahkan kolom password_hash untuk menyimpan enkripsi bcrypt
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Tabel baru untuk menyimpan kode OTP (pendaftaran & pemulihan)
CREATE TABLE IF NOT EXISTS public.auth_otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    code VARCHAR(6) NOT NULL,
    type VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Indeks komposit untuk mempercepat query pencarian OTP spesifik user
CREATE INDEX IF NOT EXISTS idx_auth_otps_user_type ON public.auth_otps(user_id, type);
