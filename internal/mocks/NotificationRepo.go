// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/bangumi/server/internal/model"
)

// NotificationRepo is an autogenerated mock type for the NotificationRepo type
type NotificationRepo struct {
	mock.Mock
}

type NotificationRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *NotificationRepo) EXPECT() *NotificationRepo_Expecter {
	return &NotificationRepo_Expecter{mock: &_m.Mock}
}

// Count provides a mock function with given fields: ctx, userID
func (_m *NotificationRepo) Count(ctx context.Context, userID model.UserID) (int64, error) {
	ret := _m.Called(ctx, userID)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, model.UserID) int64); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.UserID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NotificationRepo_Count_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Count'
type NotificationRepo_Count_Call struct {
	*mock.Call
}

// Count is a helper method to define mock.On call
//   - ctx context.Context
//   - userID model.UserID
func (_e *NotificationRepo_Expecter) Count(ctx interface{}, userID interface{}) *NotificationRepo_Count_Call {
	return &NotificationRepo_Count_Call{Call: _e.mock.On("Count", ctx, userID)}
}

func (_c *NotificationRepo_Count_Call) Run(run func(ctx context.Context, userID model.UserID)) *NotificationRepo_Count_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.UserID))
	})
	return _c
}

func (_c *NotificationRepo_Count_Call) Return(_a0 int64, _a1 error) *NotificationRepo_Count_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewNotificationRepo interface {
	mock.TestingT
	Cleanup(func())
}

// NewNotificationRepo creates a new instance of NotificationRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNotificationRepo(t mockConstructorTestingTNewNotificationRepo) *NotificationRepo {
	mock := &NotificationRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}