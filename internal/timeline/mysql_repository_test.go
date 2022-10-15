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
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/timeline"
)

func getRepo(t *testing.T) (domain.TimeLineRepo, *query.Query) {
	t.Helper()
	q := query.Use(test.GetGorm(t))
	repo, err := timeline.NewMysqlRepo(q, zap.NewNop())
	require.NoError(t, err)

	return repo, q
}

func Test_mysqlRepo_GetByID(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	var tlID model.TimeLineID = 28979826

	repo, q := getRepo(t)
	ctx := context.Background()

	tlModel, err := repo.GetByID(ctx, tlID)
	require.NoError(t, err)
	tlDAO, err := q.TimeLine.WithContext(ctx).Where(q.TimeLine.ID.Eq(tlID)).First()
	require.NoError(t, err)

	require.Equal(t, tlModel.ID, tlDAO.ID)
	require.Equal(t, tlModel.UID, tlDAO.UID)
	require.Equal(t, tlModel.Cat, tlDAO.Cat)
	require.Equal(t, tlModel.Type, tlDAO.Type)
}

func Test_mysqlRepo_ListByUID(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	var (
		uid   model.UserID = 287622
		limit              = -1
		since model.TimeLineID
	)

	repo, q := getRepo(t)
	ctx := context.Background()

	tls, err := repo.ListByUID(ctx, uid, limit, since)
	require.NoError(t, err)
	daos, err := q.TimeLine.WithContext(ctx).
		Where(q.TimeLine.UID.Eq(uid), q.TimeLine.ID.Gt(since)).
		Order(q.TimeLine.Dateline).
		Limit(limit).
		Find()
	require.NoError(t, err)
	require.Equal(t, len(daos), len(tls))
}

func Test_mysqlRepo_Create(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	var cases = []model.TimeLineID{
		28684877, // cat=1, type=2
		28682314, // cat=1, type=3
		28683701, // cat=2
		28685055, // cat=3
		28976108, // cat=4
		28684294, // cat=5, type=2
		28683198, // cat=5, type=1
		28684975, // cat=6
		28684740, // cat=7
		28685042, // cat=8
		28523860, // cat=9
	}

	repo, q := getRepo(t)
	ctx := context.Background()

	for _, tlID := range cases {
		msg := fmt.Sprintf("start testing case: %d", tlID)
		newTLID := tlID + 28523860

		test.RunAndCleanup(t, func() {
			// delete if already exists
			_, err := q.WithContext(ctx).TimeLine.Where(q.TimeLine.ID.Eq(newTLID)).Delete()
			require.NoError(t, err, msg)
		})

		// get the timeline
		tlModel, err := repo.GetByID(ctx, tlID)
		require.NoError(t, err, msg)

		// alter id and uid
		tlModel.ID = newTLID
		tlModel.UID += 654321

		// create with new id
		err = repo.Create(ctx, tlModel)
		require.NoError(t, err, msg)

		// get new timeline
		newTLModel, err := repo.GetByID(ctx, newTLID)
		require.NoError(t, err, msg)

		// check if the new timeline eq old timeline
		require.Equal(t, tlModel, newTLModel, msg)
	}
}

func Test_mysqlRepo_NewTimeLineMemo(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	type caseT struct {
		id        model.TimeLineID
		extractor func(content *model.TimeLineMemoContent) *model.TimeLineMemo
	}

	var cases = []caseT{
		{28684877, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineRelationMemo)
		}}, // cat=1, type=2
		{28682314, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineGroupMemo)
		}}, // cat=1, type=3
		{28683701, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineWikiMemo)
		}}, // cat=2
		{28685055, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineSubjectMemo)
		}}, // cat=3
		{28976108, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineProgressMemo)
		}}, // cat=4
		{28684294, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineSayMemo)
		}}, // cat=5, type=2
		{28683198, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineSayMemo)
		}}, // cat=5, type=1
		{28684975, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineBlogMemo)
		}}, // cat=6
		{28684740, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineIndexMemo)
		}}, // cat=7
		{28685042, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineMonoMemo)
		}}, // cat=8
		{28523860, func(content *model.TimeLineMemoContent) *model.TimeLineMemo {
			return model.NewTimeLineMemo(content.TimeLineDoujinMemo)
		}}, // cat=9
	}

	repo, _ := getRepo(t)
	ctx := context.Background()

	for _, cas := range cases {
		cas := cas
		t.Run(fmt.Sprintf("start testing case: %d", cas.id), func(t *testing.T) {
			t.Parallel()

			// get the model.TL
			expected, err := repo.GetByID(ctx, cas.id)
			require.NoError(t, err)

			// check if the NewTimeLineMemo works
			require.Equal(t, expected.TimeLineMemo, cas.extractor(expected.Content))
		})
	}
}
