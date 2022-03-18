// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/auth"
	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/character"
	"github.com/bangumi/server/config"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/person"
	"github.com/bangumi/server/subject"
	"github.com/bangumi/server/web"
	"github.com/bangumi/server/web/handler"
)

type Mock struct {
	SubjectRepo   domain.SubjectRepo
	PersonRepo    domain.PersonRepo
	CharacterRepo domain.CharacterRepo
	AuthRepo      domain.AuthRepo
	EpisodeRepo   domain.EpisodeRepo
	UserRepo      domain.UserRepo
	IndexRepo     domain.IndexRepo
	RevisionRepo  domain.RevisionRepo
	Cache         cache.Generic
}
type TB interface {
	Helper()
	Fatal(args ...interface{})
}

func GetWebApp(t TB, m Mock) *fiber.App {
	t.Helper()
	var f *fiber.App

	var options = []fx.Option{
		fx.NopLogger,

		fx.Supply(fx.Annotate(tally.NoopScope, fx.As(new(tally.Scope)))),
		fx.Supply(fx.Annotate(promreporter.NewReporter(promreporter.Options{}),
			fx.As(new(promreporter.Reporter)))),

		fx.Provide(
			zap.NewNop,
			config.NewAppConfig,
			dal.NewDB,
			web.New,
			handler.New,
		),

		fx.Provide(
			character.NewService,
			subject.NewService,
			person.NewService,
		),

		MockPersonRepo(m.PersonRepo),
		MockCharacterRepo(m.CharacterRepo),
		MockSubjectRepo(m.SubjectRepo),
		MockEpisodeRepo(m.EpisodeRepo),
		MockAuthRepo(m.AuthRepo),
		MockUserRepo(m.UserRepo),
		MockIndexRepo(m.IndexRepo),
		MockRevisionRepo(m.RevisionRepo),

		fx.Invoke(web.ResistRouter),

		fx.Populate(&f),
	}

	if m.Cache == nil {
		options = append(options, MockEmptyCache())
	} else {
		options = append(options, MockCache(m.Cache))
	}

	app := fx.New(options...)

	if app.Err() != nil {
		t.Fatal("can't create web app", app.Err())
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

func MockUserRepo(repo domain.UserRepo) fx.Option {
	if repo == nil {
		mocker := &mocks.UserRepo{}
		mocker.EXPECT().GetByID(mock.Anything, mock.Anything).Return(model.User{}, nil)
		mocker.On("GetByIDs", mock.Anything, mock.Anything).
			Return(func(ctx context.Context, ids ...uint32) map[uint32]model.User {
				var ret = make(map[uint32]model.User, len(ids))
				for _, id := range ids {
					ret[id] = model.User{}
				}
				return ret
			}, func(ctx context.Context, ids ...uint32) error {
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
		mocker.EXPECT().Count(mock.Anything, mock.Anything).Return(0, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.EpisodeRepo))))
}

func MockAuthRepo(m domain.AuthRepo) fx.Option {
	if m == nil {
		mocker := &mocks.AuthRepo{}
		mocker.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(domain.Auth{}, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.AuthRepo))))
}

func MockSubjectRepo(m domain.SubjectRepo) fx.Option {
	if m == nil {
		mocker := &mocks.SubjectRepo{}
		mocker.EXPECT().Get(mock.Anything, mock.Anything).Return(model.Subject{}, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.SubjectRepo))))
}

func MockCache(mock cache.Generic) fx.Option {
	return fx.Supply(fx.Annotate(mock, fx.As(new(cache.Generic))))
}

func MockEmptyCache() fx.Option {
	mc := &mocks.Generic{}
	mc.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	mc.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	return fx.Supply(fx.Annotate(mc, fx.As(new(cache.Generic))))
}

func FxE2E(t *testing.T) fx.Option {
	t.Helper()

	return fx.Provide(
		query.Use,
		driver.NewRedisClient,
		auth.NewMysqlRepo,
		subject.NewMysqlRepo,
		dal.NewConnectionPool,
	)
}
