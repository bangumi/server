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

package group

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.GroupRepo, error) {
	return mysqlRepo{q: q, log: log.Named("group.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) GetByName(ctx context.Context, name string) (model.Group, error) {
	g, err := r.q.Group.WithContext(ctx).Where(r.q.Group.Name.Eq(name)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Group{}, domain.ErrNotFound
		}

		r.log.Error("un-expected error when getting single group", zap.Error(err), zap.String("group_name", name))
		return model.Group{}, errgo.Wrap(err, "dal")
	}

	return convertDao(g), nil
}

func (r mysqlRepo) GetByID(ctx context.Context, id model.GroupID) (model.Group, error) {
	g, err := r.q.Group.WithContext(ctx).Where(r.q.Group.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Group{}, domain.ErrNotFound
		}

		r.log.Error("un-expected error when getting single group", zap.Error(err), log.GroupID(id))
		return model.Group{}, errgo.Wrap(err, "dal")
	}

	return convertDao(g), nil
}

func (r mysqlRepo) CountMembersByName(ctx context.Context, name string) (int64, error) {
	g, err := r.GetByName(ctx, name)
	if err != nil {
		return 0, err
	}

	return r.CountMembersByID(ctx, g.ID)
}

func (r mysqlRepo) CountMembersByID(ctx context.Context, id model.GroupID) (int64, error) {
	c, err := r.q.GroupMember.WithContext(ctx).Where(r.q.GroupMember.GroupID.Eq(id)).Count()
	if err != nil {
		r.log.Error("un-expected error when counting group member", zap.Error(err), log.GroupID(id))
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (r mysqlRepo) ListMembersByID(
	ctx context.Context, id model.GroupID, limit, offset int,
) ([]model.GroupMember, error) {
	c, err := r.q.GroupMember.WithContext(ctx).Where(r.q.GroupMember.GroupID.Eq(id)).Limit(limit).Offset(offset).Find()
	if err != nil {
		r.log.Error("un-expected error when counting group member", zap.Error(err), log.GroupID(id))
		return nil, errgo.Wrap(err, "dal")
	}

	var members = make([]model.GroupMember, len(c))

	for i, member := range c {
		members[i] = model.GroupMember{UserID: member.UserID}
	}

	return members, nil
}

func (r mysqlRepo) ListMembersByName(
	ctx context.Context, name string, limit, offset int,
) ([]model.GroupMember, error) {
	g, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return r.ListMembersByID(ctx, g.ID, limit, offset)
}

func convertDao(g *dao.Group) model.Group {
	return model.Group{
		Name:        g.Name,
		NSFW:        g.Nsfw,
		ID:          g.ID,
		Description: g.Description,
		CreatedAt:   time.Unix(int64(g.CreatedAt), 0),
	}
}
