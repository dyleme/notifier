-- name: AddTimetableTask :one
INSERT INTO timetable_tasks (user_id,
                             task_id,
                             text,
                             start,
                             finish,
                             description,
                             done,
                             notification)
VALUES (@user_id,
        @task_id,
        @text,
        @start,
        @finish,
        @description,
        @done,
        @notification)
RETURNING *;

-- name: GetTimetableTask :one
SELECT *
FROM timetable_tasks
WHERE id = @id
  AND user_id = @user_id;

-- name: ListTimetableTasks :many
SELECT *
FROM timetable_tasks
WHERE user_id = @user_id;

-- name: GetTimetableTasksInPeriod :many
SELECT *
FROM timetable_tasks
WHERE user_id = @user_id
  AND start BETWEEN @from_time AND @to_time;

-- name: DeleteTimetableTask :one
DELETE
FROM timetable_tasks
WHERE id = @id
  AND user_id = @user_id
RETURNING COUNT(*) AS deleted_amount;

-- name: UpdateTimetableTask :one
UPDATE timetable_tasks
SET start       = @start,
    finish      = @finish,
    text        = @text,
    description = @description,
    done        = @done
WHERE id = @id
  AND user_id = @user_id
RETURNING *;

-- name: GetTimetableReadyTasks :many
SELECT *
FROM timetable_tasks AS t
WHERE t.start <= NOW()
  AND t.done = FALSE
  AND t.notification ->> 'sended' = 'false';

-- name: MarkNotificationSended :exec
UPDATE timetable_tasks AS t
SET notification = notification || '{"sended":true}'
WHERE id = ANY(sqlc.arg(ids)::INTEGER[]);

-- name: UpdateNotificationParams :one
UPDATE timetable_tasks AS t
SET notification = JSONB_SET(notification, '{notification_params}', @params)
WHERE id = @id
  AND user_id = @user_id
RETURNING notification;

