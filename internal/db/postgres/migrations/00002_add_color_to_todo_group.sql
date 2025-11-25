-- +goose Up
-- +goose StatementBegin
ALTER TABLE todo_groups
ADD COLUMN color VARCHAR(20) NOT NULL DEFAULT '#1c7ed6';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE todo_groups
DROP COLUMN IF EXISTS color;
-- +goose StatementEnd
