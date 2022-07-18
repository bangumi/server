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

package accesor

import (
	"net"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/util"
)

type Accessor struct {
	RequestID string
	IP        net.IP
	domain.Auth
	Login bool
}

func (a *Accessor) AllowNSFW() bool {
	return a.Login && a.Auth.AllowNSFW()
}

func (a *Accessor) FillBasicInfo(c *fiber.Ctx) {
	a.Login = false
	a.RequestID = c.Get(req.HeaderCFRay)
	a.IP = util.RequestIP(c)
}

func (a *Accessor) SetAuth(auth domain.Auth) {
	a.Auth = auth
	a.Login = true
}

func (a Accessor) LogRequestID() zap.Field {
	return zap.Object("request", a)
}

func (a Accessor) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("id", a.RequestID)
	encoder.AddString("IP", a.IP.String())
	if a.Login {
		encoder.AddUint32("user_id", uint32(a.Auth.ID))
	}
	return nil
}

// reset struct to zero value before put it back to pool.
func (a *Accessor) reset() {
	a.RequestID = ""
	a.IP = nil
	a.Login = false
	a.Auth = domain.Auth{}
}
