-- +goose Up
-- +goose StatementBegin

-- 1. Ensure user_name is lowercased and add lowercased unique index
UPDATE "public"."users" SET "user_name" = LOWER(TRIM("user_name")) WHERE "user_name" IS NOT NULL;

-- Create unique index on LOWER(user_name)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower_unique ON "public"."users" (LOWER(TRIM("user_name")));

-- 2. Add display_name to personnel if not already exists
ALTER TABLE "public"."personnel" ADD COLUMN IF NOT EXISTS "display_name" VARCHAR(255);

-- Populate display_name from first_name and last_name if empty
UPDATE "public"."personnel"
SET "display_name" = TRIM(CONCAT("first_name", ' ', "last_name"))
WHERE "display_name" IS NULL OR "display_name" = '';

-- 3. Add index on personnel display_name and user_id
CREATE INDEX IF NOT EXISTS idx_personnel_user_id ON "public"."personnel"("user_id");
CREATE INDEX IF NOT EXISTS idx_personnel_display_name ON "public"."personnel"("display_name");

-- 4. Ensure composite unique index on follows table for active follows
CREATE UNIQUE INDEX IF NOT EXISTS idx_follows_active_pair ON "public"."follows" ("follower_id", "following_id") WHERE "deleted_at" IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_follows_active_pair;
DROP INDEX IF EXISTS idx_personnel_display_name;
DROP INDEX IF EXISTS idx_users_username_lower_unique;
ALTER TABLE "public"."personnel" DROP COLUMN IF EXISTS "display_name";
-- +goose StatementEnd
