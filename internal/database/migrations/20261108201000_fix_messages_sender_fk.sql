-- +goose Up
-- +goose StatementBegin
ALTER TABLE "public"."messages" DROP CONSTRAINT IF EXISTS "messages_sender_id_fkey";
ALTER TABLE "public"."messages" ADD CONSTRAINT "messages_sender_id_fkey" FOREIGN KEY ("sender_id") REFERENCES "public"."personnel"("id") ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
