
-- +goose Up
-- +goose StatementBegin
CREATE TABLE tg_images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT PRIMARY KEY,
    tg_file_id TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tg_images;
-- +goose StatementEnd
