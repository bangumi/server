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
	"errors"
	"strconv"
	"time"

	"github.com/elliotchance/phpserialize"
	"github.com/goccy/go-json"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/random"
	"github.com/bangumi/server/internal/strparse"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) domain.AuthRepo {
	return mysqlRepo{q: q, log: log.Named("auth.mysqlRepo")}
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetByEmail(ctx context.Context, email string) (domain.Auth, []byte, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.Email.Eq(email)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Auth{}, nil, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return domain.Auth{}, nil, errgo.Wrap(err, "gorm")
	}

	return domain.Auth{
		RegTime: time.Unix(u.Regdate, 0),
		ID:      u.ID,
		GroupID: u.Groupid,
	}, u.PasswordCrypt, nil
}

func (m mysqlRepo) GetByToken(ctx context.Context, token string) (domain.Auth, error) {
	access, err := m.q.AccessToken.WithContext(ctx).
		Where(m.q.AccessToken.AccessToken.Eq(token), m.q.AccessToken.ExpiredAt.Gte(time.Now())).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Auth{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))

		return domain.Auth{}, errgo.Wrap(err, "gorm")
	}

	id, err := strparse.UserID(access.UserID)
	if err != nil {
		m.log.Error("wrong UserID in OAuth Access table", zap.String("user_id", access.UserID))
		return domain.Auth{}, errgo.Wrap(err, "parsing user id")
	}

	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.log.Error("can't find user of access token",
				zap.String("token", token), zap.String("uid", access.UserID))

			return domain.Auth{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))

		return domain.Auth{}, errgo.Wrap(err, "gorm")
	}

	return domain.Auth{
		RegTime: time.Unix(u.Regdate, 0),
		ID:      u.ID,
		GroupID: u.Groupid,
	}, nil
}

func (m mysqlRepo) GetPermission(ctx context.Context, groupID uint8) (domain.Permission, error) {
	r, err := m.q.UserGroup.WithContext(ctx).Where(m.q.UserGroup.ID.Eq(groupID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.log.Error("can't find permission for group", log.UserGroup(groupID))
			return domain.Permission{}, nil
		}

		m.log.Error("unexpected error", zap.Error(err))
		return domain.Permission{}, errgo.Wrap(err, "dal")
	}

	var p phpPermission
	if len(r.Perm) > 0 {
		d, err := phpserialize.UnmarshalAssociativeArray(r.Perm)
		if err != nil {
			m.log.Error("failed to decode php serialized content", zap.Error(err), log.UserGroup(groupID))
			return domain.Permission{}, nil
		}
		if err = mapstructure.Decode(d, &p); err != nil {
			m.log.Error("failed to convert from map to struct when decoding permission",
				zap.Error(err), zap.Uint8("group_id", groupID))
			return domain.Permission{}, nil
		}
	}

	return domain.Permission{
		UserList:           parseBool(p.UserList),
		ManageUserGroup:    parseBool(p.ManageUserGroup),
		ManageUserPhoto:    parseBool(p.ManageUserPhoto),
		ManageTopicState:   parseBool(p.ManageTopicState),
		ManageReport:       parseBool(p.ManageReport),
		UserBan:            parseBool(p.UserBan),
		ManageUser:         parseBool(p.ManageUser),
		UserGroup:          parseBool(p.UserGroup),
		UserWikiApprove:    parseBool(p.UserWikiApprove),
		UserWikiApply:      parseBool(p.UserWikiApply),
		DoujinSubjectErase: parseBool(p.DoujinSubjectErase),
		DoujinSubjectLock:  parseBool(p.DoujinSubjectLock),
		SubjectEdit:        parseBool(p.SubjectEdit),
		SubjectLock:        parseBool(p.SubjectLock),
		SubjectRefresh:     parseBool(p.SubjectRefresh),
		SubjectRelated:     parseBool(p.SubjectRelated),
		SubjectMerge:       parseBool(p.SubjectMerge),
		SubjectErase:       parseBool(p.SubjectErase),
		SubjectCoverLock:   parseBool(p.SubjectCoverLock),
		SubjectCoverErase:  parseBool(p.SubjectCoverErase),
		MonoEdit:           parseBool(p.MonoEdit),
		MonoLock:           parseBool(p.MonoLock),
		MonoMerge:          parseBool(p.MonoMerge),
		MonoErase:          parseBool(p.MonoErase),
		EpEdit:             parseBool(p.EpEdit),
		EpMove:             parseBool(p.EpMove),
		EpMerge:            parseBool(p.EpMerge),
		EpLock:             parseBool(p.EpLock),
		EpErase:            parseBool(p.EpErase),
		Report:             parseBool(p.Report),
		ManageApp:          parseBool(p.ManageApp),
		AppErase:           parseBool(p.AppErase),
	}, nil
}

