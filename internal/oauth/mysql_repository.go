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

package oauth

import (
	"context"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/pkg/logger"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (Manager, error) {
	return mysqlRepo{q: q, log: log.Named("episode.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetClientByID(ctx context.Context, clientIDs ...string) (map[string]Client, error) {
	clients, err := m.q.OAuthClient.WithContext(ctx).Joins(m.q.OAuthClient.App).
		Where(m.q.OAuthClient.ClientID.In(clientIDs...)).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var data = make(map[string]Client, len(clients))
	for _, record := range clients {
		data[record.ClientID] = convertFromDao(record)
	}

	return data, nil
}

func convertFromDao(record *dao.OAuthClient) Client {
	var userID model.UserID
	if record.UserID != "" {
		uid, err := gstr.ParseUint32(record.UserID)
		if err != nil || uid == 0 {
			logger.Fatal("unexpected error when parsing userID", zap.Error(err), zap.String("raw", record.UserID))
		}
		userID = model.UserID(uid)
	}

	return Client{
		ID:          record.ClientID,
		Secret:      record.ClientSecret,
		RedirectURI: record.RedirectURI,
		GrantTypes:  record.GrantTypes,
		Scope:       record.Scope,
		UserID:      userID,
		AppID:       record.AppID,
		AppName:     record.App.Name,
	}
}
