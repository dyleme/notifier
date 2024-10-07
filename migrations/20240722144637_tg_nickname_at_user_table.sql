-- +goose Up
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN email;
ALTER TABLE users ALTER COLUMN tg_id SET NOT NULL;
ALTER TABLE users ADD tg_nickname VARCHAR(250) DEFAULT '';
ALTER TABLE users ALTER COLUMN tg_nickname SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT unique_tg_nickname UNIQUE(tg_nickname);
CREATE INDEX tg_id_idx ON users (tg_id);
CREATE INDEX tg_nickname_idx ON users (tg_nickname);

CREATE TABLE binding_attempts (
    id SERIAL PRIMARY KEY,
    tg_id INTEGER NOT NULL,
    login_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    code VARCHAR(10) NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    password_hash VARCHAR(250) NOT NULL,
    CONSTRAINT login_tries_unique UNIQUE (tg_id, login_timestamp), 
    FOREIGN KEY (tg_id) REFERENCES users (tg_id)
);
CREATE INDEX binding_attempts_tg_id_login_timestamp_idx ON users (tg_nickname);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE binding_attempts;
DROP INDEX tg_nickname_idx;
DROP INDEX tg_id_idx;
ALTER TABLE users DROP COLUMN tg_nickname;
ALTER TABLE users ALTER COLUMN tg_id DROP NOT NULL;
ALTER TABLE users ADD COLUMN email VARCHAR(250);
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE(email);
-- +goose StatementEnd
