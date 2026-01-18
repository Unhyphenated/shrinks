-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_analytics_link_id_clicked_at ON analytics(link_id, clicked_at);
DROP INDEX IF EXISTS idx_analytics_link_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_analytics_link_id_clicked_at;
CREATE INDEX idx_analytics_link_id ON analytics(link_id);
-- +goose StatementEnd
