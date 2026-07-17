-- ============================================================
-- Twistgram - Migration 006: Performance Indexes & Search History
-- ============================================================
-- Adds missing composite indexes for frequently queried columns,
-- search history table, and notification filtering.
-- ============================================================

-- 1. Notification filtering index (recipient + unread status)
CREATE INDEX IF NOT EXISTS "idx_notifications_recipient_unread" 
  ON "notifications" ("recipient_id", "is_read");

-- 2. Stories composite index (user + active status)
CREATE INDEX IF NOT EXISTS "idx_stories_user_expires" 
  ON "stories" ("user_id", "expires_at");

-- 3. Messages conversation + created_at (for pagination)
CREATE INDEX IF NOT EXISTS "idx_messages_conv_created" 
  ON "messages" ("conversation_id", "created_at DESC");

-- 4. Follows composite for follower lookup with status
CREATE INDEX IF NOT EXISTS "idx_follows_follower_status" 
  ON "follows" ("follower_id", "status");

-- 5. Follows composite for following lookup with status
CREATE INDEX IF NOT EXISTS "idx_follows_following_status" 
  ON "follows" ("following_id", "status");

-- 6. Close friend lookup
CREATE INDEX IF NOT EXISTS "idx_follows_close_friend" 
  ON "follows" ("following_id", "status", "is_close_friend") 
  WHERE "is_close_friend" = true;

-- 7. Auth OTP cleanup (expired OTPs)
CREATE INDEX IF NOT EXISTS "idx_auth_otps_expires" 
  ON "auth_otps" ("expires_at");

-- 8. Search History table
CREATE TABLE IF NOT EXISTS "search_history" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "query" varchar(255) NOT NULL,
    "query_type" varchar(20) NOT NULL,
    "created_at" timestamptz DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_search_user_query" ON "search_history" ("user_id", "query");
CREATE INDEX IF NOT EXISTS "idx_search_history_user_created" ON "search_history" ("user_id", "created_at DESC");

-- ============================================================
-- END MIGRATION 006
-- ============================================================
