// Code generated by MockGen. DO NOT EDIT.
// Source: internal/queueing/queue.go

// Package mock_queueing is a generated GoMock package.
package mock_queueing

import (
	context "context"
	reflect "reflect"

	queueing "github.com/go-seidon/chariot/internal/queueing"
	gomock "github.com/golang/mock/gomock"
)

// MockQueue is a mock of Queue interface.
type MockQueue struct {
	ctrl     *gomock.Controller
	recorder *MockQueueMockRecorder
}

// MockQueueMockRecorder is the mock recorder for MockQueue.
type MockQueueMockRecorder struct {
	mock *MockQueue
}

// NewMockQueue creates a new mock instance.
func NewMockQueue(ctrl *gomock.Controller) *MockQueue {
	mock := &MockQueue{ctrl: ctrl}
	mock.recorder = &MockQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueue) EXPECT() *MockQueueMockRecorder {
	return m.recorder
}

// DeclareQueue mocks base method.
func (m *MockQueue) DeclareQueue(ctx context.Context, p queueing.DeclareQueueParam) (*queueing.DeclareQueueResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeclareQueue", ctx, p)
	ret0, _ := ret[0].(*queueing.DeclareQueueResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeclareQueue indicates an expected call of DeclareQueue.
func (mr *MockQueueMockRecorder) DeclareQueue(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeclareQueue", reflect.TypeOf((*MockQueue)(nil).DeclareQueue), ctx, p)
}
