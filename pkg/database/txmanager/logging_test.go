package txmanager

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"

	"github.com/dyleme/Notifier/pkg/log/mocklogger"

	_ "modernc.org/sqlite"
)

// customError is a custom error type for testing error comparison.
type customError struct {
	msg string
}

func (e customError) Error() string {
	return e.msg
}

// mockDB is a mock implementation of DBTX for testing.
type mockDB struct {
	execError    error
	queryError   error
	prepareError error
}

func (m *mockDB) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, m.execError
}

func (m *mockDB) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, m.queryError
}

func (m *mockDB) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return &sql.Row{}
}

func (m *mockDB) PrepareContext(_ context.Context, _ string) (*sql.Stmt, error) {
	return nil, m.prepareError
}

func TestTxManager_WithLogging(t *testing.T) { //nolint:cyclop // tests
	t.Parallel()
	tests := []struct {
		name          string
		setupMock     func() *mockDB
		expectedError bool
		ignoredErrors []error
		testMethod    string
	}{
		{
			name: "logs exec error",
			setupMock: func() *mockDB {
				return &mockDB{
					execError: errors.New("exec failed"),
				}
			},
			expectedError: true,
			testMethod:    "ExecContext",
		},
		{
			name: "logs query error",
			setupMock: func() *mockDB {
				return &mockDB{
					queryError: errors.New("query failed"),
				}
			},
			expectedError: true,
			testMethod:    "QueryContext",
		},
		{
			name: "logs prepare error",
			setupMock: func() *mockDB {
				return &mockDB{
					prepareError: errors.New("prepare failed"),
				}
			},
			expectedError: true,
			testMethod:    "PrepareContext",
		},
		{
			name: "ignores specified errors",
			setupMock: func() *mockDB {
				return &mockDB{
					execError: customError{msg: "ignored error"},
				}
			},
			expectedError: false,
			ignoredErrors: []error{customError{msg: "ignored error"}},
			testMethod:    "ExecContext",
		},
		{
			name: "no error when operation succeeds",
			setupMock: func() *mockDB {
				return &mockDB{}
			},
			expectedError: false,
			testMethod:    "ExecContext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create mock logger handler
			mockHandler := mocklogger.NewHandler()
			logger := slog.New(mockHandler)

			// Create mock database
			mockDB := tt.setupMock()

			// Create logging wrapper
			loggingWrapper := WithLogging(
				func(_ context.Context) *slog.Logger { return logger },
				slog.LevelDebug,
				slog.LevelError,
				tt.ignoredErrors,
			)

			// Wrap the mock database with logging
			loggedDB := loggingWrapper(mockDB)

			// Test the appropriate method based on testMethod
			switch tt.testMethod {
			case "ExecContext":
				_, err := loggedDB.ExecContext(context.Background(), "SELECT 1", "arg1")
				if mockDB.execError != nil && err == nil {
					t.Error("expected error from ExecContext")
				}
			case "QueryContext":
				_, err := loggedDB.QueryContext(context.Background(), "SELECT 1", "arg1") //nolint:sqlclosecheck,gocritic,rowserrcheck // okay in test
				if mockDB.queryError != nil && err == nil {
					t.Error("expected error from QueryContext")
				}
			case "PrepareContext":
				_, err := loggedDB.PrepareContext(context.Background(), "SELECT 1") //nolint:sqlclosecheck // okay in test
				if mockDB.prepareError != nil && err == nil {
					t.Error("expected error from PrepareContext")
				}
			}

			// Check if error was logged
			loggedErr := mockHandler.Error()
			if tt.expectedError {
				if loggedErr == nil {
					t.Error("expected error to be logged")
				}
			} else {
				if loggedErr != nil {
					t.Errorf("expected no error to be logged, got: %v", loggedErr)
				}
			}
		})
	}
}

func TestTxManager_WithLogging_Integration(t *testing.T) {
	t.Parallel()
	// Create a real SQLite database for integration testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create mock logger handler
	mockHandler := mocklogger.NewHandler()
	logger := slog.New(mockHandler)

	// Create logging wrapper
	loggingWrapper := WithLogging(
		func(_ context.Context) *slog.Logger { return logger },
		slog.LevelDebug,
		slog.LevelError,
		nil,
	)

	// Wrap the database with logging
	loggedDB := loggingWrapper(db)

	// Test with a query that will fail using the logged database
	_, err = loggedDB.ExecContext(context.Background(), "INVALID SQL QUERY")

	// The query should fail
	if err == nil {
		t.Error("expected query to fail")
	}

	// Check if error was logged
	loggedErr := mockHandler.Error()
	if loggedErr == nil {
		t.Error("expected error to be logged")
	}
}
