-- +goose Up
-- +goose StatementBegin
-- Add personnel_id column to message_stat if it does not exist
ALTER TABLE "public"."message_stat" ADD COLUMN IF NOT EXISTS "personnel_id" UUID;

-- Make user_id nullable if it exists
ALTER TABLE "public"."message_stat" ALTER COLUMN "user_id" DROP NOT NULL;

-- Populate personnel_id from user_id if any rows exist
UPDATE "public"."message_stat" ms
SET "personnel_id" = p."id"
FROM "public"."personnel" p
WHERE ms."user_id" = p."user_id" AND ms."personnel_id" IS NULL;

-- Add foreign key constraint to personnel
ALTER TABLE "public"."message_stat" DROP CONSTRAINT IF EXISTS "message_stat_personnel_id_fkey";
ALTER TABLE "public"."message_stat" ADD CONSTRAINT "message_stat_personnel_id_fkey" 
    FOREIGN KEY ("personnel_id") REFERENCES "public"."personnel"("id") ON DELETE CASCADE;

-- Add unique constraint on (message_id, personnel_id)
ALTER TABLE "public"."message_stat" DROP CONSTRAINT IF EXISTS "message_stat_message_id_personnel_id_key";
ALTER TABLE "public"."message_stat" ADD CONSTRAINT "message_stat_message_id_personnel_id_key" 
    UNIQUE ("message_id", "personnel_id");

-- Add index on personnel_id
CREATE INDEX IF NOT EXISTS "idx_message_stat_personnel_id" 
    ON "public"."message_stat"("personnel_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS "idx_message_stat_personnel_id";
ALTER TABLE "public"."message_stat" DROP CONSTRAINT IF EXISTS "message_stat_message_id_personnel_id_key";
ALTER TABLE "public"."message_stat" DROP CONSTRAINT IF EXISTS "message_stat_personnel_id_fkey";
ALTER TABLE "public"."message_stat" DROP COLUMN IF EXISTS "personnel_id";
-- +goose StatementEnd
