-- +goose Up
-- +goose StatementBegin
ALTER TABLE events 
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
ALTER TABLE periodic_tasks        
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
ALTER TABLE basic_tasks        
  ADD COLUMN notify boolean NOT NULL DEFAULT true;  
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN notify;
ALTER TABLE periodic_tasks DROP COLUMN notify;
ALTER TABLE basic_tasks DROP COLUMN notify;
-- +goose StatementEnd
