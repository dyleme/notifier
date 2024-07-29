-- +goose Up
-- +goose StatementBegin
CREATE TABLE tg_images
(
    id SERIAL PRIMARY KEY,
    filename VARCHAR(250) NOT NULL UNIQUE,
    tg_file_id VARCHAR(250) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tg_images;
-- +goose StatementEnd
