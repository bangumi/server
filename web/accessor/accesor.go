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

package accessor

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/web/req/cf"
)

type Accessor struct {
	RequestID string
	IP        string
	auth.Auth
	Login bool
}

func (a *Accessor) AllowNSFW() bool {
	return a.Login && a.Auth.AllowNSFW()
}

func (a *Accessor) fillBasicInfo(c echo.Context) {
	a.Login = false
	a.RequestID = c.Request().Header.Get(cf.HeaderRequestID)
	a.IP = c.RealIP()
}

func (a *Accessor) SetAuth(auth auth.Auth) {
	a.Auth = auth
	a.Login = true
}

func (a Accessor) Log() zap.Field {
	return zap.Object("request", a)
}

func (a Accessor) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("id", a.RequestID)
	encoder.AddString("IP", a.IP)
	if a.Login {
		encoder.AddUint32("user_id", a.ID)
	}
	return nil
}

// reset struct to zero value before put it back to pool.
func (a *Accessor) reset() {
	a.RequestID = ""
	a.IP = ""
	a.Login = false
	a.Auth = auth.Auth{}
}

func (a *Accessor) Free() {
	a.reset()
	accessorPool.Put(a)
}
