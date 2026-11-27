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
	// GetTokensByUser returns valid access tokens owned by the user.
	GetTokensByUser(ctx context.Context, userID model.UserID) ([]string, error)
}

type Service interface {
	GetByToken(ctx context.Context, token string) (Auth, error)
	// RevokeToken invalidates cached authorization for a token that has been expired or deleted in database.
	RevokeToken(ctx context.Context, token string) error
	// RevokeUserTokens invalidates cached authorization for every valid access token owned by the user.
	RevokeUserTokens(ctx context.Context, userID model.UserID) error
}

type UserInfo struct {
	RegTime    time.Time
	ID         model.UserID
	GroupID    user.GroupID
	Permission PersonalPermission
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
	BanPostGroup       bool
	BanPostGroupNSFW   bool
	BanSubjectInterest bool
	BanVisit           bool
	Deleted            bool
	EpEdit             bool
	EpMove             bool
	EpMerge            bool
	EpLock             bool
	EpErase            bool
	Report             bool
	ManageApp          bool
	AppErase           bool
}

// PersonalPermission is a partial permission set read from chii_members.acl.
// A nil field is not set and inherits the user group permission.
type PersonalPermission struct {
	UserList           *bool
	ManageUserGroup    *bool
	ManageUserPhoto    *bool
	ManageTopicState   *bool
	ManageReport       *bool
	UserBan            *bool
	ManageUser         *bool
	UserGroup          *bool
	UserWikiApply      *bool
	UserWikiApprove    *bool
	DoujinSubjectErase *bool
	DoujinSubjectLock  *bool
	SubjectEdit        *bool
	SubjectLock        *bool
	SubjectRefresh     *bool
	SubjectRelated     *bool
	SubjectMerge       *bool
	SubjectErase       *bool
	SubjectCoverLock   *bool
	SubjectCoverErase  *bool
	MonoEdit           *bool
	MonoLock           *bool
	MonoMerge          *bool
	MonoErase          *bool
	BanPost            *bool
	BanPostGroup       *bool
	BanPostGroupNSFW   *bool
	BanSubjectInterest *bool
	BanVisit           *bool
	Deleted            *bool
	EpEdit             *bool
	EpMove             *bool
	EpMerge            *bool
	EpLock             *bool
	EpErase            *bool
	Report             *bool
	ManageApp          *bool
	AppErase           *bool
}

func mergeBool(group bool, personal *bool) bool {
	if personal != nil {
		return *personal
	}
	return group
}

func (p Permission) Merge(personal PersonalPermission) Permission {
	return Permission{
		UserList:           mergeBool(p.UserList, personal.UserList),
		ManageUserGroup:    mergeBool(p.ManageUserGroup, personal.ManageUserGroup),
		ManageUserPhoto:    mergeBool(p.ManageUserPhoto, personal.ManageUserPhoto),
		ManageTopicState:   mergeBool(p.ManageTopicState, personal.ManageTopicState),
		ManageReport:       mergeBool(p.ManageReport, personal.ManageReport),
		UserBan:            mergeBool(p.UserBan, personal.UserBan),
		ManageUser:         mergeBool(p.ManageUser, personal.ManageUser),
		UserGroup:          mergeBool(p.UserGroup, personal.UserGroup),
		UserWikiApply:      mergeBool(p.UserWikiApply, personal.UserWikiApply),
		UserWikiApprove:    mergeBool(p.UserWikiApprove, personal.UserWikiApprove),
		DoujinSubjectErase: mergeBool(p.DoujinSubjectErase, personal.DoujinSubjectErase),
		DoujinSubjectLock:  mergeBool(p.DoujinSubjectLock, personal.DoujinSubjectLock),
		SubjectEdit:        mergeBool(p.SubjectEdit, personal.SubjectEdit),
		SubjectLock:        mergeBool(p.SubjectLock, personal.SubjectLock),
		SubjectRefresh:     mergeBool(p.SubjectRefresh, personal.SubjectRefresh),
		SubjectRelated:     mergeBool(p.SubjectRelated, personal.SubjectRelated),
		SubjectMerge:       mergeBool(p.SubjectMerge, personal.SubjectMerge),
		SubjectErase:       mergeBool(p.SubjectErase, personal.SubjectErase),
		SubjectCoverLock:   mergeBool(p.SubjectCoverLock, personal.SubjectCoverLock),
		SubjectCoverErase:  mergeBool(p.SubjectCoverErase, personal.SubjectCoverErase),
		MonoEdit:           mergeBool(p.MonoEdit, personal.MonoEdit),
		MonoLock:           mergeBool(p.MonoLock, personal.MonoLock),
		MonoMerge:          mergeBool(p.MonoMerge, personal.MonoMerge),
		MonoErase:          mergeBool(p.MonoErase, personal.MonoErase),
		BanPost:            mergeBool(p.BanPost, personal.BanPost),
		BanPostGroup:       mergeBool(p.BanPostGroup, personal.BanPostGroup),
		BanPostGroupNSFW:   mergeBool(p.BanPostGroupNSFW, personal.BanPostGroupNSFW),
		BanSubjectInterest: mergeBool(p.BanSubjectInterest, personal.BanSubjectInterest),
		BanVisit:           mergeBool(p.BanVisit, personal.BanVisit),
		Deleted:            mergeBool(p.Deleted, personal.Deleted),
		EpEdit:             mergeBool(p.EpEdit, personal.EpEdit),
		EpMove:             mergeBool(p.EpMove, personal.EpMove),
		EpMerge:            mergeBool(p.EpMerge, personal.EpMerge),
		EpLock:             mergeBool(p.EpLock, personal.EpLock),
		EpErase:            mergeBool(p.EpErase, personal.EpErase),
		Report:             mergeBool(p.Report, personal.Report),
		ManageApp:          mergeBool(p.ManageApp, personal.ManageApp),
		AppErase:           mergeBool(p.AppErase, personal.AppErase),
	}
}
