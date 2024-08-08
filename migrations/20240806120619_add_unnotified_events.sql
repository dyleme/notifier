-- +goose Up
-- +goose StatementBegin
ALTER TABLE events 
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
ALTER TABLE periodic_tasks        
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
ALTER TABLE basic_tasks        
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
  
ALTER TABLE users
  ADD COLUMN daily_notification_time TIME WITH TIME ZONE 
  NOT NULL DEFAULT '12:00:00+00:00';
ALTER TABLE events ALTER COLUMN notification_params DROP NOT NULL;
ALTER TABLE events RENAME COLUMN next_send_time TO time;
 ALTER TABLE events RENAME COLUMN first_send_time TO first_time;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events RENAME COLUMN time TO next_send_time;
 ALTER TABLE events RENAME COLUMN first_time TO first_send_time;
ALTER TABLE users DROP COLUMN daily_notification_time;
ALTER TABLE events DROP COLUMN notify;
ALTER TABLE periodic_tasks DROP COLUMN notify;
ALTER TABLE basic_tasks DROP COLUMN notify;
-- +goose StatementEnd
