// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Dyleme/Notifier/internal/service/service (interfaces: EventsRepository)
//
// Generated by this command:
//
//	mockgen -destination=mocks/events_mocks.go -package=mocks . EventsRepository
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/Dyleme/Notifier/internal/domain"
	service "github.com/Dyleme/Notifier/internal/service/service"
	gomock "go.uber.org/mock/gomock"
)

// MockEventsRepository is a mock of EventsRepository interface.
type MockEventsRepository struct {
	ctrl     *gomock.Controller
	recorder *MockEventsRepositoryMockRecorder
	isgomock struct{}
}

// MockEventsRepositoryMockRecorder is the mock recorder for MockEventsRepository.
type MockEventsRepositoryMockRecorder struct {
	mock *MockEventsRepository
}

// NewMockEventsRepository creates a new mock instance.
func NewMockEventsRepository(ctrl *gomock.Controller) *MockEventsRepository {
	mock := &MockEventsRepository{ctrl: ctrl}
	mock.recorder = &MockEventsRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventsRepository) EXPECT() *MockEventsRepositoryMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockEventsRepository) Add(ctx context.Context, event domain.Event) (domain.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, event)
	ret0, _ := ret[0].(domain.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Add indicates an expected call of Add.
func (mr *MockEventsRepositoryMockRecorder) Add(ctx, event any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockEventsRepository)(nil).Add), ctx, event)
}

// Delete mocks base method.
func (m *MockEventsRepository) Delete(ctx context.Context, id int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockEventsRepositoryMockRecorder) Delete(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockEventsRepository)(nil).Delete), ctx, id)
}

// Get mocks base method.
func (m *MockEventsRepository) Get(ctx context.Context, id int) (domain.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(domain.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockEventsRepositoryMockRecorder) Get(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockEventsRepository)(nil).Get), ctx, id)
}

// GetLatest mocks base method.
func (m *MockEventsRepository) GetLatest(ctx context.Context, taskdID int, taskType domain.TaskType) (domain.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLatest", ctx, taskdID, taskType)
	ret0, _ := ret[0].(domain.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLatest indicates an expected call of GetLatest.
func (mr *MockEventsRepositoryMockRecorder) GetLatest(ctx, taskdID, taskType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatest", reflect.TypeOf((*MockEventsRepository)(nil).GetLatest), ctx, taskdID, taskType)
}

// List mocks base method.
func (m *MockEventsRepository) List(ctx context.Context, userID int, params service.ListEventsFilterParams) ([]domain.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, userID, params)
	ret0, _ := ret[0].([]domain.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockEventsRepositoryMockRecorder) List(ctx, userID, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockEventsRepository)(nil).List), ctx, userID, params)
}

// Update mocks base method.
func (m *MockEventsRepository) Update(ctx context.Context, event domain.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, event)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockEventsRepositoryMockRecorder) Update(ctx, event any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockEventsRepository)(nil).Update), ctx, event)
}
