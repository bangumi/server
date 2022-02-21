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
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"

	"github.com/bangumi/server/auth"
	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/config"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/subject"
	"github.com/bangumi/server/web"
)

func GetWebApp(t *testing.T, options ...fx.Option) (f *fiber.App) {
	t.Helper()

	options = append(options,
		// fx.NopLogger,

		fx.Supply(fx.Annotate(tally.NoopScope, fx.As(new(tally.Scope)))),
		fx.Supply(fx.Annotate(promreporter.NewReporter(promreporter.Options{}),
			fx.As(new(promreporter.Reporter)))),

		fx.Provide(
			logger.Copy,
			config.NewAppConfig,
			dal.NewDB,
			web.New,
			web.NewHandle,
		),

		fx.Invoke(web.ResistRouter),

		fx.Populate(&f),
	)

	app := fx.New(options...)

	if app.Err() != nil {
		t.Fatal("can't create web app", app.Err())
	}

	return
}

func MockEpisodeRepo(m domain.EpisodeRepo) fx.Option {
	if m == nil {
		mocker := &domain.MockEpisodeRepo{}
		mocker.EXPECT().Count(mock.Anything, mock.Anything).Return(0, nil)

		m = mocker
	}

	return fx.Supply(fx.Annotate(m, fx.As(new(domain.EpisodeRepo))))
}

func MockUserRepo(mock domain.AuthRepo) fx.Option {
	return fx.Supply(fx.Annotate(mock, fx.As(new(domain.AuthRepo))))
}

func MockSubjectRepo(mock domain.SubjectRepo) fx.Option {
	return fx.Supply(fx.Annotate(mock, fx.As(new(domain.SubjectRepo))))
}

func MockCache(mock cache.Generic) fx.Option {
	return fx.Supply(fx.Annotate(mock, fx.As(new(cache.Generic))))
}

func MockEmptyCache() fx.Option {
	mc := &cache.MockGeneric{}
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
