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

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/web/res"
)

// these errors result in to 400 http response.
var errMissingSubjectID = res.BadRequest("subject ID is required")
var errMissingCharacterID = res.BadRequest("character ID is required")
var errMissingPersonID = res.BadRequest("person ID is required")
var errMissingEpisodeID = res.BadRequest("episode ID is required")
var errMissingIndexID = res.BadRequest("index ID is required")
var errMissingTopicID = res.BadRequest("topic ID is required")

func ParseSubjectType(s string) (model.SubjectType, error) {
	if s == "" {
		return 0, nil
	}

	t, err := gstr.ParseUint8(s)
	if err != nil {
		return 0, res.BadRequest("bad subject type: " + strconv.Quote(s))
	}

	switch t {
	case model.SubjectTypeAnime, model.SubjectTypeBook,
		model.SubjectTypeMusic, model.SubjectTypeReal, model.SubjectTypeGame:
		return t, nil
	}

	return 0, res.BadRequest(strconv.Quote(s) + " is not a valid subject type")
}

func ParseSubjectID(s string) (model.SubjectID, error) {
	if s == "" {
		return 0, errMissingSubjectID
	}

	v, err := gstr.ParseUint32(s)

	if err != nil || v == 0 {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid subject ID")
	}

	return model.SubjectID(v), nil
}

func ParseCharacterID(s string) (model.CharacterID, error) {
	if s == "" {
		return 0, errMissingCharacterID
	}

	v, err := gstr.ParseUint32(s)

	if err != nil || v == 0 {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid character ID")
	}

	return model.CharacterID(v), nil
}

func ParsePersonID(s string) (model.PersonID, error) {
	if s == "" {
		return 0, errMissingPersonID
	}

	v, err := gstr.ParseUint32(s)

	if err != nil || v == 0 {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid person ID")
	}

	return model.PersonID(v), nil
}

func ParseEpisodeID(s string) (model.EpisodeID, error) {
	if s == "" {
		return 0, errMissingEpisodeID
	}

	v, err := gstr.ParseUint32(s)

	if err != nil || v == 0 {
		return 0, res.BadRequest(strconv.Quote(s) + " is not a valid episode ID")
	}

	return model.EpisodeID(v), nil
}

func ParseIndexID(s string) (model.IndexID, error) {
	if s == "" {
		return 0, errMissingIndexID
	}

	v, err := gstr.ParseUint32(s)

	if err != nil || v == 0 {
		return 0, res.BadRequest(strconv.Quote(s) + " is not a valid index ID")
	}

	return v, nil
}

func ParseTopicID(s string) (model.TopicID, error) {
	if s == "" {
		return 0, errMissingTopicID
	}

	v, err := gstr.ParseUint32(s)

	if err != nil || v == 0 {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid topic ID")
	}

	return model.TopicID(v), nil
}

func ParseCollectionType(s string) (model.CollectionType, error) {
	if s == "" {
		return model.CollectionTypeAll, nil
	}

	t, err := gstr.ParseUint8(s)
	if err != nil {
		return 0, res.BadRequest("bad collection type: " + strconv.Quote(s))
	}

	v := model.CollectionType(t)
	switch v {
	case model.CollectionTypeAll,
		model.CollectionTypeWish,
		model.CollectionTypeDone,
		model.CollectionTypeDoing,
		model.CollectionTypeOnHold,
		model.CollectionTypeDropped:
		return v, nil
	}

	return 0, res.BadRequest(strconv.Quote(s) + "is not a valid collection type")
}
