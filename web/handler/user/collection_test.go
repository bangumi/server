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

package user_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web/res"
)

func TestUser_ListCollection(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const subjectID model.SubjectID = 9

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(user.User{ID: userID, UserName: username}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().ListSubjectCollection(mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything, 10, 0).
		Return([]collection.UserSubjectCollection{{SubjectID: subjectID, Type: 1}}, nil)
	c.EXPECT().CountSubjectCollections(mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything).
		Return(1, nil)

	s := mocks.NewSubjectRepo(t)
	s.EXPECT().GetByIDs(mock.Anything, mock.Anything, mock.Anything).Return(map[model.SubjectID]model.Subject{
		subjectID: {Name: "v"},
	}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, CollectionRepo: c, SubjectRepo: s})

	var r struct {
		Data   json.RawMessage `json:"data"`
		Total  int64           `json:"total"`
		Limit  int             `json:"limit"`
		Offset int             `json:"offset"`
	}

	resp := htest.New(t, app).
		Query("limit", "10").
		Get(fmt.Sprintf("/v0/users/%s/collections", username)).
		JSON(&r).
		ExpectCode(http.StatusOK)

	var data []res.SubjectCollection
	require.NoError(t, json.Unmarshal(r.Data, &data))

	require.Len(t, data, 1)

	require.Equal(t, subjectID, data[0].SubjectID, resp.BodyString())
	require.Equal(t, "v", data[0].Subject.Name, resp.BodyString())
}

func TestUser_GetSubjectCollection(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const subjectID model.SubjectID = 9

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(user.User{ID: userID, UserName: username}, nil)
	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetSubjectCollection(mock.Anything, userID, mock.Anything).
		Return(collection.UserSubjectCollection{SubjectID: subjectID, Type: 1}, nil)

	s := mocks.NewSubjectRepo(t)
	s.EXPECT().Get(mock.Anything, subjectID, mock.Anything).Return(model.Subject{
		Name: "v",
	}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, CollectionRepo: c, SubjectRepo: s})

	var r res.SubjectCollection
	resp := htest.New(t, app).
		Get(fmt.Sprintf("/v0/users/%s/collections/%d", username, subjectID)).
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.Equal(t, subjectID, r.SubjectID, resp.BodyString())
	require.Equal(t, "v", r.Subject.Name, resp.BodyString())
}

func TestUser_ListSubjectCollection_other_user(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const subjectID model.SubjectID = 9

	a := mocks.NewAuthService(t)
	a.EXPECT().GetByToken(mock.Anything, "v").Return(auth.Auth{ID: userID + 1}, nil)

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(user.User{ID: userID, UserName: username}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetSubjectCollection(mock.Anything, userID, mock.Anything).
		Return(collection.UserSubjectCollection{SubjectID: subjectID, Private: true}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, AuthService: a, CollectionRepo: c})

	htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer v").
		Get(fmt.Sprintf("/v0/users/%s/collections/%d", username, subjectID)).
		ExpectCode(http.StatusNotFound)
}

func TestUser_GetPersonCollection(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const personID model.PersonID = 9

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(user.User{ID: userID, UserName: username}, nil)
	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetPersonCollection(mock.Anything, userID, mock.Anything, mock.Anything).
		Return(collection.UserPersonCollection{UserID: userID, Category: "prsn", TargetID: personID}, nil)

	p := mocks.NewPersonRepo(t)
	p.EXPECT().Get(mock.Anything, personID).Return(model.Person{
		ID:   personID,
		Name: "v",
	}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, CollectionRepo: c, PersonRepo: p})

	var r res.PersonCollection
	resp := htest.New(t, app).
		Get(fmt.Sprintf("/v0/users/%s/collections/-/persons/%d", username, personID)).
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.Equal(t, personID, r.ID, resp.BodyString())
	require.Equal(t, "v", r.Name, resp.BodyString())
}

