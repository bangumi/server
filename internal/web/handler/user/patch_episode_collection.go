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

package user

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
)

type ReqEpisodeCollectionBatch struct {
	EpisodeID []model.EpisodeID       `json:"episode_id"`
	Type      model.EpisodeCollection `json:"type"`
}

func (r ReqEpisodeCollectionBatch) Validate() error {
	switch r.Type {
	case model.EpisodeCollectionAll,
		model.EpisodeCollectionWish,
		model.EpisodeCollectionDone,
		model.EpisodeCollectionDropped:
	default:
		return res.BadRequest(fmt.Sprintf("not valid episode collection type %d", r.Type))
	}

	return nil
}

func (h User) PatchEpisodeCollectionBatch(c *fiber.Ctx) error {
	var r ReqEpisodeCollectionBatch
	if err := json.Unmarshal(c.Body(), &r); err != nil {
		return res.JSONError(c, err)
	}

	if err := r.Validate(); err != nil {
		return err
	}

	return nil
}
