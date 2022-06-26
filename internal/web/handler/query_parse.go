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
	"strconv"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/internal/web/res"
)

// these errors result in to 400 http response.
var errMissingCharacterID = res.BadRequest("character ID is required")
var errMissingSubjectID = res.BadRequest("subject ID is required")
var errMissingPersonID = res.BadRequest("person ID is required")
var errMissingEpisodeID = res.BadRequest("episode ID is required")
var errMissingIndexID = res.BadRequest("index ID is required")

func parseSubjectType(s string) (uint8, error) {
	if s == "" {
		return 0, nil
	}

	t, err := strparse.Uint8(s)
	if err != nil {
		return 0, res.BadRequest("bad subject type: " + strconv.Quote(s))
	}

	switch t {
	case model.SubjectAnime, model.SubjectBook,
		model.SubjectMusic, model.SubjectReal, model.SubjectGame:
		return t, nil
	}

	return 0, res.BadRequest(strconv.Quote(s) + " is not a valid subject type")
}

func parseSubjectID(s string) (model.SubjectID, error) {
	if s == "" {
		return 0, errMissingSubjectID
	}

	v, err := strparse.SubjectID(s)

	if err != nil {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid subject ID")
	}

	return v, nil
}

func parseCharacterID(s string) (model.CharacterID, error) {
	if s == "" {
		return 0, errMissingCharacterID
	}

	v, err := strparse.CharacterID(s)

	if err != nil {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid character ID")
	}

	return v, nil
}

func parsePersonID(s string) (model.PersonID, error) {
	if s == "" {
		return 0, errMissingPersonID
	}

	v, err := strparse.PersonID(s)

	if err != nil {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid person ID")
	}

	return v, nil
}

func parseEpisodeID(s string) (model.EpisodeID, error) {
	if s == "" {
		return 0, errMissingEpisodeID
	}

	v, err := strparse.EpisodeID(s)

	if err != nil {
		return 0, res.BadRequest(strconv.Quote(s) + " is not a valid episode ID")
	}

	return v, nil
}

func parseIndexID(s string) (model.IndexID, error) {
	if s == "" {
		return 0, errMissingIndexID
	}

	v, err := strparse.IndexID(s)

	if err != nil {
		return 0, res.BadRequest(strconv.Quote(s) + " is not a valid index ID")
	}

	return v, nil
}

func parseCollectionType(s string) (model.CollectionType, error) {
	if s == "" {
		return model.CollectionTypeAll, nil
	}

	t, err := strparse.Uint8(s)
	if err != nil {
		return 0, res.BadRequest("bad collection type: " + strconv.Quote(s))
	}

	v := model.CollectionType(t)
	switch v {
	case model.CollectionTypeWish,
		model.CollectionTypeDone,
		model.CollectionTypeDoing,
		model.CollectionTypeOnHold,
		model.CollectionTypeDropped:
		return v, nil
	case model.CollectionTypeAll:
	}

	return 0, res.BadRequest(strconv.Quote(s) + "is not a valid collection type")
}
