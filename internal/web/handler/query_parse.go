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
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/internal/web/res/code"
)

// these errors result in to 400 http response.
var errMissingCharacterID = fiber.NewError(code.BadRequest, "character ID is required")
var errMissingSubjectID = fiber.NewError(code.BadRequest, "subject ID is required")
var errMissingPersonID = fiber.NewError(code.BadRequest, "person ID is required")
var errMissingEpisodeID = fiber.NewError(code.BadRequest, "episode ID is required")
var errMissingIndexID = fiber.NewError(code.BadRequest, "index ID is required")
var errMissingTopicID = fiber.NewError(code.BadRequest, "topic ID is required")

func parseSubjectType(s string) (uint8, error) {
	if s == "" {
		return 0, nil
	}

	t, err := strparse.Uint8(s)
	if err != nil {
		return 0, fiber.NewError(http.StatusBadRequest, "bad subject type: "+strconv.Quote(s))
	}

	switch t {
	case model.SubjectAnime, model.SubjectBook,
		model.SubjectMusic, model.SubjectReal, model.SubjectGame:
		return t, nil
	}

	return 0, fiber.NewError(http.StatusBadRequest, strconv.Quote(s)+" is not a valid subject type")
}

func parseSubjectID(s string) (model.SubjectIDType, error) {
	if s == "" {
		return 0, errMissingSubjectID
	}

	v, err := strparse.SubjectID(s)

	if err != nil {
		return 0, fiber.NewError(code.BadRequest, strconv.Quote(s)+" is not valid subject ID")
	}

	return v, nil
}

func parseCharacterID(s string) (model.CharacterIDType, error) {
	if s == "" {
		return 0, errMissingCharacterID
	}

	v, err := strparse.CharacterID(s)

	if err != nil {
		return 0, fiber.NewError(code.BadRequest, strconv.Quote(s)+" is not valid character ID")
	}

	return v, nil
}

func parsePersonID(s string) (model.PersonIDType, error) {
	if s == "" {
		return 0, errMissingPersonID
	}

	v, err := strparse.PersonID(s)

	if err != nil {
		return 0, fiber.NewError(code.BadRequest, strconv.Quote(s)+" is not valid person ID")
	}

	return v, nil
}

func parseEpisodeID(s string) (model.EpisodeIDType, error) {
	if s == "" {
		return 0, errMissingEpisodeID
	}

	v, err := strparse.EpisodeID(s)

	if err != nil {
		return 0, fiber.NewError(code.BadRequest, strconv.Quote(s)+" is not a valid episode ID")
	}

	return v, nil
}

func parseIndexID(s string) (model.IndexIDType, error) {
	if s == "" {
		return 0, errMissingIndexID
	}

	v, err := strparse.IndexID(s)

	if err != nil {
		return 0, fiber.NewError(code.BadRequest, strconv.Quote(s)+" is not a valid index ID")
	}

	return v, nil
}

func parseTopicID(s string) (model.TopicIDType, error) {
	if s == "" {
		return 0, errMissingSubjectID
	}

	v, err := strparse.TopicID(s)

	if err != nil {
		return 0, fiber.NewError(code.BadRequest, strconv.Quote(s)+" is not valid topic ID")
	}

	return v, nil
}
