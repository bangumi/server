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

package tag

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/model"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger, db *sqlx.DB) (Repo, error) {
	return mysqlRepo{q: q, log: log.Named("tag.mysqlRepo"), db: db}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
	db  *sqlx.DB
}

func (r mysqlRepo) Get(ctx context.Context, id model.SubjectID) ([]Tag, error) {
	var s []struct {
		Tid        uint   `db:"tlt_tid"`
		Name       string `db:"tag_name"`
		TotalCount uint   `db:"tag_results"`
	}

	err := r.db.SelectContext(ctx, &s, `
		select tlt_tid, tag_name, tag_results
		from chii_tag_neue_list
		inner join chii_tag_neue_index on tlt_tid = tag_id and tlt_type = tag_type
		where tlt_uid = 0 and tag_cat = ? and tlt_mid = ?
		`, CatSubject, id)
	if err != nil {
		return nil, err
	}

	tags := make([]Tag, len(s))
	for i, t := range s {
		tags[i] = Tag{
			Name:       t.Name,
			TotalCount: t.TotalCount,
		}
	}

	return tags, nil
}

func (r mysqlRepo) GetByIDs(ctx context.Context, ids []model.SubjectID) (map[model.SubjectID][]Tag, error) {
	var s []struct {
		Tid        uint            `db:"tlt_tid"`
		Name       string          `db:"tag_name"`
		TotalCount uint            `db:"tag_results"`
		Mid        model.SubjectID `db:"tlt_mid"`
	}

	q, v, err := sqlx.In(`
		select tlt_tid, tag_name, tag_results, tlt_mid
		from chii_tag_neue_list
		inner join chii_tag_neue_index on tlt_tid = tag_id and tlt_type = tag_type
		where tlt_uid = 0 and tag_cat = ? and tlt_mid IN (?)
		`, CatSubject, ids)
	if err != nil {
		return nil, err
	}

	err = r.db.SelectContext(ctx, &s, q, v...)
	if err != nil {
		return nil, err
	}

	tags := make(map[model.SubjectID][]Tag, len(s))
	for _, t := range s {
		tags[t.Mid] = append(tags[t.Mid], Tag{
			Name:       t.Name,
			TotalCount: t.TotalCount,
		})
	}

	return tags, nil
}
