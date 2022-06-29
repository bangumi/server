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

var _ domain.GroupRepo = mysqlRepo{}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.GroupRepo, error) {
	return mysqlRepo{q: q, log: log.Named("group.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) getIDByName(ctx context.Context, name string) (model.GroupID, error) {
	g, err := r.q.Group.WithContext(ctx).Where(r.q.Group.Name.Eq(name)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, domain.ErrNotFound
		}

		r.log.Error("un-expected error when getting single group", zap.Error(err), zap.String("group_name", name))
		return 0, errgo.Wrap(err, "dal")
	}

	return g.ID, nil
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

func (r mysqlRepo) CountMembersByName(
	ctx context.Context, name string, memberType domain.GroupMemberType,
) (int64, error) {
	id, err := r.getIDByName(ctx, name)
	if err != nil {
		return 0, err
	}

	return r.countMembersByID(ctx, id, memberType)
}

func (r mysqlRepo) countMembersByID(
	ctx context.Context, id model.GroupID, memberType domain.GroupMemberType,
) (int64, error) {
	q := r.q.GroupMember.WithContext(ctx).Where(r.q.GroupMember.GroupID.Eq(id))
	switch memberType {
	case domain.GroupMemberMod:
		q = q.Where(r.q.GroupMember.Moderator.Is(true))
	case domain.GroupMemberNormal:
		q = q.Where(r.q.GroupMember.Moderator.Is(false))
	case domain.GroupMemberAll:
		// do nothing
	}

	c, err := q.Count()
	if err != nil {
		r.log.Error("un-expected error when counting group member", zap.Error(err), log.GroupID(id))
		return 0, errgo.Wrap(err, "dal")
	}

	return c, nil
}

func (r mysqlRepo) listMembersByID(
	ctx context.Context, id model.GroupID, memberType domain.GroupMemberType, limit, offset int,
) ([]model.GroupMember, error) {
	q := r.q.GroupMember.WithContext(ctx).Where(r.q.GroupMember.GroupID.Eq(id))

	switch memberType {
	case domain.GroupMemberMod:
		q = q.Where(r.q.GroupMember.Moderator.Is(true))
	case domain.GroupMemberNormal:
		q = q.Where(r.q.GroupMember.Moderator.Is(false))
	case domain.GroupMemberAll:
		// do nothing
	}

	c, err := q.Limit(limit).Offset(offset).Order(r.q.GroupMember.CreatedTime.Desc()).Find()
	if err != nil {
		r.log.Error("un-expected error when counting group member", zap.Error(err), log.GroupID(id))
		return nil, errgo.Wrap(err, "dal")
	}

	members := make([]model.GroupMember, len(c))
	for i, member := range c {
		members[i] = convertMember(member)
	}

	return members, nil
}

func (r mysqlRepo) ListMembersByName(
	ctx context.Context, name string, memberType domain.GroupMemberType, limit, offset int,
) ([]model.GroupMember, error) {
	id, err := r.getIDByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return r.listMembersByID(ctx, id, memberType, limit, offset)
}

func (r mysqlRepo) CountMembersByID(
	ctx context.Context, id model.GroupID, memberType domain.GroupMemberType,
) (int64, error) {
	return r.countMembersByID(ctx, id, memberType)
}

func (r mysqlRepo) ListMembersByID(
	ctx context.Context, id model.GroupID, memberType domain.GroupMemberType, limit, offset int,
) ([]model.GroupMember, error) {
	return r.listMembersByID(ctx, id, memberType, limit, offset)
}

func convertDao(g *dao.Group) model.Group {
	return model.Group{
		Name:        g.Name,
		NSFW:        g.Nsfw,
		ID:          g.ID,
		Description: g.Description,
		Icon:        g.Icon,
		MemberCount: int64(g.Members),
		Title:       g.Title,
		CreatedAt:   time.Unix(int64(g.CreatedTime), 0),
	}
}

func convertMember(m *dao.GroupMember) model.GroupMember {
	return model.GroupMember{
		UserID: m.UserID,
		Mod:    m.Moderator,
		JoinAt: time.Unix(int64(m.CreatedTime), 0),
	}
}