func TestUser_ListPersonCollection(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const personID model.PersonID = 9

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(user.User{ID: userID, UserName: username}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().ListPersonCollection(mock.Anything, userID, mock.Anything, 10, 0).
		Return([]collection.UserPersonCollection{{UserID: userID, Category: "prsn", TargetID: personID}}, nil)
	c.EXPECT().CountPersonCollections(mock.Anything, userID, mock.Anything).
		Return(1, nil)

	p := mocks.NewPersonRepo(t)
	p.EXPECT().GetByIDs(mock.Anything, mock.Anything).Return(map[model.PersonID]model.Person{
		personID: {ID: personID, Name: "v"},
	}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, CollectionRepo: c, PersonRepo: p})

	var r struct {
		Data   json.RawMessage `json:"data"`
		Total  int64           `json:"total"`
		Limit  int             `json:"limit"`
		Offset int             `json:"offset"`
	}

	resp := htest.New(t, app).
		Query("limit", "10").
		Get(fmt.Sprintf("/v0/users/%s/collections/-/persons", username)).
		JSON(&r).
		ExpectCode(http.StatusOK)

	var data []res.PersonCollection
	require.NoError(t, json.Unmarshal(r.Data, &data))

	require.Len(t, data, 1)

	require.Equal(t, personID, data[0].ID, resp.BodyString())
	require.Equal(t, "v", data[0].Name, resp.BodyString())
}

func TestUser_GetCharacterCollection(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const characterID model.CharacterID = 9

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(user.User{ID: userID, UserName: username}, nil)
	c := mocks.NewCollectionRepo(t)
	c.EXPECT().GetPersonCollection(mock.Anything, userID, mock.Anything, mock.Anything).
		Return(collection.UserPersonCollection{UserID: userID, Category: "crt", TargetID: characterID}, nil)

	p := mocks.NewCharacterRepo(t)
	p.EXPECT().Get(mock.Anything, characterID).Return(model.Character{
		ID:   characterID,
		Name: "v",
	}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, CollectionRepo: c, CharacterRepo: p})

	var r res.PersonCollection
	resp := htest.New(t, app).
		Get(fmt.Sprintf("/v0/users/%s/collections/-/characters/%d", username, characterID)).
		JSON(&r).
		ExpectCode(http.StatusOK)

	require.Equal(t, characterID, r.ID, resp.BodyString())
	require.Equal(t, "v", r.Name, resp.BodyString())
}

func TestUser_ListCharacterCollection(t *testing.T) {
	t.Parallel()
	const username = "ni"
	const userID model.UserID = 7
	const characterID model.CharacterID = 9

	m := mocks.NewUserRepo(t)
	m.EXPECT().GetByName(mock.Anything, username).Return(user.User{ID: userID, UserName: username}, nil)

	c := mocks.NewCollectionRepo(t)
	c.EXPECT().ListPersonCollection(mock.Anything, userID, mock.Anything, 10, 0).
		Return([]collection.UserPersonCollection{{UserID: userID, Category: "crt", TargetID: characterID}}, nil)
	c.EXPECT().CountPersonCollections(mock.Anything, userID, mock.Anything).
		Return(1, nil)

	p := mocks.NewCharacterRepo(t)
	p.EXPECT().GetByIDs(mock.Anything, mock.Anything).Return(map[model.CharacterID]model.Character{
		characterID: {ID: characterID, Name: "v"},
	}, nil)

	app := test.GetWebApp(t, test.Mock{UserRepo: m, CollectionRepo: c, CharacterRepo: p})

	var r struct {
		Data   json.RawMessage `json:"data"`
		Total  int64           `json:"total"`
		Limit  int             `json:"limit"`
		Offset int             `json:"offset"`
	}

	resp := htest.New(t, app).
		Query("limit", "10").
		Get(fmt.Sprintf("/v0/users/%s/collections/-/characters", username)).
		JSON(&r).
		ExpectCode(http.StatusOK)

	var data []res.PersonCollection
	require.NoError(t, json.Unmarshal(r.Data, &data))

	require.Len(t, data, 1)

	require.Equal(t, characterID, data[0].ID, resp.BodyString())
	require.Equal(t, "v", data[0].Name, resp.BodyString())
}
