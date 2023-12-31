// Code generated by MockGen. DO NOT EDIT.
// Source: jobs.go

// Package mockgen is a generated GoMock package.
package mockgen

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockJobController is a mock of JobController interface.
type MockJobController struct {
	ctrl     *gomock.Controller
	recorder *MockJobControllerMockRecorder
}

// MockJobControllerMockRecorder is the mock recorder for MockJobController.
type MockJobControllerMockRecorder struct {
	mock *MockJobController
}

// NewMockJobController creates a new mock instance.
func NewMockJobController(ctrl *gomock.Controller) *MockJobController {
	mock := &MockJobController{ctrl: ctrl}
	mock.recorder = &MockJobControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJobController) EXPECT() *MockJobControllerMockRecorder {
	return m.recorder
}

// TriggerJob mocks base method.
func (m *MockJobController) TriggerJob(ctx context.Context, queue string, data interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TriggerJob", ctx, queue, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// TriggerJob indicates an expected call of TriggerJob.
func (mr *MockJobControllerMockRecorder) TriggerJob(ctx, queue, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TriggerJob", reflect.TypeOf((*MockJobController)(nil).TriggerJob), ctx, queue, data)
}
