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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetSubjectTopic(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	topicID, err := req.ParseTopicID(c.Params("topic_id"))
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

	_, err = h.app.Query.GetSubjectNoRedirect(c.Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to subject", log.SubjectID(subjectID))
	}

	return h.getResTopicWithComments(c, domain.TopicTypeSubject, topic)
}

func (h Handler) ListSubjectTopics(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil || id == 0 {
		return res.BadRequest(err.Error())
	}

	_, err = h.app.Query.GetSubjectNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to subject", log.SubjectID(id))
	}

	return h.listTopics(c, domain.TopicTypeSubject, uint32(id))
}
