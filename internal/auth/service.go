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

package auth

import (
	"context" //nolint:gosec
	"time"

	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth/internal/cachekey"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/user"
)

const TokenTypeOauthToken = 0
const TokenTypeAccessToken = 1

func NewService(repo Repo, u user.Repo, logger *zap.Logger, c cache.RedisCache) Service {
	return service{
		permCache: cache.NewMemoryCache[user.GroupID, Permission](),
		cache:     c,
		repo:      repo,
		log:       logger.Named("auth.Service"),
		user:      u,
	}
}

type service struct {
	permCache *cache.MemoryCache[user.GroupID, Permission]
	cache     cache.RedisCache
	repo      Repo
	user      user.Repo
	log       *zap.Logger
}

func (s service) GetByToken(ctx context.Context, token string) (Auth, error) {
	var a UserInfo
	var cacheKey = cachekey.Auth(token)

	ok, err := s.cache.Get(ctx, cacheKey, &a)
	if err != nil {
		return Auth{}, errgo.Wrap(err, "cache.Get")
	}

	if !ok {
		a, err = s.repo.GetByToken(ctx, token)
		if err != nil {
			return Auth{}, errgo.Wrap(err, "AuthRepo.GetByID")
		}

		_ = s.cache.Set(ctx, cacheKey, a, time.Minute*10)
	}

	permission, err := s.getPermission(ctx, a.GroupID)
	if err != nil {
		return Auth{}, err
	}

	return Auth{
		Login:   true,
		RegTime: a.RegTime,
		ID:      a.ID,
		GroupID: a.GroupID,
		Permission: Permission{
			UserBan: permission.UserBan || a.Permission.UserBan,
			BanPost: permission.BanPost || a.Permission.BanPost,

			UserList:           permission.UserList && a.Permission.UserList,
			ManageUserGroup:    permission.ManageUserGroup && a.Permission.ManageUserGroup,
			ManageUserPhoto:    permission.ManageUserPhoto && a.Permission.ManageUserPhoto,
			ManageTopicState:   permission.ManageTopicState && a.Permission.ManageTopicState,
			ManageReport:       permission.ManageReport && a.Permission.ManageReport,
			ManageUser:         permission.ManageUser && a.Permission.ManageUser,
			UserGroup:          permission.UserGroup && a.Permission.UserGroup,
			UserWikiApply:      permission.UserWikiApply && a.Permission.UserWikiApply,
			UserWikiApprove:    permission.UserWikiApprove && a.Permission.UserWikiApprove,
			DoujinSubjectErase: permission.DoujinSubjectErase && a.Permission.DoujinSubjectErase,
			DoujinSubjectLock:  permission.DoujinSubjectLock && a.Permission.DoujinSubjectLock,
			SubjectEdit:        permission.SubjectEdit && a.Permission.SubjectEdit,
			SubjectLock:        permission.SubjectLock && a.Permission.SubjectLock,
			SubjectRefresh:     permission.SubjectRefresh && a.Permission.SubjectRefresh,
			SubjectRelated:     permission.SubjectRelated && a.Permission.SubjectRelated,
			SubjectMerge:       permission.SubjectMerge && a.Permission.SubjectMerge,
			SubjectErase:       permission.SubjectErase && a.Permission.SubjectErase,
			SubjectCoverLock:   permission.SubjectCoverLock && a.Permission.SubjectCoverLock,
			SubjectCoverErase:  permission.SubjectCoverErase && a.Permission.SubjectCoverErase,
			MonoEdit:           permission.MonoEdit && a.Permission.MonoEdit,
			MonoLock:           permission.MonoLock && a.Permission.MonoLock,
			MonoMerge:          permission.MonoMerge && a.Permission.MonoMerge,
			MonoErase:          permission.MonoErase && a.Permission.MonoErase,
			EpEdit:             permission.EpEdit && a.Permission.EpEdit,
			EpMove:             permission.EpMove && a.Permission.EpMove,
			EpMerge:            permission.EpMerge && a.Permission.EpMerge,
			EpLock:             permission.EpLock && a.Permission.EpLock,
			EpErase:            permission.EpErase && a.Permission.EpErase,
			Report:             permission.Report && a.Permission.Report,
			ManageApp:          permission.ManageApp && a.Permission.ManageApp,
			AppErase:           permission.AppErase && a.Permission.AppErase,
		},
	}, nil
}

func (s service) getPermission(ctx context.Context, id user.GroupID) (Permission, error) {
	p, ok := s.permCache.Get(ctx, id)

	if ok {
		return p, nil
	}

	p, err := s.repo.GetPermission(ctx, id)
	if err != nil {
		return Permission{}, errgo.Wrap(err, "AuthRepo.GetPermission")
	}

	s.permCache.Set(ctx, id, p, time.Minute)

	return p, nil
}
