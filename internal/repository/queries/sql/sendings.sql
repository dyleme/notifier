

  
-- -- name: ListUserSending :many
-- SELECT DISTINCT e.*
-- FROM sendings as e
-- JOIN tasks as t
-- ON e.task_id = t.id
-- WHERE t.user_id = ?
--   AND next_sending <= @to_time
--   AND next_sending >= @from_time
-- ORDER BY next_sending DESC
-- LIMIT ? OFFSET ?;



-- -- name: ListNotSendedSending :many
-- SELECT * FROM sendings
-- WHERE next_sending <= @till
--   AND done = 0
--   AND notify = 1;


-- -- name: RescheduleSending :exec
-- UPDATE sendings
-- SET next_sending = ?
-- WHERE id = ?;