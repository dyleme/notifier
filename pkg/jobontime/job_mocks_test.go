// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Dyleme/Notifier/pkg/jobontime (interfaces: Job)
//
// Generated by this command:
//
//	mockgen -destination=job_mocks_test.go -package=jobontime_test . Job
//

// Package jobontime_test is a generated GoMock package.
package jobontime_test

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockJob is a mock of Job interface.
type MockJob struct {
	ctrl     *gomock.Controller
	recorder *MockJobMockRecorder
}

// MockJobMockRecorder is the mock recorder for MockJob.
type MockJobMockRecorder struct {
	mock *MockJob
}

// NewMockJob creates a new mock instance.
func NewMockJob(ctrl *gomock.Controller) *MockJob {
	mock := &MockJob{ctrl: ctrl}
	mock.recorder = &MockJobMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJob) EXPECT() *MockJobMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockJob) Do(arg0 context.Context, arg1 time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Do", arg0, arg1)
}

// Do indicates an expected call of Do.
func (mr *MockJobMockRecorder) Do(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockJob)(nil).Do), arg0, arg1)
}

// GetNextTime mocks base method.
func (m *MockJob) GetNextTime(arg0 context.Context) time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNextTime", arg0)
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetNextTime indicates an expected call of GetNextTime.
func (mr *MockJobMockRecorder) GetNextTime(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNextTime", reflect.TypeOf((*MockJob)(nil).GetNextTime), arg0)
}
