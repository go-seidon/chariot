// Code generated by MockGen. DO NOT EDIT.
// Source: internal/service/barrel.go

// Package mock_service is a generated GoMock package.
package mock_service

import (
	context "context"
	reflect "reflect"

	service "github.com/go-seidon/chariot/internal/service"
	system "github.com/go-seidon/provider/system"
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
func (m *MockBarrel) CreateBarrel(ctx context.Context, p service.CreateBarrelParam) (*service.CreateBarrelResult, *system.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBarrel", ctx, p)
	ret0, _ := ret[0].(*service.CreateBarrelResult)
	ret1, _ := ret[1].(*system.Error)
	return ret0, ret1
}

// CreateBarrel indicates an expected call of CreateBarrel.
func (mr *MockBarrelMockRecorder) CreateBarrel(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBarrel", reflect.TypeOf((*MockBarrel)(nil).CreateBarrel), ctx, p)
}

// FindBarrelById mocks base method.
func (m *MockBarrel) FindBarrelById(ctx context.Context, p service.FindBarrelByIdParam) (*service.FindBarrelByIdResult, *system.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindBarrelById", ctx, p)
	ret0, _ := ret[0].(*service.FindBarrelByIdResult)
	ret1, _ := ret[1].(*system.Error)
	return ret0, ret1
}

// FindBarrelById indicates an expected call of FindBarrelById.
func (mr *MockBarrelMockRecorder) FindBarrelById(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindBarrelById", reflect.TypeOf((*MockBarrel)(nil).FindBarrelById), ctx, p)
}

// SearchBarrel mocks base method.
func (m *MockBarrel) SearchBarrel(ctx context.Context, p service.SearchBarrelParam) (*service.SearchBarrelResult, *system.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchBarrel", ctx, p)
	ret0, _ := ret[0].(*service.SearchBarrelResult)
	ret1, _ := ret[1].(*system.Error)
	return ret0, ret1
}

// SearchBarrel indicates an expected call of SearchBarrel.
func (mr *MockBarrelMockRecorder) SearchBarrel(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchBarrel", reflect.TypeOf((*MockBarrel)(nil).SearchBarrel), ctx, p)
}

// UpdateBarrelById mocks base method.
func (m *MockBarrel) UpdateBarrelById(ctx context.Context, p service.UpdateBarrelByIdParam) (*service.UpdateBarrelByIdResult, *system.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateBarrelById", ctx, p)
	ret0, _ := ret[0].(*service.UpdateBarrelByIdResult)
	ret1, _ := ret[1].(*system.Error)
	return ret0, ret1
}

// UpdateBarrelById indicates an expected call of UpdateBarrelById.
func (mr *MockBarrelMockRecorder) UpdateBarrelById(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateBarrelById", reflect.TypeOf((*MockBarrel)(nil).UpdateBarrelById), ctx, p)
}
