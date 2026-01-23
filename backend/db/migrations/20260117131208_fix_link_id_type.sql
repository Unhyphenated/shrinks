-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN id TYPE BIGINT;
ALTER TABLE links ALTER COLUMN id TYPE BIGINT;
ALTER TABLE links ALTER COLUMN user_id TYPE BIGINT;
ALTER TABLE analytics ALTER COLUMN link_id TYPE BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE analytics ALTER COLUMN link_id TYPE INTEGER;
ALTER TABLE links ALTER COLUMN user_id TYPE INTEGER;
ALTER TABLE links ALTER COLUMN id TYPE INTEGER;
ALTER TABLE users ALTER COLUMN id TYPE INTEGER;
-- +goose StatementEnd