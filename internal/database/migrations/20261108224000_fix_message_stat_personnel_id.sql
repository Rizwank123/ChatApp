-- +goose Up
-- +goose StatementBegin
-- Add personnel_id column to message_stat if it does not exist
ALTER TABLE "public"."message_stat" ADD COLUMN IF NOT EXISTS "personnel_id" UUID;

-- Older schemas used user_id; the current schema already uses personnel_id.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'message_stat'
          AND column_name = 'user_id'
    ) THEN
        EXECUTE 'ALTER TABLE "public"."message_stat" ALTER COLUMN "user_id" DROP NOT NULL';
        EXECUTE 'UPDATE "public"."message_stat" ms
                 SET "personnel_id" = p."id"
                 FROM "public"."personnel" p
                 WHERE ms."user_id" = p."user_id" AND ms."personnel_id" IS NULL';
    END IF;
END
$$;

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
-- +goose StatementEnd
