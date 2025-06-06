// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/tag"
	mock "github.com/stretchr/testify/mock"
)

// NewTagRepo creates a new instance of TagRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTagRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *TagRepo {
	mock := &TagRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// TagRepo is an autogenerated mock type for the Repo type
type TagRepo struct {
	mock.Mock
}

type TagRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *TagRepo) EXPECT() *TagRepo_Expecter {
	return &TagRepo_Expecter{mock: &_m.Mock}
}

// Get provides a mock function for the type TagRepo
func (_mock *TagRepo) Get(ctx context.Context, id model.SubjectID, typeID model.SubjectType) ([]tag.Tag, error) {
	ret := _mock.Called(ctx, id, typeID)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 []tag.Tag
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, model.SubjectID, model.SubjectType) ([]tag.Tag, error)); ok {
		return returnFunc(ctx, id, typeID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, model.SubjectID, model.SubjectType) []tag.Tag); ok {
		r0 = returnFunc(ctx, id, typeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]tag.Tag)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, model.SubjectID, model.SubjectType) error); ok {
		r1 = returnFunc(ctx, id, typeID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// TagRepo_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type TagRepo_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - id model.SubjectID
//   - typeID model.SubjectType
func (_e *TagRepo_Expecter) Get(ctx interface{}, id interface{}, typeID interface{}) *TagRepo_Get_Call {
	return &TagRepo_Get_Call{Call: _e.mock.On("Get", ctx, id, typeID)}
}

func (_c *TagRepo_Get_Call) Run(run func(ctx context.Context, id model.SubjectID, typeID model.SubjectType)) *TagRepo_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 model.SubjectID
		if args[1] != nil {
			arg1 = args[1].(model.SubjectID)
		}
		var arg2 model.SubjectType
		if args[2] != nil {
			arg2 = args[2].(model.SubjectType)
		}
		run(
			arg0,
			arg1,
			arg2,
		)
	})
	return _c
}

func (_c *TagRepo_Get_Call) Return(tags []tag.Tag, err error) *TagRepo_Get_Call {
	_c.Call.Return(tags, err)
	return _c
}

func (_c *TagRepo_Get_Call) RunAndReturn(run func(ctx context.Context, id model.SubjectID, typeID model.SubjectType) ([]tag.Tag, error)) *TagRepo_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetByIDs provides a mock function for the type TagRepo
func (_mock *TagRepo) GetByIDs(ctx context.Context, ids []model.SubjectID) (map[model.SubjectID][]tag.Tag, error) {
	ret := _mock.Called(ctx, ids)

	if len(ret) == 0 {
		panic("no return value specified for GetByIDs")
	}

	var r0 map[model.SubjectID][]tag.Tag
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, []model.SubjectID) (map[model.SubjectID][]tag.Tag, error)); ok {
		return returnFunc(ctx, ids)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, []model.SubjectID) map[model.SubjectID][]tag.Tag); ok {
		r0 = returnFunc(ctx, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[model.SubjectID][]tag.Tag)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, []model.SubjectID) error); ok {
		r1 = returnFunc(ctx, ids)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// TagRepo_GetByIDs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByIDs'
type TagRepo_GetByIDs_Call struct {
	*mock.Call
}

// GetByIDs is a helper method to define mock.On call
//   - ctx context.Context
//   - ids []model.SubjectID
func (_e *TagRepo_Expecter) GetByIDs(ctx interface{}, ids interface{}) *TagRepo_GetByIDs_Call {
	return &TagRepo_GetByIDs_Call{Call: _e.mock.On("GetByIDs", ctx, ids)}
}

func (_c *TagRepo_GetByIDs_Call) Run(run func(ctx context.Context, ids []model.SubjectID)) *TagRepo_GetByIDs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 []model.SubjectID
		if args[1] != nil {
			arg1 = args[1].([]model.SubjectID)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *TagRepo_GetByIDs_Call) Return(vToTags map[model.SubjectID][]tag.Tag, err error) *TagRepo_GetByIDs_Call {
	_c.Call.Return(vToTags, err)
	return _c
}

func (_c *TagRepo_GetByIDs_Call) RunAndReturn(run func(ctx context.Context, ids []model.SubjectID) (map[model.SubjectID][]tag.Tag, error)) *TagRepo_GetByIDs_Call {
	_c.Call.Return(run)
	return _c
}
