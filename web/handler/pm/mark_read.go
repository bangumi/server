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

package pm

import (
	"errors"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h PrivateMessage) MarkRead(c *fiber.Ctx) error {
	accessor := h.Common.GetHTTPAccessor(c)
	var r req.PrivateMessageMarkRead
	if err := sonic.Unmarshal(c.Body(), &r); err != nil {
		return res.JSONError(c, err)
	}

	if err := h.Common.V.Struct(r); err != nil {
		return h.ValidationError(c, err)
	}
	err := h.pmRepo.MarkRead(c.Context(), accessor.ID, model.PrivateMessageID(r.ID))
	if err != nil {
		if errors.Is(err, pm.ErrPmInvalidOperation) {
			return res.BadRequest(err.Error())
		}
		return res.InternalError(c, err, "failed to mark private message(s) read")
	}
	return c.SendStatus(http.StatusNoContent)
}
