// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package test

import (
	"encoding/json"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/collections"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/notification"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/timeline"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web"
	"github.com/bangumi/server/web/handler"
	"github.com/bangumi/server/web/session"
)

type Mock struct {
	SubjectRepo        subject.Repo
	SubjectCachedRepo  subject.CachedRepo
	PersonRepo         person.Repo
	CharacterRepo      character.Repo
	AuthRepo           auth.Repo
	AuthService        auth.Service
	EpisodeRepo        episode.Repo
	UserRepo           user.Repo
	IndexRepo          index.Repo
	RevisionRepo       revision.Repo
	CollectionRepo     collections.Repo
	TimeLineSrv        timeline.Service
	SessionManager     session.Manager
	Cache              cache.RedisCache
	PrivateMessageRepo pm.Repo
	NotificationRepo   notification.Repo
	HTTPMock           *httpmock.MockTransport
	Dam                *dam.Dam
}

//nolint:funlen
func GetWebApp(tb testing.TB, m Mock) *echo.Echo {
	tb.Helper()
	var e *echo.Echo

	httpClient := resty.New().SetJSONEscapeHTML(false)
	httpClient.JSONUnmarshal = json.Unmarshal
	httpClient.JSONMarshal = json.Marshal

	var options = []fx.Option{
		fx.NopLogger,

		handler.Module, ctrl.Module,

		fx.Provide(func() dal.Transaction { return dal.NoopTransaction{} }),

		fx.Supply(httpClient),

		fx.Provide(
			logger.Copy, config.NewAppConfig, web.NewTestingApp,
			person.NewService,
		),

		MockPersonRepo(m.PersonRepo),
		MockCharacterRepo(m.CharacterRepo),
		MockSubjectRepo(m.SubjectRepo),
		MockSubjectReadRepo(m.SubjectCachedRepo),
		MockEpisodeRepo(m.EpisodeRepo),
		MockAuthRepo(m.AuthRepo),
		MockAuthService(m.AuthService),
		MockUserRepo(m.UserRepo),
		MockIndexRepo(m.IndexRepo),
		MockRevisionRepo(m.RevisionRepo),
		MockPrivateMessageRepo(m.PrivateMessageRepo),
		MockNoticationRepo(m.NotificationRepo),
		MockSessionManager(m.SessionManager),
		MockTimeLineSrv(m.TimeLineSrv),

		// don't need a default mock for these repositories.
		fx.Provide(func() collections.Repo { return m.CollectionRepo }),
		fx.Provide(func() search.Handler { return search.NoopClient{} }),

		fx.Invoke(web.AddRouters),

		fx.Populate(&e),
	}

	if m.Dam != nil {
		options = append(options, fx.Supply(*m.Dam))
	} else {
		options = append(options, fx.Provide(dam.New))
	}

	if m.Cache == nil {
		options = append(options, MockEmptyCache())
	} else {
		options = append(options, MockCache(m.Cache))
	}

	if err := fx.New(options...).Err(); err != nil {
		tb.Fatal("can't create web app", err)
	}

	if m.HTTPMock != nil {
		httpClient.GetClient().Transport = m.HTTPMock
	}

	return e
}

func MockRevisionRepo(repo revision.Repo) fx.Option {
	if repo == nil {
		repo = &mocks.RevisionRepo{}
	}
	return fx.Supply(fx.Annotate(repo, fx.As(new(revision.Repo))))
}

func MockIndexRepo(repo index.Repo) fx.Option {
	if repo == nil {
		mocker := &mocks.IndexRepo{}

		repo = mocker
	}

	return fx.Supply(fx.Annotate(repo, fx.As(new(index.Repo))))
}

func MockPrivateMessageRepo(repo pm.Repo) fx.Option {
	if repo == nil {
		repo = &mocks.PrivateMessageRepo{}
	}
	return fx.Supply(fx.Annotate(repo, fx.As(new(pm.Repo))))
}

func MockNoticationRepo(repo notification.Repo) fx.Option {
	if repo == nil {
		repo = &mocks.NotificationRepo{}
	}
	return fx.Supply(fx.Annotate(repo, fx.As(new(notification.Repo))))
}

func MockSessionManager(repo session.Manager) fx.Option {
	if repo == nil {
		mocker := &mocks.SessionManager{}
		mocker.EXPECT().Create(mock.Anything, mock.Anything).Return("mocked random string", session.Session{}, nil)
		mocker.EXPECT().Get(mock.Anything, mock.Anything).Return(session.Session{}, nil)

		repo = mocker
	}

	return fx.Provide(func() session.Manager { return repo })
}

func MockUserRepo(repo user.Repo) fx.Option {
	if repo == nil {
		return fx.Provide(AnyUserMock)
	}

	return fx.Supply(fx.Annotate(repo, fx.As(new(user.Repo))))
}

func MockPersonRepo(m person.Repo) fx.Option {
	if m == nil {
		mocker := &mocks.PersonRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Person{}, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(person.Repo))))
}

func MockCharacterRepo(m character.Repo) fx.Option {
	if m == nil {
		mocker := &mocks.CharacterRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Character{}, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(character.Repo))))
}

func MockEpisodeRepo(m episode.Repo) fx.Option {
	if m == nil {
		mocker := &mocks.EpisodeRepo{}
		mocker.EXPECT().Count(mock.Anything, mock.Anything, mock.Anything).Return(0, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(episode.Repo))))
}

func MockAuthRepo(m auth.Repo) fx.Option {
	if m == nil {
		mocker := &mocks.AuthRepo{}
		mocker.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.UserInfo{}, nil)
		mocker.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(auth.Permission{}, nil)

		m = mocker
	}

	return fx.Provide(func() auth.Repo { return m })
}

func MockAuthService(m auth.Service) fx.Option {
	if m == nil {
		return fx.Provide(auth.NewService)
	}

	return fx.Provide(func() auth.Service { return m })
}

func MockSubjectRepo(m subject.Repo) fx.Option {
	if m == nil {
		mocker := &mocks.SubjectRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(model.Subject{}, nil)

		m = mocker
	}

	return fx.Provide(func() subject.Repo { return m })
}

func MockSubjectReadRepo(m subject.CachedRepo) fx.Option {
	if m == nil {
		mocker := &mocks.SubjectCachedRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(model.Subject{}, nil)

		m = mocker
	}

	return fx.Provide(func() subject.CachedRepo { return m })
}

func MockTimeLineSrv(m timeline.Service) fx.Option {
	if m == nil {
		mocker := &mocks.TimeLineService{}

		mocker.EXPECT().ChangeSubjectCollection(mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mocker.EXPECT().ChangeEpisodeStatus(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mocker.EXPECT().ChangeSubjectProgress(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(timeline.Service))))
}

func MockCache(mock cache.RedisCache) fx.Option {
	return fx.Supply(fx.Annotate(mock, fx.As(new(cache.RedisCache))))
}

func MockEmptyCache() fx.Option {
	return fx.Provide(cache.NewNoop)
}
