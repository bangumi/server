package index_test

import (
	"net/http"
	"testing"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/trim21/htest"
)

func TestCollectIndex(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(233)).Return(model.Index{ID: 233}, nil)
	mockIndex.EXPECT().GetIndexCollect(mock.Anything, mock.Anything, mock.Anything).Return(nil, gerr.ErrNotFound)
	mockIndex.EXPECT().AddIndexCollect(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		Post("/v0/indices/233/collect")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUncollectIndex(t *testing.T) {
	t.Parallel()
	mockIndex := mocks.NewIndexRepo(t)
	mockIndex.EXPECT().Get(mock.Anything, uint32(322)).Return(model.Index{ID: 322}, nil)
	mockIndex.EXPECT().GetIndexCollect(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	// mockIndex.EXPECT().DeleteIndexCollect(mock.Anything, uint32(322), uint32(6)).Return(nil)
	mockAuth := mocks.NewAuthRepo(t)
	mockAuth.EXPECT().GetByToken(mock.Anything, mock.Anything).Return(auth.UserInfo{ID: 6}, nil)
	mockAuth.EXPECT().GetPermission(mock.Anything, mock.Anything).Return(auth.Permission{}, nil)

	app := test.GetWebApp(t, test.Mock{IndexRepo: mockIndex, AuthRepo: mockAuth})

	resp := htest.New(t, app).
		Header(echo.HeaderAuthorization, "Bearer token").
		Delete("/v0/indices/322/collect")

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
