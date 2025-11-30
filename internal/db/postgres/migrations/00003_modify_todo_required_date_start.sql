-- +goose Up
-- +goose StatementBegin
ALTER TABLE todos
ALTER COLUMN date_start SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE todos
ALTER COLUMN date_start DROP NOT NULL;
-- +goose StatementEnd
