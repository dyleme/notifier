# SQLite-Compatible SQL Queries

This directory contains SQLite-compatible versions of the SQL queries used by sqlc. These queries have been converted from the original PostgreSQL queries to work with SQLite.

## Key Changes Made

### 1. Array Handling
**PostgreSQL:**
```sql
t.id = ANY (@tag_ids::int[])
array_length(@tag_ids::int[], 1) is null
```

**SQLite:**
```sql
t.id IN (SELECT value FROM json_each(@tag_ids))
@tag_ids IS NULL
```

### 2. Boolean Values
**PostgreSQL:**
```sql
done = false
notify = true
```

**SQLite:**
```sql
done = 0
notify = 1
```

### 3. Date/Time Functions
**PostgreSQL:**
```sql
CURRENT_DATE
NOW()
next_send + @time_offset BETWEEN CURRENT_DATE AND CURRENT_DATE + 1
```

**SQLite:**
```sql
date('now')
datetime('now')
datetime(next_send, @time_offset) BETWEEN date('now') AND date('now', '+1 day')
```

## Files Converted

1. **single_tasks.sql** - Basic task operations with tag filtering
2. **events.sql** - Event operations with date/time queries
3. **periodic_tasks.sql** - Periodic task operations with tag filtering
4. **tag.sql** - Tag operations with array handling
5. **key_value.sql** - Key-value store operations (no changes needed)
6. **default_user_notifications.sql** - User notification parameters (no changes needed)
7. **tg_images.sql** - Telegram image operations (no changes needed)

## Usage

To use these SQLite-compatible queries with sqlc:

1. Update your `sqlc.yaml` configuration to point to this directory:
   ```yaml
   sql:
     - engine: "sqlite"
       queries: "internal/service/repository/queries/sql_sqlite"
       schema: "migrations_sqlite"
   ```

2. Generate the Go code:
   ```bash
   sqlc generate
   ```

## Important Notes

- **Array Parameters**: In Go code, pass arrays as JSON strings instead of Go slices
- **Boolean Parameters**: Use integers (0/1) instead of booleans for database operations
- **Date/Time**: SQLite uses different date/time functions, but the Go types remain the same
- **JSON Operations**: SQLite 3.38.0+ supports native JSON operations on JSON columns

## Example Usage in Go

```go
// For array parameters, pass as JSON string
tagIDs := `[1, 2, 3]` // JSON string instead of []int
tasks, err := queries.ListBasicTasks(ctx, ListBasicTasksParams{
    UserID: userID,
    TagIds: tagIDs,
    Lim:    10,
    Off:    0,
})

// For boolean parameters, use integers
events, err := queries.ListNotSendedEvents(ctx, ListNotSendedEventsParams{
    Till: time.Now(),
    // done = 0 (false), notify = 1 (true) in SQL
})
```
