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

package timeline_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/elliotchance/phpserialize"
	"github.com/goccy/go-json"
	"github.com/gookit/goutil/dump"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/internal/timeline"
)

func getRepo(t *testing.T) (domain.TimeLineRepo, *query.Query) {
	t.Helper()
	q := query.Use(test.GetGorm(t))
	repo, err := timeline.NewMysqlRepo(q, zap.NewNop())
	require.NoError(t, err)

	return repo, q
}

// func TestMysqlRepo_GetByID(t *testing.T) {
// 	test.RequireEnv(t, test.EnvMysql)
// 	t.Parallel()
//
// 	repo := getRepo(t)
// 	tl, err := repo.GetByID(context.Background(), 27379258)
// 	require.NoError(t, err)
//
// 	t.Log(dump.Format(tl))
// }

func TestDump(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)

	t.Parallel()

	require.NoError(t, os.MkdirAll("./testdata/", 0644))

	repo, q := getRepo(t)

	var ids []model.TimeLineID
	err := q.TimeLine.WithContext(context.Background()).Pluck(q.TimeLine.ID, &ids)
	require.NoError(t, err)

	for _, id := range ids {
		func() {
			tl, err := repo.GetByID(context.Background(), id)
			require.NoError(t, err)
			file, err := os.Create(fmt.Sprintf("./testdata/dump-%d.json", id))
			require.NoError(t, err)
			defer file.Close()
			enc := json.NewEncoder(file)
			enc.SetIndent("", "  ")
			require.NoError(t, enc.Encode(tl))
		}()
	}
}

type RawMessage []byte

func (r RawMessage) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}

	data, err := phpserialize.UnmarshalAssociativeArray(r)
	if err != nil {
		return nil, fmt.Errorf("phpserialize.UnmarshalAssociativeArray: %w: %s", err, string(r))
	}

	var m = make(map[string]interface{}, len(data))

	for k, value := range data {
		key, ok := k.(string)
		if !ok {
			panic(fmt.Sprintf("%v is not string", k))
		}

		m[key] = value
	}

	b, err := json.Marshal(m)
	if err != nil {
		dump.P(data)
	}

	return b, err
}

var _ json.Marshaler = (RawMessage)(nil)

type TimeLine struct {
	Image    RawMessage       `json:"image"`
	Related  string           `json:"related"`
	Memo     RawMessage       `json:"memo"`
	UID      uint32           `json:"uid"`
	Replies  uint32           `json:"replies"`
	ID       model.TimeLineID `json:"id"`
	Dateline uint32           `json:"dateline"`
	Cat      uint16           `json:"cat"`
	Type     uint16           `json:"type"`
	Batch    uint8            `json:"batch"`
	Source   uint8            `json:"source"`
}
