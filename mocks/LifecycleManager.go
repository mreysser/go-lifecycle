// Code generated by mockery v2.24.0. DO NOT EDIT.

package mocks

import (
	context "context"

	lifecycle "github.com/mreysser/go-lifecycle"
	mock "github.com/stretchr/testify/mock"
)

// MockLifecycleManager is an autogenerated mock type for the LifecycleManager type
type MockLifecycleManager struct {
	mock.Mock
}

// GetContext provides a mock function with given fields:
func (_m *MockLifecycleManager) GetContext() context.Context {
	ret := _m.Called()

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// RegisterShutdownHandler provides a mock function with given fields: handler
func (_m *MockLifecycleManager) RegisterShutdownHandler(handler lifecycle.ShutdownHandler) {
	_m.Called(handler)
}

// TerminateLifecycle provides a mock function with given fields:
func (_m *MockLifecycleManager) TerminateLifecycle() {
	_m.Called()
}

type mockConstructorTestingTNewMockLifecycleManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockLifecycleManager creates a new instance of MockLifecycleManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockLifecycleManager(t mockConstructorTestingTNewMockLifecycleManager) *MockLifecycleManager {
	mock := &MockLifecycleManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
