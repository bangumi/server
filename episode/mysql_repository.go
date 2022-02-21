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

package episode

import (
	"context"

	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.EpisodeRepo, error) {
	return mysqlRepo{q: q, log: log.Named("repository.mysqlRepo")}, nil
}

func (r mysqlRepo) Count(ctx context.Context, subjectID uint32) (int, error) {
	c, err := r.q.Episode.WithContext(ctx).
		Where(r.q.Episode.SubjectID.Eq(subjectID)).Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}

	return int(c), nil
}
