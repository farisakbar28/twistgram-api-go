-- ============================================================
-- Twistgram Database Schema — Generated from GORM AutoMigrate
-- Referensi: SRS Bagian 10 (Desain Basis Data)
-- ============================================================
-- Jalankan file ini di Supabase SQL Editor jika tidak menggunakan
-- AutoMigrate dari Go backend.
-- ============================================================

-- 1. users
CREATE TABLE "users" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "name" varchar(255) NOT NULL,
    "username" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL,
    "phone" varchar(50),
    "phone_verified" boolean DEFAULT false,
    "email_verified" boolean DEFAULT false,
    "bio" text,
    "avatar_url" varchar(500),
    "is_private" boolean DEFAULT false,
    "last_username_at" timestamptz,
    "created_at" timestamptz,
    "updated_at" timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_username" ON "users" ("username");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");

-- 2. user_interests
CREATE TABLE "user_interests" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "users"("id"),
    "interest_category" varchar(100) NOT NULL,
    "created_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_user_interests_user_id" ON "user_interests" ("user_id");

-- 3. follows
CREATE TABLE "follows" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "follower_id" uuid NOT NULL REFERENCES "users"("id"),
    "following_id" uuid NOT NULL REFERENCES "users"("id"),
    "status" varchar(20) NOT NULL DEFAULT 'accepted',   -- accepted, pending
    "is_close_friend" boolean DEFAULT false,           -- [ADV]
    "created_at" timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_follow_pair" ON "follows" ("follower_id","following_id");

-- 4. blocks
CREATE TABLE "blocks" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "blocker_id" uuid NOT NULL REFERENCES "users"("id"),
    "blocked_id" uuid NOT NULL REFERENCES "users"("id"),
    "created_at" timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_block_pair" ON "blocks" ("blocker_id","blocked_id");

-- 5. posts
CREATE TABLE "posts" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "users"("id"),
    "caption" text,
    "location" varchar(255),                 -- [ADV]
    "is_archived" boolean DEFAULT false,
    "comments_disabled" boolean DEFAULT false, -- [ADV]
    "created_at" timestamptz,
    "deleted_at" timestamptz,
    "updated_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_posts_user_id" ON "posts" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_posts_deleted_at" ON "posts" ("deleted_at");

-- 6. post_media
CREATE TABLE "post_media" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "post_id" uuid NOT NULL REFERENCES "posts"("id"),
    "media_url" varchar(500) NOT NULL,
    "media_type" varchar(20) NOT NULL,       -- image, video
    "order_index" bigint DEFAULT 0,          -- [ADV]
    "music_track_url" varchar(500)           -- [ADV]
);
CREATE INDEX IF NOT EXISTS "idx_post_media_post_id" ON "post_media" ("post_id");

-- 7. post_tags
CREATE TABLE "post_tags" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "post_id" uuid NOT NULL REFERENCES "posts"("id"),
    "tagged_user_id" uuid NOT NULL REFERENCES "users"("id"),
    "position_x" decimal,                    -- [ADV]
    "position_y" decimal                     -- [ADV]
);
CREATE INDEX IF NOT EXISTS "idx_post_tags_post_id" ON "post_tags" ("post_id");
CREATE INDEX IF NOT EXISTS "idx_post_tags_tagged_user_id" ON "post_tags" ("tagged_user_id");

-- 8. hashtags
CREATE TABLE "hashtags" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "tag" varchar(100) NOT NULL,
    "created_at" timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_hashtags_tag" ON "hashtags" ("tag");

