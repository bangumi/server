// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// NewTimelineService creates a new instance of TimelineService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTimelineService(t interface {
	mock.TestingT
	Cleanup(func())
}) *TimelineService {
	mock := &TimelineService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// TimelineService is an autogenerated mock type for the Service type
type TimelineService struct {
	mock.Mock
}

type TimelineService_Expecter struct {
	mock *mock.Mock
}

func (_m *TimelineService) EXPECT() *TimelineService_Expecter {
	return &TimelineService_Expecter{mock: &_m.Mock}
}

// ChangeEpisodeStatus provides a mock function for the type TimelineService
func (_mock *TimelineService) ChangeEpisodeStatus(ctx context.Context, u auth.Auth, sbj model.Subject, episode1 episode.Episode, t collection.EpisodeCollection) error {
	ret := _mock.Called(ctx, u, sbj, episode1, t)

	if len(ret) == 0 {
		panic("no return value specified for ChangeEpisodeStatus")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, auth.Auth, model.Subject, episode.Episode, collection.EpisodeCollection) error); ok {
		r0 = returnFunc(ctx, u, sbj, episode1, t)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// TimelineService_ChangeEpisodeStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChangeEpisodeStatus'
type TimelineService_ChangeEpisodeStatus_Call struct {
	*mock.Call
}

// ChangeEpisodeStatus is a helper method to define mock.On call
//   - ctx
//   - u
//   - sbj
//   - episode1
//   - t
func (_e *TimelineService_Expecter) ChangeEpisodeStatus(ctx interface{}, u interface{}, sbj interface{}, episode1 interface{}, t interface{}) *TimelineService_ChangeEpisodeStatus_Call {
	return &TimelineService_ChangeEpisodeStatus_Call{Call: _e.mock.On("ChangeEpisodeStatus", ctx, u, sbj, episode1, t)}
}

func (_c *TimelineService_ChangeEpisodeStatus_Call) Run(run func(ctx context.Context, u auth.Auth, sbj model.Subject, episode1 episode.Episode, t collection.EpisodeCollection)) *TimelineService_ChangeEpisodeStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(auth.Auth), args[2].(model.Subject), args[3].(episode.Episode), args[4].(collection.EpisodeCollection))
	})
	return _c
}

func (_c *TimelineService_ChangeEpisodeStatus_Call) Return(err error) *TimelineService_ChangeEpisodeStatus_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *TimelineService_ChangeEpisodeStatus_Call) RunAndReturn(run func(ctx context.Context, u auth.Auth, sbj model.Subject, episode1 episode.Episode, t collection.EpisodeCollection) error) *TimelineService_ChangeEpisodeStatus_Call {
	_c.Call.Return(run)
	return _c
}

// ChangeSubjectCollection provides a mock function for the type TimelineService
func (_mock *TimelineService) ChangeSubjectCollection(ctx context.Context, u model.UserID, sbj model.Subject, collect collection.SubjectCollection, collectID uint64, comment string, rate uint8) error {
	ret := _mock.Called(ctx, u, sbj, collect, collectID, comment, rate)

	if len(ret) == 0 {
		panic("no return value specified for ChangeSubjectCollection")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, model.UserID, model.Subject, collection.SubjectCollection, uint64, string, uint8) error); ok {
		r0 = returnFunc(ctx, u, sbj, collect, collectID, comment, rate)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// TimelineService_ChangeSubjectCollection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChangeSubjectCollection'
type TimelineService_ChangeSubjectCollection_Call struct {
	*mock.Call
}

// ChangeSubjectCollection is a helper method to define mock.On call
//   - ctx
//   - u
//   - sbj
//   - collect
//   - collectID
//   - comment
//   - rate
func (_e *TimelineService_Expecter) ChangeSubjectCollection(ctx interface{}, u interface{}, sbj interface{}, collect interface{}, collectID interface{}, comment interface{}, rate interface{}) *TimelineService_ChangeSubjectCollection_Call {
	return &TimelineService_ChangeSubjectCollection_Call{Call: _e.mock.On("ChangeSubjectCollection", ctx, u, sbj, collect, collectID, comment, rate)}
}

func (_c *TimelineService_ChangeSubjectCollection_Call) Run(run func(ctx context.Context, u model.UserID, sbj model.Subject, collect collection.SubjectCollection, collectID uint64, comment string, rate uint8)) *TimelineService_ChangeSubjectCollection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.UserID), args[2].(model.Subject), args[3].(collection.SubjectCollection), args[4].(uint64), args[5].(string), args[6].(uint8))
	})
	return _c
}

func (_c *TimelineService_ChangeSubjectCollection_Call) Return(err error) *TimelineService_ChangeSubjectCollection_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *TimelineService_ChangeSubjectCollection_Call) RunAndReturn(run func(ctx context.Context, u model.UserID, sbj model.Subject, collect collection.SubjectCollection, collectID uint64, comment string, rate uint8) error) *TimelineService_ChangeSubjectCollection_Call {
	_c.Call.Return(run)
	return _c
}

// ChangeSubjectProgress provides a mock function for the type TimelineService
func (_mock *TimelineService) ChangeSubjectProgress(ctx context.Context, u model.UserID, sbj model.Subject, epsUpdate uint32, volsUpdate uint32) error {
	ret := _mock.Called(ctx, u, sbj, epsUpdate, volsUpdate)

	if len(ret) == 0 {
		panic("no return value specified for ChangeSubjectProgress")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, model.UserID, model.Subject, uint32, uint32) error); ok {
		r0 = returnFunc(ctx, u, sbj, epsUpdate, volsUpdate)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// TimelineService_ChangeSubjectProgress_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChangeSubjectProgress'
type TimelineService_ChangeSubjectProgress_Call struct {
	*mock.Call
}

// ChangeSubjectProgress is a helper method to define mock.On call
//   - ctx
//   - u
//   - sbj
//   - epsUpdate
//   - volsUpdate
func (_e *TimelineService_Expecter) ChangeSubjectProgress(ctx interface{}, u interface{}, sbj interface{}, epsUpdate interface{}, volsUpdate interface{}) *TimelineService_ChangeSubjectProgress_Call {
	return &TimelineService_ChangeSubjectProgress_Call{Call: _e.mock.On("ChangeSubjectProgress", ctx, u, sbj, epsUpdate, volsUpdate)}
}

func (_c *TimelineService_ChangeSubjectProgress_Call) Run(run func(ctx context.Context, u model.UserID, sbj model.Subject, epsUpdate uint32, volsUpdate uint32)) *TimelineService_ChangeSubjectProgress_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.UserID), args[2].(model.Subject), args[3].(uint32), args[4].(uint32))
	})
	return _c
}

func (_c *TimelineService_ChangeSubjectProgress_Call) Return(err error) *TimelineService_ChangeSubjectProgress_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *TimelineService_ChangeSubjectProgress_Call) RunAndReturn(run func(ctx context.Context, u model.UserID, sbj model.Subject, epsUpdate uint32, volsUpdate uint32) error) *TimelineService_ChangeSubjectProgress_Call {
	_c.Call.Return(run)
	return _c
}
