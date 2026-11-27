-- +goose Up
-- +goose StatementBegin
ALTER TABLE "public"."conversation_participants" ADD COLUMN IF NOT EXISTS "personnel_id" UUID;

-- Some earlier schemas used user_id while the current schema uses personnel_id.
-- Keep the migration compatible with either shape.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'conversation_participants'
          AND column_name = 'user_id'
    ) THEN
        EXECUTE 'UPDATE "public"."conversation_participants" cp
                 SET "personnel_id" = p."id"
                 FROM "public"."personnel" p
                 WHERE cp."user_id" = p."user_id" AND cp."personnel_id" IS NULL';

        EXECUTE 'UPDATE "public"."conversation_participants" cp
                 SET "user_id" = p."user_id"
                 FROM "public"."personnel" p
                 WHERE cp."personnel_id" = p."id" AND cp."user_id" IS NULL';

        EXECUTE 'CREATE INDEX IF NOT EXISTS "idx_conversation_participants_user"
                 ON "public"."conversation_participants"("user_id")';
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS "idx_conversation_participants_personnel" ON "public"."conversation_participants"("personnel_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS "idx_conversation_participants_personnel";
DROP INDEX IF EXISTS "idx_conversation_participants_user";
-- +goose StatementEnd
