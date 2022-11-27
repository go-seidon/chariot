// Code generated by MockGen. DO NOT EDIT.
// Source: internal/file/file.go

// Package mock_file is a generated GoMock package.
package mock_file

import (
	context "context"
	reflect "reflect"

	file "github.com/go-seidon/chariot/internal/file"
	system "github.com/go-seidon/provider/system"
	gomock "github.com/golang/mock/gomock"
)

// MockFile is a mock of File interface.
type MockFile struct {
	ctrl     *gomock.Controller
	recorder *MockFileMockRecorder
}

// MockFileMockRecorder is the mock recorder for MockFile.
type MockFileMockRecorder struct {
	mock *MockFile
}

// NewMockFile creates a new mock instance.
func NewMockFile(ctrl *gomock.Controller) *MockFile {
	mock := &MockFile{ctrl: ctrl}
	mock.recorder = &MockFileMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFile) EXPECT() *MockFileMockRecorder {
	return m.recorder
}

// GetFileById mocks base method.
func (m *MockFile) GetFileById(ctx context.Context, p file.GetFileByIdParam) (*file.GetFileByIdResult, *system.SystemError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileById", ctx, p)
	ret0, _ := ret[0].(*file.GetFileByIdResult)
	ret1, _ := ret[1].(*system.SystemError)
	return ret0, ret1
}

// GetFileById indicates an expected call of GetFileById.
func (mr *MockFileMockRecorder) GetFileById(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileById", reflect.TypeOf((*MockFile)(nil).GetFileById), ctx, p)
}

// RetrieveFileBySlug mocks base method.
func (m *MockFile) RetrieveFileBySlug(ctx context.Context, p file.RetrieveFileBySlugParam) (*file.RetrieveFileBySlugResult, *system.SystemError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetrieveFileBySlug", ctx, p)
	ret0, _ := ret[0].(*file.RetrieveFileBySlugResult)
	ret1, _ := ret[1].(*system.SystemError)
	return ret0, ret1
}

// RetrieveFileBySlug indicates an expected call of RetrieveFileBySlug.
func (mr *MockFileMockRecorder) RetrieveFileBySlug(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetrieveFileBySlug", reflect.TypeOf((*MockFile)(nil).RetrieveFileBySlug), ctx, p)
}

// UploadFile mocks base method.
func (m *MockFile) UploadFile(ctx context.Context, p file.UploadFileParam) (*file.UploadFileResult, *system.SystemError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadFile", ctx, p)
	ret0, _ := ret[0].(*file.UploadFileResult)
	ret1, _ := ret[1].(*system.SystemError)
	return ret0, ret1
}

// UploadFile indicates an expected call of UploadFile.
func (mr *MockFileMockRecorder) UploadFile(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadFile", reflect.TypeOf((*MockFile)(nil).UploadFile), ctx, p)
}
