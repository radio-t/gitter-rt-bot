// Code generated by MockGen. DO NOT EDIT.
// Source: rtjc.go

// Package mock_events is a generated GoMock package.
package events

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockSubmitter is a mock of Submitter interface
type MockSubmitter struct {
	ctrl     *gomock.Controller
	recorder *MockSubmitterMockRecorder
}

// MockSubmitterMockRecorder is the mock recorder for MockSubmitter
type MockSubmitterMockRecorder struct {
	mock *MockSubmitter
}

// NewMockSubmitter creates a new mock instance
func NewMockSubmitter(ctrl *gomock.Controller) *MockSubmitter {
	mock := &MockSubmitter{ctrl: ctrl}
	mock.recorder = &MockSubmitterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSubmitter) EXPECT() *MockSubmitterMockRecorder {
	return m.recorder
}

// Submit mocks base method
func (m *MockSubmitter) Submit(ctx context.Context, msg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Submit", ctx, msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Submit indicates an expected call of Submit
func (mr *MockSubmitterMockRecorder) Submit(ctx, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Submit", reflect.TypeOf((*MockSubmitter)(nil).Submit), ctx, msg)
}
