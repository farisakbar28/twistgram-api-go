ALTER TABLE "users"
ADD COLUMN IF NOT EXISTS "external_link" varchar(500);
