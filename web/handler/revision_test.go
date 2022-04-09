// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
//
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
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/res"
)

func TestHandler_ListPersonRevision_HappyPath(t *testing.T) {
	t.Parallel()
	m := &mocks.RevisionRepo{}
	m.EXPECT().ListPersonRelated(mock.Anything, uint32(9), 30, 0).Return([]model.Revision{{ID: 348475}}, nil)
	m.EXPECT().CountPersonRelated(mock.Anything, uint32(9)).Return(1, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/revisions/persons?person_id=9", http.NoBody)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r res.Paged

	err = json.NewDecoder(resp.Body).Decode(&r)

	require.NoError(t, err)

	if result, ok := r.Data.([]interface{})[0].(map[string]interface{}); ok {
		if id, ok := result["id"].(float64); ok {
			require.Equal(t, uint32(348475), uint32(id))
		}
	}
}

func TestHandler_ListPersonRevision_Bad_ID(t *testing.T) {
	t.Parallel()
	m := &mocks.RevisionRepo{}

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	badIDs := []string{"-1", "a", "0"}

	for _, id := range badIDs {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v0/revisions/persons?person_id=%s", id), http.NoBody)

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandler_GetPersonRevision_HappyPath(t *testing.T) {
	t.Parallel()
	m := &mocks.RevisionRepo{}
	m.EXPECT().GetPersonRelated(mock.Anything, uint32(348475)).Return(model.Revision{ID: 348475}, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/revisions/persons/348475", http.NoBody)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var r res.PersonRevision
	err = json.NewDecoder(resp.Body).Decode(&r)
	require.NoError(t, err)
	require.Equal(t, uint32(348475), r.ID)
}

func TestHandler_ListSubjectRevision_HappyPath(t *testing.T) {
	t.Parallel()
	m := &mocks.RevisionRepo{}
	m.EXPECT().ListSubjectRelated(mock.Anything, uint32(26), 30, 0).Return([]model.Revision{{ID: 665556}}, nil)
	m.EXPECT().CountSubjectRelated(mock.Anything, uint32(26)).Return(1, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/revisions/subjects?subject_id=26", http.NoBody)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r res.Paged

	err = json.NewDecoder(resp.Body).Decode(&r)

	require.NoError(t, err)

	if result, ok := r.Data.([]interface{})[0].(map[string]interface{}); ok {
		if id, ok := result["id"].(float64); ok {
			require.Equal(t, uint32(665556), uint32(id))
		}
	}
}

func TestHandler_ListSubjectRevision_Bad_ID(t *testing.T) {
	t.Parallel()
	m := &mocks.RevisionRepo{}

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	badIDs := []string{"-1", "a", "0"}

	for _, id := range badIDs {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v0/revisions/subjects?subject_id=%s", id), http.NoBody)

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandler_GetSubjectRevision_HappyPath(t *testing.T) {
	t.Parallel()
	m := &mocks.RevisionRepo{}
	m.EXPECT().GetSubjectRelated(mock.Anything, uint32(665556)).Return(model.Revision{ID: 665556}, nil)

	app := test.GetWebApp(t, test.Mock{RevisionRepo: m})

	req := httptest.NewRequest(http.MethodGet, "/v0/revisions/subjects/665556", http.NoBody)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var r res.SubjectRevision
	err = json.NewDecoder(resp.Body).Decode(&r)
	require.NoError(t, err)
	require.Equal(t, uint32(665556), r.ID)
}
