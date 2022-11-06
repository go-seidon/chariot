// Code generated by MockGen. DO NOT EDIT.
// Source: internal/auth/client.go

// Package mock_auth is a generated GoMock package.
package mock_auth

import (
	context "context"
	reflect "reflect"

	auth "github.com/go-seidon/chariot/internal/auth"
	system "github.com/go-seidon/provider/system"
	gomock "github.com/golang/mock/gomock"
)

// MockAuthClient is a mock of AuthClient interface.
type MockAuthClient struct {
	ctrl     *gomock.Controller
	recorder *MockAuthClientMockRecorder
}

// MockAuthClientMockRecorder is the mock recorder for MockAuthClient.
type MockAuthClientMockRecorder struct {
	mock *MockAuthClient
}

// NewMockAuthClient creates a new mock instance.
func NewMockAuthClient(ctrl *gomock.Controller) *MockAuthClient {
	mock := &MockAuthClient{ctrl: ctrl}
	mock.recorder = &MockAuthClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthClient) EXPECT() *MockAuthClientMockRecorder {
	return m.recorder
}

// CreateClient mocks base method.
func (m *MockAuthClient) CreateClient(ctx context.Context, p auth.CreateClientParam) (*auth.CreateClientResult, *system.SystemError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateClient", ctx, p)
	ret0, _ := ret[0].(*auth.CreateClientResult)
	ret1, _ := ret[1].(*system.SystemError)
	return ret0, ret1
}

// CreateClient indicates an expected call of CreateClient.
func (mr *MockAuthClientMockRecorder) CreateClient(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateClient", reflect.TypeOf((*MockAuthClient)(nil).CreateClient), ctx, p)
}
