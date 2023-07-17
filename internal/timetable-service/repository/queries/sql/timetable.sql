-- name: AddTimetableTask :one
INSERT INTO timetable_tasks (
            user_id,
            task_id,
            text,
            start,
            finish,
            description,
            done
) VALUES (
            @user_id,
            @task_id,
            @text,
            @start,
            @finish,
            @description,
            @done
)
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
RETURNING count(*) as deleted_amount;

-- name: UpdateTimetableTask :one
UPDATE timetable_tasks
   SET start = @start,
       finish = @finish,
       text = @text,
       description = @description,
       done = @done
 WHERE id = @id
   AND user_id = @user_id
RETURNING *;
