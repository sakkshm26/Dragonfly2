// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/rpc/manager/server/server.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	manager "d7y.io/dragonfly/v2/pkg/rpc/manager"
	gomock "github.com/golang/mock/gomock"
)

// MockManagerServer is a mock of ManagerServer interface.
type MockManagerServer struct {
	ctrl     *gomock.Controller
	recorder *MockManagerServerMockRecorder
}

// MockManagerServerMockRecorder is the mock recorder for MockManagerServer.
type MockManagerServerMockRecorder struct {
	mock *MockManagerServer
}

// NewMockManagerServer creates a new mock instance.
func NewMockManagerServer(ctrl *gomock.Controller) *MockManagerServer {
	mock := &MockManagerServer{ctrl: ctrl}
	mock.recorder = &MockManagerServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockManagerServer) EXPECT() *MockManagerServerMockRecorder {
	return m.recorder
}

// GetCDN mocks base method.
func (m *MockManagerServer) GetCDN(arg0 context.Context, arg1 *manager.GetCDNRequest) (*manager.CDN, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCDN", arg0, arg1)
	ret0, _ := ret[0].(*manager.CDN)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCDN indicates an expected call of GetCDN.
func (mr *MockManagerServerMockRecorder) GetCDN(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCDN", reflect.TypeOf((*MockManagerServer)(nil).GetCDN), arg0, arg1)
}

// GetScheduler mocks base method.
func (m *MockManagerServer) GetScheduler(arg0 context.Context, arg1 *manager.GetSchedulerRequest) (*manager.Scheduler, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetScheduler", arg0, arg1)
	ret0, _ := ret[0].(*manager.Scheduler)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetScheduler indicates an expected call of GetScheduler.
func (mr *MockManagerServerMockRecorder) GetScheduler(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetScheduler", reflect.TypeOf((*MockManagerServer)(nil).GetScheduler), arg0, arg1)
}

// KeepAlive mocks base method.
func (m *MockManagerServer) KeepAlive(arg0 manager.Manager_KeepAliveServer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KeepAlive", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// KeepAlive indicates an expected call of KeepAlive.
func (mr *MockManagerServerMockRecorder) KeepAlive(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeepAlive", reflect.TypeOf((*MockManagerServer)(nil).KeepAlive), arg0)
}

// ListSchedulers mocks base method.
func (m *MockManagerServer) ListSchedulers(arg0 context.Context, arg1 *manager.ListSchedulersRequest) (*manager.ListSchedulersResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSchedulers", arg0, arg1)
	ret0, _ := ret[0].(*manager.ListSchedulersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSchedulers indicates an expected call of ListSchedulers.
func (mr *MockManagerServerMockRecorder) ListSchedulers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSchedulers", reflect.TypeOf((*MockManagerServer)(nil).ListSchedulers), arg0, arg1)
}

// UpdateCDN mocks base method.
func (m *MockManagerServer) UpdateCDN(arg0 context.Context, arg1 *manager.UpdateCDNRequest) (*manager.CDN, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCDN", arg0, arg1)
	ret0, _ := ret[0].(*manager.CDN)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateCDN indicates an expected call of UpdateCDN.
func (mr *MockManagerServerMockRecorder) UpdateCDN(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCDN", reflect.TypeOf((*MockManagerServer)(nil).UpdateCDN), arg0, arg1)
}

// UpdateScheduler mocks base method.
func (m *MockManagerServer) UpdateScheduler(arg0 context.Context, arg1 *manager.UpdateSchedulerRequest) (*manager.Scheduler, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateScheduler", arg0, arg1)
	ret0, _ := ret[0].(*manager.Scheduler)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateScheduler indicates an expected call of UpdateScheduler.
func (mr *MockManagerServerMockRecorder) UpdateScheduler(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateScheduler", reflect.TypeOf((*MockManagerServer)(nil).UpdateScheduler), arg0, arg1)
}