func parseBool(s string) bool {
	return s == "1"
}

type phpPermission struct {
	UserList           string `mapstruct:"user_list"`
	ManageUserGroup    string `mapstruct:"manage_user_group"`
	ManageUserPhoto    string `mapstruct:"manage_user_photo"`
	ManageTopicState   string `mapstruct:"manage_topic_state"`
	ManageReport       string `mapstruct:"manage_report"`
	UserBan            string `mapstruct:"user_ban"`
	ManageUser         string `mapstruct:"manage_user"`
	UserGroup          string `mapstruct:"user_group"`
	UserWikiApprove    string `mapstruct:"user_wiki_approve"`
	DoujinSubjectErase string `mapstruct:"doujin_subject_erase"`
	UserWikiApply      string `mapstruct:"user_wiki_apply"`
	DoujinSubjectLock  string `mapstruct:"doujin_subject_lock"`
	SubjectEdit        string `mapstruct:"subject_edit"`
	SubjectLock        string `mapstruct:"subject_lock"`
	SubjectRefresh     string `mapstruct:"subject_refresh"`
	SubjectRelated     string `mapstruct:"subject_related"`
	SubjectMerge       string `mapstruct:"subject_merge"`
	SubjectErase       string `mapstruct:"subject_erase"`
	SubjectCoverLock   string `mapstruct:"subject_cover_lock"`
	SubjectCoverErase  string `mapstruct:"subject_cover_erase"`
	MonoEdit           string `mapstruct:"mono_edit"`
	MonoLock           string `mapstruct:"mono_lock"`
	MonoMerge          string `mapstruct:"mono_merge"`
	MonoErase          string `mapstruct:"mono_erase"`
	EpEdit             string `mapstruct:"ep_edit"`
	EpMove             string `mapstruct:"ep_move"`
	EpMerge            string `mapstruct:"ep_merge"`
	EpLock             string `mapstruct:"ep_lock"`
	EpErase            string `mapstruct:"ep_erase"`
	Report             string `mapstruct:"report"`
	ManageApp          string `mapstruct:"manage_app"`
	AppErase           string `mapstruct:"app_erase"`
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

func (m mysqlRepo) ListAccessToken(ctx context.Context, userID model.UserID) ([]domain.AccessToken, error) {
	records, err := m.q.AccessToken.WithContext(ctx).
		Where(m.q.AccessToken.UserID.Eq(strconv.FormatUint(uint64(userID), 10)),
			m.q.AccessToken.ExpiredAt.Gte(time.Now())).Find()
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var tokens = make([]domain.AccessToken, len(records))
	for i, record := range records {
		tokens[i] = convertAccessToken(record)
	}

	return tokens, errgo.Wrap(err, "dal")
}

const defaultOauthAccessExpiration = time.Hour * 168

func convertAccessToken(t *dao.AccessToken) domain.AccessToken {
	var createdAt time.Time
	var name = "oauth token"

	switch t.Type {
	case TokenTypeAccessToken:
		if len(t.Info) > 0 {
			var info TokenInfo
			if err := json.UnmarshalNoEscape(t.Info, &info); err != nil {
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

	return domain.AccessToken{
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

func (m mysqlRepo) GetTokenByID(ctx context.Context, id uint32) (domain.AccessToken, error) {
	record, err := m.q.AccessToken.WithContext(ctx).Where(m.q.AccessToken.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AccessToken{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return domain.AccessToken{}, errgo.Wrap(err, "dal")
	}

	return convertAccessToken(record), errgo.Wrap(err, "dal")
}
