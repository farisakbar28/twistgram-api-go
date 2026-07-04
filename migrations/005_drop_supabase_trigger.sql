-- ============================================================
-- Twistgram - Migration: Remove Supabase Auth Trigger
-- ============================================================
-- Mengingat sistem telah menggunakan Self-Hosted Authentication
-- dan menulis langsung ke public.users, trigger sinkronisasi
-- dari skema auth.users (Supabase GoTrue) menjadi usang.
-- Script ini menghapus trigger dan function sinkronisasinya.
-- ============================================================

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
DROP TRIGGER IF EXISTS on_auth_user_updated ON auth.users;

DROP FUNCTION IF EXISTS public.handle_new_user();
DROP FUNCTION IF EXISTS public.handle_auth_user_update();
