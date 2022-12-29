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

	CreateAccessToken(
		ctx context.Context, userID model.UserID, name string, expiration time.Duration,
	) (token string, err error)

	ListAccessToken(ctx context.Context, userID model.UserID) ([]AccessToken, error)
	DeleteAccessToken(ctx context.Context, tokenID uint32) (bool, error)

	// GetByEmail return (Auth, HashedPassword, error)
	GetByEmail(ctx context.Context, email string) (UserInfo, []byte, error)
	GetTokenByID(ctx context.Context, id uint32) (AccessToken, error)
}

type Service interface {
	GetByToken(ctx context.Context, token string) (Auth, error)
	GetByID(ctx context.Context, userID model.UserID) (Auth, error)

	ComparePassword(hashed []byte, password string) (bool, error)

	Login(ctx context.Context, email, password string) (Auth, bool, error)

	GetTokenByID(ctx context.Context, tokenID uint32) (AccessToken, error)
	CreateAccessToken(
		ctx context.Context, userID model.UserID, name string, expiration time.Duration,
	) (token string, err error)
	ListAccessToken(ctx context.Context, userID model.UserID) ([]AccessToken, error)
	DeleteAccessToken(ctx context.Context, tokenID uint32) (bool, error)

	// GetPermission(ctx context.Context, id model.UserGroupID) (Permission, error)
}

type UserInfo struct {
	RegTime time.Time
	ID      model.UserID
	GroupID user.GroupID
}

// Auth is the basic authorization represent a user.
type Auth struct {
	RegTime    time.Time
	ID         model.UserID // user id
	GroupID    user.GroupID
	Permission Permission `json:"-"` // disable cache for this field.
}

const nsfwThreshold = gtime.OneDay * 60

// AllowNSFW return if current user is allowed to see NSFW resource.
func (u Auth) AllowNSFW() bool {
	return u.RegisteredLongerThan(nsfwThreshold)
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
