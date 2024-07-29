-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
ADD COLUMN first_send_time TIMESTAMP WITH TIME ZONE,
ADD COLUMN last_sended_time TIMESTAMP WITH TIME ZONE;
UPDATE events
SET
    first_send_time = send_time,
    last_sended_time
    = CASE
        WHEN sended
            THEN send_time
        ELSE TIMESTAMP '1970-01-01 00:00:00'
    END;
ALTER TABLE events ADD CONSTRAINT first_send_time_not_null CHECK (
    first_send_time IS NOT NULL
),
ADD CONSTRAINT last_sended_time_not_null CHECK (last_sended_time IS NOT NULL);

ALTER TABLE events
RENAME COLUMN send_time TO next_send_time;
ALTER TABLE events
DROP COLUMN sended;

ALTER TABLE events
ADD CONSTRAINT notification_params_period_not_null CHECK (
    notification_params ->> 'period' != '0'
    AND notification_params ->> 'params' != '{}'
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events
DROP CONSTRAINT notification_params_period_not_null;
ALTER TABLE events
ADD COLUMN sended BOOL NOT NULL DEFAULT TRUE;

UPDATE events
SET sended = FALSE
WHERE last_sended_time = '1970-01-01 00:00:00';

ALTER TABLE events
RENAME COLUMN next_send_time TO send_time;

ALTER TABLE events
DROP COLUMN first_send_time,
DROP COLUMN last_sended_time;
-- +goose StatementEnd
