-- +goose Up
-- +goose StatementBegin
ALTER TABLE basic_tasks        
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
ALTER TABLE periodic_tasks        
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
ALTER TABLE events 
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
  
ALTER TABLE users
  ADD COLUMN daily_notification_time TIME WITH TIME ZONE 
  NOT NULL DEFAULT '12:00:00+00:00';
ALTER TABLE events RENAME COLUMN next_send_time TO next_send;
ALTER TABLE events RENAME COLUMN first_send_time TO first_send;
ALTER TABLE events DROP COLUMN last_sended_time;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events ADD COLUMN last_sended_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE events RENAME COLUMN first_send TO first_send_time;
ALTER TABLE events RENAME COLUMN next_send TO next_send_time;
ALTER TABLE users DROP COLUMN daily_notification_time;
ALTER TABLE events DROP COLUMN notify;
ALTER TABLE periodic_tasks DROP COLUMN notify;
ALTER TABLE basic_tasks DROP COLUMN notify;
-- +goose StatementEnd
