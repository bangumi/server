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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/random"
	"github.com/bangumi/server/internal/user"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger, db *sqlx.DB) Repo {
	return mysqlRepo{
		q:   q,
		log: log.Named("auth.mysqlRepo"),
		db:  db,
	}
}

type mysqlRepo struct {
	q   *query.Query
	db  *sqlx.DB
	log *zap.Logger
}

func (m mysqlRepo) GetByEmail(ctx context.Context, email string) (UserInfo, []byte, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.Email.Eq(email)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return UserInfo{}, nil, gerr.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return UserInfo{}, nil, errgo.Wrap(err, "gorm")
	}

	return UserInfo{
		RegTime: time.Unix(u.Regdate, 0),
		ID:      u.ID,
		GroupID: u.Groupid,
	}, u.PasswordCrypt, nil
}

func (m mysqlRepo) GetByToken(ctx context.Context, token string) (UserInfo, error) {
	var access struct {
		UserID string `db:"user_id"`
	}
	err := m.db.GetContext(ctx, &access,
		`select user_id from chii_oauth_access_tokens where access_token = ? collate utf8_bin and expires > ? limit 1`, token, time.Now())
	if err != nil {
		fmt.Println("1", err)
		if errors.Is(err, sql.ErrNoRows) {
			return UserInfo{}, gerr.ErrNotFound
		}

		return UserInfo{}, errgo.Wrap(err, "gorm")
	}

	id, err := gstr.ParseUint32(access.UserID)
	if err != nil || id == 0 {
		fmt.Println("2", err)
		m.log.Error("wrong UserID in OAuth Access table", zap.String("user_id", access.UserID))
		return UserInfo{}, errgo.Wrap(err, "parsing user id")
	}

	var u struct {
		Regdate int64
		GroupID user.GroupID
	}

	err = m.db.QueryRowContext(ctx, `select regdate, groupid from chii_members where uid = ? limit 1`, id).
		Scan(&u.Regdate, &u.GroupID)
	if err != nil {
		fmt.Println("3", err)
		if errors.Is(err, sql.ErrNoRows) {
			return UserInfo{}, gerr.ErrNotFound
		}

		return UserInfo{}, errgo.Wrap(err, "gorm")
	}

	return UserInfo{
		RegTime: time.Unix(u.Regdate, 0),
		ID:      id,
		GroupID: u.GroupID,
	}, nil
}

func (m mysqlRepo) GetPermission(ctx context.Context, groupID uint8) (Permission, error) {
	r, err := m.q.UserGroup.WithContext(ctx).Where(m.q.UserGroup.ID.Eq(groupID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.log.Error("can't find permission for group", zap.Uint8("user_group_id", groupID))
			return Permission{}, nil
		}

		m.log.Error("unexpected error", zap.Error(err))
		return Permission{}, errgo.Wrap(err, "dal")
	}

	p, err := parsePhpSerializedPermission(r.Perm)
	if err != nil {
		m.log.Error("failed to decode php serialized content", zap.Error(err), zap.Uint8("user_group_id", groupID))
		return Permission{}, nil
	}

	return p, nil
}

const defaultAccessTokenLength = 40

func (m mysqlRepo) CreateAccessToken(
	ctx context.Context, id model.UserID, name string, expiration time.Duration,
) (string, error) {
	token := random.Base62String(defaultAccessTokenLength)
	var now = time.Now()

	var info = TokenInfo{
		Name:      name,
		CreatedAt: now,
	}

	var expiredAt = now.Add(expiration)
	if expiration < 0 {
		expiredAt = time.Time{}
	}

	infoByte, err := json.Marshal(info)
	if err != nil {
		// marshal simple struct should never fail
		m.log.Fatal("marshal simple struct should never fail",
			zap.Error(err), zap.String("name", name), zap.Time("now", now))
		panic("unexpected json encode error")
	}

	err = m.q.AccessToken.WithContext(ctx).Create(&dao.AccessToken{
		Type:        TokenTypeAccessToken,
		AccessToken: token,
		ClientID:    "access token",
		UserID:      strconv.FormatUint(uint64(id), 10),
		ExpiredAt:   expiredAt,
		Scope:       nil,
		Info:        infoByte,
	})
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))
		return "", errgo.Wrap(err, "dal")
	}

	return token, nil
}

type TokenInfo struct {
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
}

func (m mysqlRepo) ListAccessToken(ctx context.Context, userID model.UserID) ([]AccessToken, error) {
	records, err := m.q.AccessToken.WithContext(ctx).
		Where(m.q.AccessToken.UserID.Eq(strconv.FormatUint(uint64(userID), 10)),
			m.q.AccessToken.ExpiredAt.Gte(time.Now())).Find()
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var tokens = make([]AccessToken, len(records))
	for i, record := range records {
		tokens[i] = convertAccessToken(record)
	}

	return tokens, errgo.Wrap(err, "dal")
}

const defaultOauthAccessExpiration = time.Hour * 168

func convertAccessToken(t *dao.AccessToken) AccessToken {
	var createdAt time.Time
	var name = "oauth token"

	switch t.Type {
	case TokenTypeAccessToken:
		if len(t.Info) > 0 {
			var info TokenInfo
			if err := json.Unmarshal(t.Info, &info); err != nil {
				logger.Fatal("unexpected error when trying to unmarshal json data",
					zap.Error(err), zap.ByteString("raw", t.Info))
			}
			name = info.Name
			createdAt = info.CreatedAt
		} else {
			name = "personal access token"
		}
	case TokenTypeOauthToken:
		createdAt = t.ExpiredAt.Add(-defaultOauthAccessExpiration)
	}

	v, err := strconv.ParseUint(t.UserID, 10, 32)
	if err != nil {
		logger.Fatal("parsing UserID", zap.String("raw", t.UserID), zap.Error(err))
	}

	return AccessToken{
		ExpiredAt: t.ExpiredAt,
		CreatedAt: createdAt,
		Name:      name,
		UserID:    model.UserID(v),
		ClientID:  t.ClientID,
		ID:        t.ID,
	}
}

func (m mysqlRepo) DeleteAccessToken(ctx context.Context, id uint32) (bool, error) {
	info, err := m.q.AccessToken.WithContext(ctx).Where(m.q.AccessToken.ID.Eq(id)).Delete()

	return info.RowsAffected > 0, errgo.Wrap(err, "dal.Delete")
}

func (m mysqlRepo) GetTokenByID(ctx context.Context, id uint32) (AccessToken, error) {
	record, err := m.q.AccessToken.WithContext(ctx).Where(m.q.AccessToken.ID.Eq(id)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AccessToken{}, gerr.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return AccessToken{}, errgo.Wrap(err, "dal")
	}

	return convertAccessToken(record), errgo.Wrap(err, "dal")
}
