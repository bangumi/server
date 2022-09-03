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
	"context"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dam"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/oauth"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/web"
	"github.com/bangumi/server/internal/web/captcha"
	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/handler"
	"github.com/bangumi/server/internal/web/rate"
	"github.com/bangumi/server/internal/web/session"
)

type Mock struct {
	SubjectRepo    domain.SubjectRepo
	PersonRepo     domain.PersonRepo
	CharacterRepo  domain.CharacterRepo
	AuthRepo       domain.AuthRepo
	AuthService    domain.AuthService
	EpisodeRepo    domain.EpisodeRepo
	TopicRepo      domain.TopicRepo
	GroupRepo      domain.GroupRepo
	UserRepo       domain.UserRepo
	IndexRepo      domain.IndexRepo
	RevisionRepo   domain.RevisionRepo
	CollectionRepo domain.CollectionRepo
	TimeLineRepo   domain.TimeLineRepo
	CaptchaManager captcha.Manager
	SessionManager session.Manager
	Cache          cache.Cache
	RateLimiter    rate.Manager
	OAuthManager   oauth.Manager
	HTTPMock       *httpmock.MockTransport
	Dam            *dam.Dam
}

//nolint:funlen
func GetWebApp(tb testing.TB, m Mock) *fiber.App {
	tb.Helper()
	var f *fiber.App

	httpClient := resty.New().SetJSONEscapeHTML(false)
	httpClient.JSONUnmarshal = json.Unmarshal
	httpClient.JSONMarshal = json.Marshal

	var options = []fx.Option{
		fx.NopLogger,

		handler.Module, ctrl.Module,

		fx.Provide(func() tally.Scope { return tally.NoopScope }),
		fx.Supply(fx.Annotate(promreporter.NewReporter(promreporter.Options{}), fx.As(new(promreporter.Reporter)))),
		fx.Provide(func() dal.Transaction { return dal.NoopTransaction{} }),

		fx.Supply(httpClient),

		fx.Provide(
			logger.Copy, config.NewAppConfig, dal.NewDB, web.New,
			person.NewService, frontend.NewTemplateEngine,
		),

		MockPersonRepo(m.PersonRepo),
		MockCharacterRepo(m.CharacterRepo),
		MockSubjectRepo(m.SubjectRepo),
		MockEpisodeRepo(m.EpisodeRepo),
		fx.Provide(func() domain.TopicRepo { return m.TopicRepo }),
		MockAuthRepo(m.AuthRepo),
		MockOAuthManager(m.OAuthManager),
		MockAuthService(m.AuthService),
		MockUserRepo(m.UserRepo),
		MockIndexRepo(m.IndexRepo),
		MockRevisionRepo(m.RevisionRepo),
		MockCaptchaManager(m.CaptchaManager),
		MockSessionManager(m.SessionManager),
		MockRateLimiter(m.RateLimiter),
		MockTimeLineRepo(m.TimeLineRepo),

		// don't need a default mock for these repositories.
		fx.Provide(func() domain.GroupRepo { return m.GroupRepo }),
		fx.Provide(func() domain.CollectionRepo { return m.CollectionRepo }),
		fx.Provide(func() search.Handler { return search.NoopClient{} }),

		fx.Invoke(web.AddRouters),

		fx.Populate(&f),
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

	return f
}

func MockRevisionRepo(repo domain.RevisionRepo) fx.Option {
	if repo == nil {
		repo = &mocks.RevisionRepo{}
	}
	return fx.Supply(fx.Annotate(repo, fx.As(new(domain.RevisionRepo))))
}

func MockIndexRepo(repo domain.IndexRepo) fx.Option {
	if repo == nil {
		mocker := &mocks.IndexRepo{}

		repo = mocker
	}

	return fx.Supply(fx.Annotate(repo, fx.As(new(domain.IndexRepo))))
}

func MockRateLimiter(repo rate.Manager) fx.Option {
	if repo == nil {
		mocker := &mocks.RateLimiter{}
		mocker.EXPECT().Login(mock.Anything, mock.Anything).Return(true, 5, nil) //nolint:gomnd
		mocker.EXPECT().Reset(mock.Anything, mock.Anything).Return(nil)

		repo = mocker
	}

	return fx.Provide(func() rate.Manager { return repo })
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

func MockCaptchaManager(repo captcha.Manager) fx.Option {
	if repo == nil {
		mocker := &mocks.CaptchaManager{}
		mocker.EXPECT().Verify(mock.Anything, mock.Anything).Return(true, nil)

		repo = mocker
	}

	return fx.Provide(func() captcha.Manager { return repo })
}

func MockUserRepo(repo domain.UserRepo) fx.Option {
	if repo == nil {
		mocker := &mocks.UserRepo{}
		mocker.EXPECT().GetByID(mock.Anything, mock.Anything).Return(model.User{}, nil)
		mocker.On("GetByIDs", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, ids []model.UserID) map[model.UserID]model.User {
				var ret = make(map[model.UserID]model.User, len(ids))
				for _, id := range ids {
					ret[id] = model.User{}
				}
				return ret
			}, func(ctx context.Context, ids []model.UserID) error {
				return nil
			})
		repo = mocker
	}

	return fx.Supply(fx.Annotate(repo, fx.As(new(domain.UserRepo))))
}

func MockPersonRepo(m domain.PersonRepo) fx.Option {
	if m == nil {
		mocker := &mocks.PersonRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Person{}, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.PersonRepo))))
}

func MockCharacterRepo(m domain.CharacterRepo) fx.Option {
	if m == nil {
		mocker := &mocks.CharacterRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Character{}, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.CharacterRepo))))
}

func MockEpisodeRepo(m domain.EpisodeRepo) fx.Option {
	if m == nil {
		mocker := &mocks.EpisodeRepo{}
		mocker.EXPECT().Count(mock.Anything, mock.Anything, mock.Anything).Return(0, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.EpisodeRepo))))
}

func MockAuthRepo(m domain.AuthRepo) fx.Option {
	if m == nil {
		mocker := &mocks.AuthRepo{}
		mocker.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.AuthUserInfo{}, nil)
		mocker.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(domain.Permission{}, nil)

		m = mocker
	}

	return fx.Provide(func() domain.AuthRepo { return m })
}

func MockAuthService(m domain.AuthService) fx.Option {
	if m == nil {
		return fx.Provide(auth.NewService)
	}

	return fx.Provide(func() domain.AuthService { return m })
}

func MockOAuthManager(m oauth.Manager) fx.Option {
	if m == nil {
		m = &mocks.OAuthManger{}
	}

	return fx.Provide(func() oauth.Manager { return m })
}

func MockSubjectRepo(m domain.SubjectRepo) fx.Option {
	if m == nil {
		mocker := &mocks.SubjectRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Subject{}, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.SubjectRepo))))
}

func MockTimeLineRepo(m domain.TimeLineRepo) fx.Option {
	if m == nil {
		mocker := &mocks.TimeLineRepo{}
		mocker.EXPECT().WithQuery(mock.Anything).Return(mocker)
		mocker.EXPECT().Create(mock.Anything, mock.Anything).Return(nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.TimeLineRepo))))
}

func MockCache(mock cache.Cache) fx.Option {
	return fx.Supply(fx.Annotate(mock, fx.As(new(cache.Cache))))
}

func MockEmptyCache() fx.Option {
	return fx.Provide(NopCache)
}

func NopCache() cache.Cache {
	mc := &mocks.Cache{}
	mc.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	mc.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mc.EXPECT().Del(mock.Anything, mock.Anything).Return(nil)

	return mc
}
