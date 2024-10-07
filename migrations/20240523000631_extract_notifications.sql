-- +goose Up
-- +goose StatementBegin
ALTER TABLE events RENAME TO basic_events;

CREATE TYPE event_type AS ENUM ('periodic_event', 'basic_event');

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    user_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    description TEXT,
    event_id INTEGER NOT NULL,
    event_type EVENT_TYPE NOT NULL,
    send_time TIMESTAMP WITH TIME ZONE NOT NULL,
    sended BOOLEAN DEFAULT FALSE NOT NULL,
    done BOOLEAN DEFAULT FALSE NOT NULL,
    notification_params JSONB,
    UNIQUE (event_id, event_type)
);

CREATE INDEX notifications_user_id_idx ON notifications (user_id, event_id);

INSERT INTO notifications (
    created_at,
    user_id,
    text,
    description,
    event_id,
    event_type,
    send_time,
    sended,
    done,
    notification_params
) SELECT
    created_at,
    user_id,
    text,
    description,
    id,
    'basic_event' AS event_type,
    send_time,
    sended,
    done,
    notification_params
FROM basic_events;

INSERT INTO notifications (
    created_at,
    user_id,
    text,
    description,
    event_id,
    event_type,
    send_time,
    sended,
    done,
    notification_params
) SELECT
    n.created_at,
    ev.user_id,
    ev.text,
    ev.description,
    ev.id,
    'periodic_event' AS event_type,
    n.send_time,
    n.sended,
    n.done,
    ev.notification_params
FROM periodic_events AS ev
INNER JOIN periodic_events_notifications AS n
    ON ev.id = n.periodic_event_id;

DROP TABLE IF EXISTS periodic_events_notifications;

ALTER TABLE basic_events
DROP COLUMN send_time,
DROP COLUMN sended,
DROP COLUMN done;



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE basic_events
ADD COLUMN send_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
+ INTERVAL '1 year',
ADD COLUMN sended BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN done BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE basic_events SET
    send_time = n.send_time,
    sended = n.sended,
    done = n.done
FROM notifications AS n
WHERE basic_events.id = n.event_id;


CREATE TABLE periodic_events_notifications
(
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    periodic_event_id INTEGER NOT NULL,
    send_time TIMESTAMP WITH TIME ZONE NOT NULL,
    sended BOOLEAN DEFAULT FALSE NOT NULL,
    done BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT fk_periodic_events_notifications_periodic_event_id FOREIGN KEY (
        periodic_event_id
    ) REFERENCES periodic_events
);

INSERT INTO periodic_events_notifications
(
    created_at,
    periodic_event_id,
    send_time,
    sended,
    done
) (
    SELECT
        created_at,
        event_id,
        send_time,
        sended,
        done
    FROM notifications
    WHERE event_type = 'periodic_event'
);

DROP TABLE IF EXISTS notifications;

DROP TYPE EVENT_TYPE;

ALTER TABLE basic_events RENAME TO events;
-- +goose StatementEnd
