-- ============================================================
-- Twistgram - Migration 007: Token Version for AUTH-05
-- ============================================================
-- Adds token_version column to users table for session invalidation
-- on password change. All existing tokens remain valid until refresh.
-- ============================================================

ALTER TABLE public.users ADD COLUMN IF NOT EXISTS "token_version" integer DEFAULT 1;

-- ============================================================
-- END MIGRATION 007
-- ============================================================
