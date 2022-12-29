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

	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dam"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/group"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/notification"
	"github.com/bangumi/server/internal/oauth"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/timeline"
	"github.com/bangumi/server/internal/topic"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web"
	"github.com/bangumi/server/web/captcha"
	"github.com/bangumi/server/web/frontend"
	"github.com/bangumi/server/web/handler"
	"github.com/bangumi/server/web/rate"
	"github.com/bangumi/server/web/session"
)

type Mock struct {
	SubjectRepo        subject.Repo
	PersonRepo         person.Repo
	CharacterRepo      character.Repo
	AuthRepo           auth.Repo
	AuthService        auth.Service
	EpisodeRepo        episode.Repo
	TopicRepo          topic.Repo
	GroupRepo          group.Repo
	UserRepo           user.Repo
	IndexRepo          index.Repo
	RevisionRepo       revision.Repo
	CollectionRepo     collection.Repo
	TimeLineRepo       timeline.Repo
	CaptchaManager     captcha.Manager
	SessionManager     session.Manager
	Cache              cache.RedisCache
	RateLimiter        rate.Manager
	OAuthManager       oauth.Manager
	PrivateMessageRepo pm.Repo
	NotificationRepo   notification.Repo
	HTTPMock           *httpmock.MockTransport
	Dam                *dam.Dam
}

//nolint:funlen
func GetWebApp(tb testing.TB, m Mock) *fiber.App {
	tb.Helper()
	var f *fiber.App

	httpClient := resty.New().SetJSONEscapeHTML(false)
	httpClient.JSONUnmarshal = sonic.Unmarshal
	httpClient.JSONMarshal = sonic.Marshal

	var options = []fx.Option{
		fx.NopLogger,

		handler.Module, ctrl.Module,

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
		fx.Provide(func() topic.Repo { return m.TopicRepo }),
		MockAuthRepo(m.AuthRepo),
		MockOAuthManager(m.OAuthManager),
		MockAuthService(m.AuthService),
		MockUserRepo(m.UserRepo),
		MockIndexRepo(m.IndexRepo),
		MockRevisionRepo(m.RevisionRepo),
		MockPrivateMessageRepo(m.PrivateMessageRepo),
		MockNoticationRepo(m.NotificationRepo),
		MockCaptchaManager(m.CaptchaManager),
		MockSessionManager(m.SessionManager),
		MockRateLimiter(m.RateLimiter),
		MockTimeLineRepo(m.TimeLineRepo),

		// don't need a default mock for these repositories.
		fx.Provide(func() group.Repo { return m.GroupRepo }),
		fx.Provide(func() collection.Repo { return m.CollectionRepo }),
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

func MockUserRepo(repo user.Repo) fx.Option {
	if repo == nil {
		mocker := &mocks.UserRepo{}
		mocker.EXPECT().GetByID(mock.Anything, mock.Anything).Return(user.User{}, nil)
		mocker.On("GetByIDs", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, ids []model.UserID) map[model.UserID]user.User {
				var ret = make(map[model.UserID]user.User, len(ids))
				for _, id := range ids {
					ret[id] = user.User{}
				}
				return ret
			}, func(ctx context.Context, ids []model.UserID) error {
				return nil
			})
		mocker.On("GetFieldsByIDs", mock.Anything, mock.Anything, mock.Anything).
			Return(func(ctx context.Context, ids []model.UserID) map[model.UserID]user.Fields {
				var ret = make(map[model.UserID]user.Fields, len(ids))
				for _, id := range ids {
					ret[id] = user.Fields{}
				}
				return ret
			}, func(ctx context.Context, ids []model.UserID) error {
				return nil
			})
		repo = mocker
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

func MockOAuthManager(m oauth.Manager) fx.Option {
	if m == nil {
		m = &mocks.OAuthManger{}
	}

	return fx.Provide(func() oauth.Manager { return m })
}

func MockSubjectRepo(m subject.Repo) fx.Option {
	if m == nil {
		mocker := &mocks.SubjectRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(model.Subject{}, nil)

		m = mocker
	}

	return fx.Provide(func() subject.Repo { return m })
}

func MockTimeLineRepo(m timeline.Repo) fx.Option {
	if m == nil {
		mocker := &mocks.TimeLineRepo{}
		mocker.EXPECT().WithQuery(mock.Anything).Return(mocker)
		mocker.EXPECT().Create(mock.Anything, mock.Anything).Return(nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(timeline.Repo))))
}

func MockCache(mock cache.RedisCache) fx.Option {
	return fx.Supply(fx.Annotate(mock, fx.As(new(cache.RedisCache))))
}

func MockEmptyCache() fx.Option {
	return fx.Provide(cache.NewNoop)
}
