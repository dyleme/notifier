package service

import (
	"context"
	"errors"
	"time"

	"github.com/Dyleme/Notifier/internal/notification-service/notifier/notification"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

type Repository interface {
	Atomic(ctx context.Context, fn func(ctx context.Context, repo Repository) error) error

	AddTask(ctx context.Context, task models.Task) (models.Task, error)
	GetTask(ctx context.Context, taskID, userID int) (models.Task, error)
	DeleteTask(ctx context.Context, taskID, userID int) error
	UpdateTask(ctx context.Context, task models.Task) error
	ListTasks(ctx context.Context, userID int) ([]models.Task, error)

	AddTimetableTask(context.Context, models.TimetableTask) (models.TimetableTask, error)
	ListTimetableTasks(ctx context.Context, userID int) ([]models.TimetableTask, error)
	UpdateTimetableTask(ctx context.Context, timetableTask models.TimetableTask) (models.TimetableTask, error)
	DeleteTimetableTask(ctx context.Context, timetableTaskID, userID int) error
	ListTimetableTasksInPeriod(ctx context.Context, userID int, from, to time.Time) ([]models.TimetableTask, error)
	GetTimetableTask(ctx context.Context, timetableTaskID, userID int) (models.TimetableTask, error)
}

type Notifier interface {
	AddNotification(ctx context.Context, n notification.Notification) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

var (
	ErrNotFound = errors.New("not found")
)
