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

package subject_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/res"
)

func TestSubject_GetRelatedSubjects(t *testing.T) {
	t.Parallel()

	var subjectID model.SubjectID = 7

	m := mocks.NewSubjectRepo(t)
	m.EXPECT().Get(mock.Anything, subjectID, mock.Anything).Return(model.Subject{ID: subjectID}, nil)
	m.EXPECT().GetByIDs(mock.Anything, mock.Anything, subject.Filter{NSFW: null.New(false)}).
		Return(map[model.SubjectID]model.Subject{1: {ID: 1}}, nil)
	m.EXPECT().GetSubjectRelated(mock.Anything, subjectID).Return([]domain.SubjectInternalRelation{
		{TypeID: 1, SourceID: subjectID, DestinationID: 1},
	}, nil)

	app := test.GetWebApp(t,
		test.Mock{
			SubjectRepo: m,
		},
	)

	var r []res.SubjectRelatedSubject
	resp := htest.New(t, app).
		Get("/v0/subjects/7/subjects").
		JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, r, 1)
	require.Equal(t, model.SubjectID(1), r[0].SubjectID)
}
