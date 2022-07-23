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

package req

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/web/res"
)

type PageQuery struct {
	Limit  int
	Offset int
}

func (q PageQuery) Check(count int64) error {
	if q.Offset > int(count) {
		return res.BadRequest("offset should be less than or equal to " + strconv.FormatInt(count, 10))
	}

	return nil
}

func GetPageQuery(c *fiber.Ctx, defaultLimit int, maxLimit int) (PageQuery, error) {
	q := PageQuery{Limit: defaultLimit}
	var err error

	raw := c.Query("limit")
	if raw != "" {
		q.Limit, err = strconv.Atoi(raw)
		if err != nil {
			return q, res.BadRequest("can't parse query args limit as int: " + strconv.Quote(raw))
		}

		if q.Limit > maxLimit {
			return q, res.BadRequest("limit should less equal than " + strconv.Itoa(maxLimit))
		}
		if q.Limit <= 0 {
			return q, res.BadRequest("limit should be greater than zero")
		}
	}

	raw = c.Query("offset")
	if raw != "" {
		q.Offset, err = strconv.Atoi(raw)
		if err != nil {
			return q, res.BadRequest("can't parse query args offset as int: " + strconv.Quote(raw))
		}

		if q.Offset < 0 {
			return q, res.BadRequest("offset should be greater than or equal to 0")
		}
	}

	return q, nil
}

const DefaultPageLimit = 30
const DefaultMaxPageLimit = 100

const EpisodeDefaultLimit = 100
const EpisodeMaxLimit = 1000
