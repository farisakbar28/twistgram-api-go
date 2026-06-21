-- ============================================================
-- Twistgram - Sync Trigger: auth.users → public.users
-- ============================================================
-- Jalankan file ini di Supabase SQL Editor SETELAH migration
-- 001_schema.sql sukses.
--
-- Trigger ini memastikan setiap user yang register via Supabase
-- Auth otomatis memiliki row di tabel public.users dan menjaga
-- email/email_verified tetap sinkron saat auth.users berubah.
-- ============================================================

CREATE OR REPLACE FUNCTION public.twistgram_normalize_username(value text)
RETURNS text
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT trim(both '_' from regexp_replace(lower(COALESCE(NULLIF(value, ''), 'user')), '[^a-z0-9_]+', '_', 'g'));
$$;

CREATE OR REPLACE FUNCTION public.twistgram_unique_username(base_username text, user_id uuid)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  normalized text;
  candidate text;
  suffix text;
  counter integer := 0;
BEGIN
  normalized := public.twistgram_normalize_username(base_username);

  IF normalized IS NULL OR normalized = '' THEN
    normalized := 'user';
  END IF;

  suffix := '_' || replace(user_id::text, '-', '');
  candidate := left(normalized, 255);

  WHILE EXISTS (
    SELECT 1 FROM public.users WHERE username = candidate AND id <> user_id
  ) LOOP
    counter := counter + 1;

    IF counter = 1 THEN
      candidate := left(normalized, greatest(1, 255 - length(suffix))) || suffix;
    ELSE
      candidate := left(normalized, greatest(1, 255 - length(suffix) - length(counter::text) - 1)) || suffix || '_' || counter::text;
    END IF;
  END LOOP;

  RETURN candidate;
END;
$$;

CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  default_name text;
  default_username text;
BEGIN
  default_name := COALESCE(
    NULLIF(NEW.raw_user_meta_data->>'name', ''),
    NULLIF(NEW.raw_user_meta_data->>'full_name', ''),
    NULLIF(split_part(NEW.email, '@', 1), ''),
    'User'
  );

  default_username := public.twistgram_unique_username(
    COALESCE(
      NULLIF(NEW.raw_user_meta_data->>'username', ''),
      NULLIF(split_part(NEW.email, '@', 1), ''),
      'user'
    ),
    NEW.id
  );

  INSERT INTO public.users (id, name, username, email, email_verified)
  VALUES (
    NEW.id,
    default_name,
    default_username,
    NEW.email,
    NEW.email_confirmed_at IS NOT NULL
  )
  ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    email_verified = EXCLUDED.email_verified,
    updated_at = now();

  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.handle_auth_user_update()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = public
AS $$
DECLARE
  default_name text;
  default_username text;
BEGIN
  UPDATE public.users
  SET
    email = NEW.email,
    email_verified = NEW.email_confirmed_at IS NOT NULL,
    updated_at = now()
  WHERE id = NEW.id;

  IF NOT FOUND THEN
    default_name := COALESCE(
      NULLIF(NEW.raw_user_meta_data->>'name', ''),
      NULLIF(NEW.raw_user_meta_data->>'full_name', ''),
      NULLIF(split_part(NEW.email, '@', 1), ''),
      'User'
    );

    default_username := public.twistgram_unique_username(
      COALESCE(
        NULLIF(NEW.raw_user_meta_data->>'username', ''),
        NULLIF(split_part(NEW.email, '@', 1), ''),
        'user'
      ),
      NEW.id
    );

    INSERT INTO public.users (id, name, username, email, email_verified)
    VALUES (
      NEW.id,
      default_name,
      default_username,
      NEW.email,
      NEW.email_confirmed_at IS NOT NULL
    )
    ON CONFLICT (id) DO NOTHING;
  END IF;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW
  EXECUTE FUNCTION public.handle_new_user();

DROP TRIGGER IF EXISTS on_auth_user_updated ON auth.users;
CREATE TRIGGER on_auth_user_updated
  AFTER UPDATE OF email, email_confirmed_at, raw_user_meta_data ON auth.users
  FOR EACH ROW
  EXECUTE FUNCTION public.handle_auth_user_update();

-- Optional one-time backfill for existing auth.users without public.users rows.
-- It intentionally uses the same trigger function for collision-safe usernames.
DO $$
DECLARE
  auth_user auth.users%ROWTYPE;
BEGIN
  FOR auth_user IN
    SELECT au.*
    FROM auth.users au
    LEFT JOIN public.users pu ON pu.id = au.id
    WHERE pu.id IS NULL
  LOOP
    INSERT INTO public.users (id, name, username, email, email_verified)
    VALUES (
      auth_user.id,
      COALESCE(
        NULLIF(auth_user.raw_user_meta_data->>'name', ''),
        NULLIF(auth_user.raw_user_meta_data->>'full_name', ''),
        NULLIF(split_part(auth_user.email, '@', 1), ''),
        'User'
      ),
      public.twistgram_unique_username(
        COALESCE(
          NULLIF(auth_user.raw_user_meta_data->>'username', ''),
          NULLIF(split_part(auth_user.email, '@', 1), ''),
          'user'
        ),
        auth_user.id
      ),
      auth_user.email,
      auth_user.email_confirmed_at IS NOT NULL
    )
    ON CONFLICT (id) DO NOTHING;
  END LOOP;
END;
$$;

-- ============================================================
-- END TRIGGER
-- ============================================================
