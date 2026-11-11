-- +goose Up
-- +goose StatementBegin
ALTER TABLE "public"."conversation_participants" ADD COLUMN IF NOT EXISTS "personnel_id" UUID;

-- Populate personnel_id from user_id if present
UPDATE "public"."conversation_participants" cp
SET "personnel_id" = p."id"
FROM "public"."personnel" p
WHERE cp."user_id" = p."user_id" AND cp."personnel_id" IS NULL;

-- Populate user_id from personnel_id if present
UPDATE "public"."conversation_participants" cp
SET "user_id" = p."user_id"
FROM "public"."personnel" p
WHERE cp."personnel_id" = p."id" AND cp."user_id" IS NULL;

CREATE INDEX IF NOT EXISTS "idx_conversation_participants_personnel" ON "public"."conversation_participants"("personnel_id");
CREATE INDEX IF NOT EXISTS "idx_conversation_participants_user" ON "public"."conversation_participants"("user_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS "idx_conversation_participants_personnel";
DROP INDEX IF EXISTS "idx_conversation_participants_user";
ALTER TABLE "public"."conversation_participants" DROP COLUMN IF EXISTS "personnel_id";
-- +goose StatementEnd