-- 9. post_hashtags (many-to-many)
CREATE TABLE "post_hashtags" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "post_id" uuid NOT NULL REFERENCES "posts"("id"),
    "hashtag_id" uuid NOT NULL REFERENCES "hashtags"("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_post_hashtag" ON "post_hashtags" ("post_id","hashtag_id");

-- 10. likes (polymorphic: post, comment)
CREATE TABLE "likes" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "users"("id"),
    "likeable_type" varchar(20) NOT NULL,    -- post, comment
    "likeable_id" uuid NOT NULL,
    "created_at" timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_like_unique" ON "likes" ("user_id","likeable_type","likeable_id");

-- 11. comments
CREATE TABLE "comments" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "post_id" uuid NOT NULL REFERENCES "posts"("id"),
    "user_id" uuid NOT NULL REFERENCES "users"("id"),
    "parent_comment_id" uuid REFERENCES "comments"("id"),
    "content" text NOT NULL,
    "is_pinned" boolean DEFAULT false,       -- [ADV]
    "created_at" timestamptz,
    "deleted_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_comments_post_id" ON "comments" ("post_id");
CREATE INDEX IF NOT EXISTS "idx_comments_user_id" ON "comments" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_comments_parent_comment_id" ON "comments" ("parent_comment_id");
CREATE INDEX IF NOT EXISTS "idx_comments_deleted_at" ON "comments" ("deleted_at");

-- 12. saved_posts
CREATE TABLE "saved_posts" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "users"("id"),
    "post_id" uuid NOT NULL REFERENCES "posts"("id"),
    "collection_name" varchar(100) DEFAULT 'All', -- [ADV]
    "created_at" timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_saved_unique" ON "saved_posts" ("user_id","post_id");

-- 13. stories
CREATE TABLE "stories" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "users"("id"),
    "media_url" varchar(500),
    "media_type" varchar(20) NOT NULL DEFAULT 'text',  -- image, video, text
    "text_content" text,
    "music_track_url" varchar(500),                   -- [ADV]
    "visibility" varchar(20) NOT NULL DEFAULT 'all_followers', -- [ADV] all_followers, close_friends
    "expires_at" timestamptz NOT NULL,
    "created_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_stories_user_id" ON "stories" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_stories_expires_at" ON "stories" ("expires_at");

-- 14. story_views
CREATE TABLE "story_views" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "story_id" uuid NOT NULL REFERENCES "stories"("id"),
    "viewer_id" uuid NOT NULL REFERENCES "users"("id"),
    "viewed_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_story_views_story_id" ON "story_views" ("story_id");
CREATE INDEX IF NOT EXISTS "idx_story_views_viewer_id" ON "story_views" ("viewer_id");

-- 15. story_tags [ADV]
CREATE TABLE "story_tags" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "story_id" uuid NOT NULL REFERENCES "stories"("id"),
    "tagged_user_id" uuid NOT NULL REFERENCES "users"("id")
);
CREATE INDEX IF NOT EXISTS "idx_story_tags_story_id" ON "story_tags" ("story_id");
CREATE INDEX IF NOT EXISTS "idx_story_tags_tagged_user_id" ON "story_tags" ("tagged_user_id");

-- 16. highlights [ADV]
CREATE TABLE "highlights" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "users"("id"),
    "title" varchar(100) NOT NULL,
    "created_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_highlights_user_id" ON "highlights" ("user_id");

-- 17. highlight_stories [ADV]
CREATE TABLE "highlight_stories" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "highlight_id" uuid NOT NULL REFERENCES "highlights"("id"),
    "story_id" uuid NOT NULL REFERENCES "stories"("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_highlight_story" ON "highlight_stories" ("highlight_id","story_id");

-- 18. conversations
CREATE TABLE "conversations" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "is_group" boolean DEFAULT false,        -- [ADV]
    "created_at" timestamptz
);

-- 19. conversation_participants
CREATE TABLE "conversation_participants" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "conversation_id" uuid NOT NULL REFERENCES "conversations"("id"),
    "user_id" uuid NOT NULL REFERENCES "users"("id")
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_cnv_part" ON "conversation_participants" ("conversation_id","user_id");

-- 20. messages
CREATE TABLE "messages" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "conversation_id" uuid NOT NULL REFERENCES "conversations"("id"),
    "sender_id" uuid NOT NULL REFERENCES "users"("id"),
    "content" text,
    "media_url" varchar(500),
    "reply_to_story_id" uuid REFERENCES "stories"("id"),
    "is_read" boolean DEFAULT false,         -- [ADV]
    "created_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_messages_conversation_id" ON "messages" ("conversation_id");
CREATE INDEX IF NOT EXISTS "idx_messages_sender_id" ON "messages" ("sender_id");
CREATE INDEX IF NOT EXISTS "idx_messages_reply_to_story_id" ON "messages" ("reply_to_story_id");

-- 21. notifications
CREATE TABLE "notifications" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "recipient_id" uuid NOT NULL REFERENCES "users"("id"),
    "actor_id" uuid NOT NULL REFERENCES "users"("id"),
    "type" varchar(30) NOT NULL,             -- like, comment, follow, follow_request, mention, story_reply
    "reference_id" uuid,
    "is_read" boolean DEFAULT false,
    "created_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_notifications_recipient_id" ON "notifications" ("recipient_id");
CREATE INDEX IF NOT EXISTS "idx_notifications_actor_id" ON "notifications" ("actor_id");

-- 22. reports
CREATE TABLE "reports" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "reporter_id" uuid NOT NULL REFERENCES "users"("id"),
    "target_type" varchar(20) NOT NULL,      -- user, post, comment
    "target_id" uuid NOT NULL,
    "reason" varchar(30) NOT NULL,           -- spam, inappropriate, harassment, fake_account, other
    "status" varchar(20) NOT NULL DEFAULT 'pending', -- pending, reviewed, action_taken, dismissed
    "created_at" timestamptz
);
CREATE INDEX IF NOT EXISTS "idx_reports_reporter_id" ON "reports" ("reporter_id");

-- ============================================================
-- END SCHEMA
-- ============================================================
