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
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func getExpectSubjectID(c *fiber.Ctx, topic model.Topic) (model.SubjectID, error) {
	subjectID, err := req.ParseSubjectID(c.Params("id"))
	if err != nil || subjectID == 0 {
		subjectID = model.SubjectID(topic.ObjectID)
	} else if subjectID != model.SubjectID(topic.ObjectID) {
		return model.SubjectID(0), res.ErrNotFound
	}
	return subjectID, nil
}
