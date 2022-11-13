// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repository/barrel.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	reflect "reflect"

	repository "github.com/go-seidon/chariot/internal/repository"
	gomock "github.com/golang/mock/gomock"
)

// MockBarrel is a mock of Barrel interface.
type MockBarrel struct {
	ctrl     *gomock.Controller
	recorder *MockBarrelMockRecorder
}

// MockBarrelMockRecorder is the mock recorder for MockBarrel.
type MockBarrelMockRecorder struct {
	mock *MockBarrel
}

// NewMockBarrel creates a new mock instance.
func NewMockBarrel(ctrl *gomock.Controller) *MockBarrel {
	mock := &MockBarrel{ctrl: ctrl}
	mock.recorder = &MockBarrelMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBarrel) EXPECT() *MockBarrelMockRecorder {
	return m.recorder
}

// CreateBarrel mocks base method.
func (m *MockBarrel) CreateBarrel(ctx context.Context, p repository.CreateBarrelParam) (*repository.CreateBarrelResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBarrel", ctx, p)
	ret0, _ := ret[0].(*repository.CreateBarrelResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateBarrel indicates an expected call of CreateBarrel.
func (mr *MockBarrelMockRecorder) CreateBarrel(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBarrel", reflect.TypeOf((*MockBarrel)(nil).CreateBarrel), ctx, p)
}

// FindBarrel mocks base method.
func (m *MockBarrel) FindBarrel(ctx context.Context, p repository.FindBarrelParam) (*repository.FindBarrelResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindBarrel", ctx, p)
	ret0, _ := ret[0].(*repository.FindBarrelResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindBarrel indicates an expected call of FindBarrel.
func (mr *MockBarrelMockRecorder) FindBarrel(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindBarrel", reflect.TypeOf((*MockBarrel)(nil).FindBarrel), ctx, p)
}
