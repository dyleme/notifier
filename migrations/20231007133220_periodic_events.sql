-- +goose Up
-- +goose StatementBegin
CREATE TABLE periodic_events
(
    id                  SERIAL PRIMARY KEY,
    created_at          TIMESTAMP DEFAULT NOW()  NOT NULL,
    text                VARCHAR(250)             NOT NULL,
    description         VARCHAR(250),
    user_id             INTEGER                  NOT NULL,
    start               TIMESTAMP WITH TIME ZONE NOT NULL,
    smallest_period     INTEGER                  NOT NULL,
    biggest_period      INTEGER                  NOT NULL,
    notification_params jsonb,
    CONSTRAINT fk_periodic_events_user_id FOREIGN KEY (user_id) REFERENCES users
);

CREATE TABLE periodic_events_notifications
(
    id                SERIAL PRIMARY KEY,
    periodic_event_id INTEGER                  NOT NULL,
    send_time         TIMESTAMP WITH TIME ZONE NOT NULL,
    sended            BOOLEAN DEFAULT FALSE    NOT NULL,
    done              BOOLEAN DEFAULT FALSE    NOT NULL,
    CONSTRAINT fk_periodic_events_notifications_periodic_event_id FOREIGN KEY (periodic_event_id) REFERENCES periodic_events
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE periodic_events_notifications;
DROP TABLE periodic_events;
-- +goose StatementEnd
