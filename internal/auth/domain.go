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
	"context"
	"time"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/user"
)

// AuthRepo presents an authorization.
type Repo interface {
	// GetByToken return an authorized user by a valid access token.
	GetByToken(ctx context.Context, token string) (UserInfo, error)
	GetPermission(ctx context.Context, groupID uint8) (Permission, error)
}

type Service interface {
	GetByToken(ctx context.Context, token string) (Auth, error)
}

type UserInfo struct {
	RegTime    time.Time
	ID         model.UserID
	GroupID    user.GroupID
	Permission Permission
	Scope      Scope
	Legacy     bool
}

// Auth is the basic authorization represent a user.
type Auth struct {
	Login      bool
	RegTime    time.Time
	ID         model.UserID // user id
	GroupID    user.GroupID
	Permission Permission
	Scope      Scope
	Legacy     bool
}

type Scope map[string]bool

func (u Auth) HasScope(s string) bool {
	if u.Legacy || u.Scope == nil {
		return true
	}

	return u.Scope[s]
}

const nsfwThreshold = gtime.OneDay * 60

// AllowNSFW return if current user is allowed to see NSFW resource.
func (u Auth) AllowNSFW() bool {
	return u.Login && u.RegisteredLongerThan(nsfwThreshold)
}

func (u Auth) RegisteredLongerThan(t time.Duration) bool {
	if u.ID == 0 {
		return false
	}

	return time.Since(u.RegTime) >= t
}

type AccessToken struct {
	ExpiredAt time.Time
	CreatedAt time.Time
	Name      string
	ClientID  string
	ID        uint32
	UserID    model.UserID
}

type Permission struct {
	UserList           bool
	ManageUserGroup    bool
	ManageUserPhoto    bool
	ManageTopicState   bool
	ManageReport       bool
	UserBan            bool
	ManageUser         bool
	UserGroup          bool
	UserWikiApply      bool `doc:"申请 wiki 人"`
	UserWikiApprove    bool
	DoujinSubjectErase bool
	DoujinSubjectLock  bool
	SubjectEdit        bool
	SubjectLock        bool
	SubjectRefresh     bool
	SubjectRelated     bool
	SubjectMerge       bool
	SubjectErase       bool
	SubjectCoverLock   bool
	SubjectCoverErase  bool
	MonoEdit           bool
	MonoLock           bool
	MonoMerge          bool
	MonoErase          bool
	BanPost            bool
	EpEdit             bool
	EpMove             bool
	EpMerge            bool
	EpLock             bool
	EpErase            bool
	Report             bool
	ManageApp          bool
	AppErase           bool
}

func (p Permission) Merge(other Permission) Permission {
	return Permission{
		UserBan: p.UserBan || other.UserBan,
		BanPost: p.BanPost || other.BanPost,

		UserList:           p.UserList && other.UserList,
		ManageUserGroup:    p.ManageUserGroup && other.ManageUserGroup,
		ManageUserPhoto:    p.ManageUserPhoto && other.ManageUserPhoto,
		ManageTopicState:   p.ManageTopicState && other.ManageTopicState,
		ManageReport:       p.ManageReport && other.ManageReport,
		ManageUser:         p.ManageUser && other.ManageUser,
		UserGroup:          p.UserGroup && other.UserGroup,
		UserWikiApply:      p.UserWikiApply && other.UserWikiApply,
		UserWikiApprove:    p.UserWikiApprove && other.UserWikiApprove,
		DoujinSubjectErase: p.DoujinSubjectErase && other.DoujinSubjectErase,
		DoujinSubjectLock:  p.DoujinSubjectLock && other.DoujinSubjectLock,
		SubjectEdit:        p.SubjectEdit && other.SubjectEdit,
		SubjectLock:        p.SubjectLock && other.SubjectLock,
		SubjectRefresh:     p.SubjectRefresh && other.SubjectRefresh,
		SubjectRelated:     p.SubjectRelated && other.SubjectRelated,
		SubjectMerge:       p.SubjectMerge && other.SubjectMerge,
		SubjectErase:       p.SubjectErase && other.SubjectErase,
		SubjectCoverLock:   p.SubjectCoverLock && other.SubjectCoverLock,
		SubjectCoverErase:  p.SubjectCoverErase && other.SubjectCoverErase,
		MonoEdit:           p.MonoEdit && other.MonoEdit,
		MonoLock:           p.MonoLock && other.MonoLock,
		MonoMerge:          p.MonoMerge && other.MonoMerge,
		MonoErase:          p.MonoErase && other.MonoErase,
		EpEdit:             p.EpEdit && other.EpEdit,
		EpMove:             p.EpMove && other.EpMove,
		EpMerge:            p.EpMerge && other.EpMerge,
		EpLock:             p.EpLock && other.EpLock,
		EpErase:            p.EpErase && other.EpErase,
		Report:             p.Report && other.Report,
		ManageApp:          p.ManageApp && other.ManageApp,
		AppErase:           p.AppErase && other.AppErase,
	}
}
