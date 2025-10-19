package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/pkg/database/sqldatabase"
	"github.com/dyleme/Notifier/pkg/database/txmanager"
)

func setupTestDB(t *testing.T) (db *sql.DB, cleanupFunc func()) {
	t.Helper()

	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	// Initialize the database
	ctx := context.Background()
	db, closeDB, err := sqldatabase.NewSQLite(ctx, tmpFile.Name())
	require.NoError(t, err)

	// Run migrations using goose
	err = goose.SetDialect("sqlite3")
	require.NoError(t, err)

	err = goose.Up(db, "../../migrations")
	require.NoError(t, err)

	// Insert a test user with the correct schema
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (tg_id, timezone_offset, timezone_dst, notification_retry_period_s) 
		VALUES (12345, 0, 0, 3600)
	`)
	require.NoError(t, err)

	cleanup := func() {
		closeDB()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

func setupTaskRepo(t *testing.T) (repo *TasksRepository, cleanup func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)

	txGetter := txmanager.NewGetter(db)
	repo = NewTasksRepository(txGetter)

	return repo, cleanup
}

func TestTasksRepository_AddAndGet(t *testing.T) {
	t.Parallel()

	t.Run("add and get task", func(t *testing.T) {
		t.Parallel()

		repo, cleanup := setupTaskRepo(t)
		defer cleanup()

		ctx := context.Background()
		// Create a test task
		originalTask := domain.Task{
			Text:        "Test task",
			Description: "This is a test task",
			UserID:      1,               // The user we created in setup
			Type:        domain.Periodic, // Just use one type for testing
			Start:       2 * time.Hour,   // 2 hours from midnight
			Params: map[domain.TaskParamKey]any{
				"test_param": "test_value",
			},
		}

		// Add the task
		addedTask, err := repo.Add(ctx, originalTask)
		require.NoError(t, err)
		require.NotZero(t, addedTask.ID)
		require.NotZero(t, addedTask.CreatedAt)

		// Get the task back
		retrievedTask, err := repo.Get(ctx, addedTask.ID, originalTask.UserID)
		require.NoError(t, err)

		// Compare the tasks - verify that what we stored is what we get back
		assert.Equal(t, addedTask.ID, retrievedTask.ID)
		assert.Equal(t, originalTask.Text, retrievedTask.Text)
		assert.Equal(t, originalTask.Description, retrievedTask.Description)
		assert.Equal(t, originalTask.UserID, retrievedTask.UserID)
		assert.Equal(t, originalTask.Type, retrievedTask.Type)
		assert.Equal(t, originalTask.Start, retrievedTask.Start)
		assert.Equal(t, originalTask.Params, retrievedTask.Params)
		assert.Equal(t, addedTask.CreatedAt, retrievedTask.CreatedAt)
	})

	t.Run("get non-existent task", func(t *testing.T) {
		t.Parallel()

		repo, cleanup := setupTaskRepo(t)
		defer cleanup()
		ctx := context.Background()

		_, err := repo.Get(ctx, 999, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no rows")
	})

	t.Run("get task with wrong user", func(t *testing.T) {
		t.Parallel()

		repo, cleanup := setupTaskRepo(t)
		defer cleanup()
		ctx := context.Background()

		// First add a task
		originalTask := domain.Task{
			Text:        "Test task",
			Description: "This is a test task",
			UserID:      1,
			Type:        domain.Periodic,
			Start:       1 * time.Hour,
			Params:      map[domain.TaskParamKey]any{},
		}

		addedTask, err := repo.Add(ctx, originalTask)
		require.NoError(t, err)

		// Try to get it with wrong user ID
		_, err = repo.Get(ctx, addedTask.ID, 999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no rows")
	})
}

func TestTasksRepository_Add_ComplexParams(t *testing.T) {
	t.Parallel()

	t.Run("add task with complex JSON parameters", func(t *testing.T) {
		t.Parallel()

		repo, cleanup := setupTaskRepo(t)
		defer cleanup()
		ctx := context.Background()
		// Create a task with complex nested parameters to test JSON serialization/deserialization
		originalTask := domain.Task{
			Text:        "Complex task",
			Description: "Task with complex parameters",
			UserID:      1,
			Type:        domain.Periodic,
			Start:       3 * time.Hour,
			Params: map[domain.TaskParamKey]any{
				"config": map[string]any{
					"enabled": true,
					"timeout": 30.0, // Use float64 to match JSON unmarshaling
				},
				"tags": []any{"urgent", "work"},
			},
		}

		// Add the task
		addedTask, err := repo.Add(ctx, originalTask)
		require.NoError(t, err)

		// Get the task back
		retrievedTask, err := repo.Get(ctx, addedTask.ID, originalTask.UserID)
		require.NoError(t, err)

		// Compare the complex parameters - verify JSON round-trip works correctly
		assert.Equal(t, originalTask.Params, retrievedTask.Params)

		// Verify specific nested values
		originalConfig := originalTask.Params["config"].(map[string]any)
		retrievedConfig := retrievedTask.Params["config"].(map[string]any)
		assert.Equal(t, originalConfig["enabled"], retrievedConfig["enabled"])
		assert.Equal(t, originalConfig["timeout"], retrievedConfig["timeout"])

		originalTags := originalTask.Params["tags"].([]any)
		retrievedTags := retrievedTask.Params["tags"].([]any)
		assert.Len(t, retrievedTags, 2)
		assert.Equal(t, originalTags, retrievedTags)
	})
}
