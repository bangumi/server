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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetSubjectTopic(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	topicID, err := parseTopicID(c.Params("topic_id"))
	if err != nil {
		return err
	}

	topic, err := h.getTopic(c, domain.TopicTypeSubject, topicID)
	if err != nil {
		return err
	}

	subjectID, err := getExpectSubjectID(c, topic)
	if err != nil {
		return err
	}
	r, ok, err := h.getSubjectWithCache(c.Context(), subjectID)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 || (r.NSFW && !u.AllowNSFW()) {
		return res.ErrNotFound
	}

	return h.getResTopicWithComments(c, domain.TopicTypeSubject, topic)
}

func (h Handler) ListSubjectTopics(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	id, err := parseSubjectID(c.Params("id"))
	if err != nil || id == 0 {
		return res.BadRequest(err.Error())
	}

	r, ok, err := h.getSubjectWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 || (r.NSFW && !u.AllowNSFW()) {
		return res.ErrNotFound
	}

	return h.listTopics(c, domain.TopicTypeSubject, uint32(id))
}
