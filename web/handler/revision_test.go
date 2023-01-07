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

package handler_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/web/res"
)

func TestHandler_ListPersonRevision_HappyPath(t *testing.T) {
	t.Parallel()
	const uid model.PersonID = 9

	m := mocks.NewRevisionRepo(t)
	m.EXPECT().ListPersonRelated(mock.Anything, uid, 30, 0).Return(
		[]model.PersonRevision{{RevisionCommon: model.RevisionCommon{ID: 348475}}}, nil)
	m.EXPECT().CountPersonRelated(mock.Anything, uid).Return(1, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	var r res.Paged
	resp := htest.New(t, app).Get("/v0/revisions/persons?person_id=9").JSON(&r)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	result, ok := r.Data.([]any)[0].(map[string]any)
	require.True(t, ok)

	id, ok := result["id"].(float64)
	require.True(t, ok)
	require.Equal(t, uint32(348475), uint32(id))
}

func TestHandler_ListPersonRevision_Bad_ID(t *testing.T) {
	t.Parallel()
	m := mocks.NewRevisionRepo(t)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	badIDs := []string{"-1", "a", "0"}

	for _, id := range badIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			htest.New(t, app).
				Query("person_id", id).
				Get("/v0/revisions/persons").
				ExpectCode(http.StatusBadRequest)
		})
	}
}

func TestHandler_GetPersonRevision_HappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewRevisionRepo(t)
	m.EXPECT().GetPersonRelated(mock.Anything, uint32(348475)).Return(
		model.PersonRevision{RevisionCommon: model.RevisionCommon{ID: 348475}}, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	var r res.PersonRevision
	resp := htest.New(t, app).Get("/v0/revisions/persons/348475").JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, uint32(348475), r.ID)
}

func TestHandler_ListSubjectRevision_HappyPath(t *testing.T) {
	t.Parallel()
	const subjectID model.SubjectID = 26

	m := mocks.NewRevisionRepo(t)
	m.EXPECT().ListSubjectRelated(mock.Anything, subjectID, 30, 0).Return(
		[]model.SubjectRevision{{RevisionCommon: model.RevisionCommon{ID: 665556}}}, nil)
	m.EXPECT().CountSubjectRelated(mock.Anything, subjectID).Return(1, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	var r res.Paged
	resp := htest.New(t, app).Get("/v0/revisions/subjects?subject_id=26").JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	result, ok := r.Data.([]any)[0].(map[string]any)
	require.Equal(t, true, ok)

	id, ok := result["id"].(float64)
	require.Equal(t, true, ok)
	require.Equal(t, uint32(665556), uint32(id))
}

func TestHandler_ListSubjectRevision_Bad_ID(t *testing.T) {
	t.Parallel()
	m := mocks.NewRevisionRepo(t)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	badIDs := []string{"-1", "a", "0"}

	for _, id := range badIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			htest.New(t, app).
				Query("subject_id", id).
				Get("/v0/revisions/subjects").
				ExpectCode(http.StatusBadRequest)
		})
	}
}

func TestHandler_GetSubjectRevision_HappyPath(t *testing.T) {
	t.Parallel()
	m := mocks.NewRevisionRepo(t)
	m.EXPECT().GetSubjectRelated(mock.Anything, uint32(665556)).Return(
		model.SubjectRevision{RevisionCommon: model.RevisionCommon{ID: 665556}}, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	var r res.SubjectRevision
	resp := htest.New(t, app).Get("/v0/revisions/subjects/665556").JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, uint32(665556), r.ID)
}

func TestHandler_ListCharacterRevision_HappyPath(t *testing.T) {
	var cid model.CharacterID = 20        // character id
	var mockRID model.RevisionID = 307134 // revision id

	t.Parallel()
	m := mocks.NewRevisionRepo(t)
	m.EXPECT().ListCharacterRelated(mock.Anything, cid, 30, 0).Return([]model.CharacterRevision{
		{RevisionCommon: model.RevisionCommon{ID: mockRID}},
	}, nil)
	m.EXPECT().CountCharacterRelated(mock.Anything, cid).Return(1, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	var r res.Paged
	resp := htest.New(t, app).
		Query("character_id", strconv.FormatUint(uint64(cid), 10)).
		Get("/v0/revisions/characters").
		JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	rdms, ok := r.Data.([]any) // rdm: r.Data.Maps
	require.True(t, ok)
	require.GreaterOrEqual(t, len(rdms), 1)

	rdm, ok := rdms[0].(map[string]any)
	require.True(t, ok)

	id, ok := rdm["id"].(float64)
	require.True(t, ok)
	require.Equal(t, mockRID, uint32(id))
}

func TestHandler_GetCharacterRevision_HappyPath(t *testing.T) {
	var mockRID uint32 = 1269825

	t.Parallel()
	m := mocks.NewRevisionRepo(t)
	m.EXPECT().GetCharacterRelated(mock.Anything, mockRID).Return(model.CharacterRevision{
		RevisionCommon: model.RevisionCommon{ID: mockRID},
	}, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	var r res.CharacterRevision
	resp := htest.New(t, app).
		Get(fmt.Sprintf("/v0/revisions/characters/%d", mockRID)).
		JSON(&r)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, mockRID, r.ID)
}
