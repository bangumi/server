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

package handler

import (
	"errors"
	"net"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/pkg/strutil"
	"github.com/bangumi/server/internal/web/cookie"
	"github.com/bangumi/server/internal/web/handler/internal/ctxkey"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/session"
	"github.com/bangumi/server/internal/web/util"
)

var accessorPool = sync.Pool{New: func() interface{} { return &accessor{} }} //nolint:gochecknoglobals

func (h Handler) SessionAuthMiddleware(c *fiber.Ctx) error {
	var a = accessorPool.Get().(*accessor) //nolint:forcetypeassert
	defer accessorPool.Put(a)
	defer a.reset()
	a.fillBasicInfo(c)

	value := utils.UnsafeString(c.Context().Request.Header.Cookie(session.Key))
	if value != "" {
		s, err := h.getSession(c, value)
		if err != nil {
			if errors.Is(err, session.ErrExpired) || errors.Is(err, domain.ErrNotFound) {
				cookie.Clear(c, session.Key)
				return c.Next()
			}

			h.log.Error("failed to get session", zap.Error(err), a.LogRequestID())
			return res.InternalError(c, err, "failed to read session, please try clear your browser cookies and re-try")
		}

		auth, err := h.a.GetByIDWithCache(c.Context(), s.UserID)
		if err != nil {
			return h.InternalError(c, err, "failed to user with permission", a.LogRequestID(), log.UserID(s.UserID))
		}

		a.fillAuth(auth)
	}

	c.Context().SetUserValue(ctxkey.User, a)

	return c.Next()
}

func (h Handler) getSession(c *fiber.Ctx, value string) (session.Session, error) {
	s, err := h.session.Get(c.Context(), value)
	if err != nil {
		return session.Session{}, errgo.Wrap(err, "sessionManager.Get")
	}

	return s, nil
}

func (h Handler) AccessTokenAuthMiddleware(ctx *fiber.Ctx) error {
	var a = accessorPool.Get().(*accessor) //nolint:forcetypeassert
	defer accessorPool.Put(a)
	a.fillBasicInfo(ctx)

	authorization := ctx.Get(fiber.HeaderAuthorization)
	if authorization != "" {
		key, token := strutil.Partition(authorization, ' ')
		if key != "Bearer" {
			return res.Unauthorized("http Authorization header has wrong scope")
		}

		var auth domain.Auth
		var err error
		if auth, err = h.a.GetByTokenWithCache(ctx.Context(), token); err != nil {
			if errors.Is(err, domain.ErrNotFound) || errors.Is(err, session.ErrExpired) {
				cookie.Clear(ctx, session.Key)
				return res.Unauthorized("access token has been expired or doesn't exist")
			}

			return errgo.Wrap(err, "auth.GetByTokenWithCache")
		}

		a.fillAuth(auth)
	}

	ctx.Context().SetUserValue(ctxkey.User, a)
	return ctx.Next()
}

func (h Handler) getHTTPAccessor(c *fiber.Ctx) *accessor {
	u, ok := c.Context().UserValue(ctxkey.User).(*accessor) // get visitor
	if !ok {
		panic("can't convert type")
	}

	return u
}

type accessor struct {
	cfRay string
	ip    net.IP
	domain.Auth
	login bool
}

func (a *accessor) AllowNSFW() bool {
	return a.login && a.Auth.AllowNSFW()
}

func (a *accessor) fillBasicInfo(c *fiber.Ctx) {
	a.login = false
	a.cfRay = c.Get(req.HeaderCFRay)
	a.ip = util.RequestIP(c)
}

func (a *accessor) fillAuth(auth domain.Auth) {
	a.Auth = auth
	a.login = true
}

func (a accessor) LogRequestID() zap.Field {
	return zap.Object("request", a)
}

func (a accessor) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("id", a.cfRay)
	encoder.AddString("ip", a.ip.String())
	if a.login {
		encoder.AddUint32("user_id", uint32(a.Auth.ID))
	}
	return nil
}

// reset struct to zero value before put it back to pool.
func (a *accessor) reset() {
	a.cfRay = ""
	a.ip = nil
	a.login = false
	a.Auth = domain.Auth{}
}
