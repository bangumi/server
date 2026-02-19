// SPDX-License-Identifier: AGPL-3.0-only

package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/internal/ctxkey"
)

func TestNeedScope_NeedLogin(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NeedScope(ScopeWriteCollection)(func(c echo.Context) error {
		return nil
	})

	err := h(c)
	require.ErrorIs(t, err, errNeedLogin)
}

func TestNeedScope_Insufficient(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	a := &accessor.Accessor{}
	a.SetAuth(auth.Auth{Scope: auth.Scope{}})
	c.Set(ctxkey.User, a)

	h := NeedScope(ScopeWriteCollection)(func(c echo.Context) error {
		return nil
	})

	err := h(c)
	require.ErrorIs(t, err, errInsufficientScope)
}

func TestNeedScope_Match(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	a := &accessor.Accessor{}
	a.SetAuth(auth.Auth{Scope: auth.Scope{ScopeWriteCollection: true}})
	c.Set(ctxkey.User, a)

	reached := false
	h := NeedScope(ScopeWriteCollection)(func(c echo.Context) error {
		reached = true
		return nil
	})

	err := h(c)
	require.NoError(t, err)
	require.True(t, reached)
}

func TestNeedScope_Legacy(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	a := &accessor.Accessor{}
	a.SetAuth(auth.Auth{Legacy: true})
	c.Set(ctxkey.User, a)

	reached := false
	h := NeedScope(ScopeWriteCollection)(func(c echo.Context) error {
		reached = true
		return nil
	})

	err := h(c)
	require.NoError(t, err)
	require.True(t, reached)
}
