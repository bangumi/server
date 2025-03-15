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

	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/web/res"
)

// these errors result in to 400 http response.
var errMissingID = res.BadRequest("character ID is required")

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

func ParseSubjectCategory(stype model.SubjectType, s string) (uint16, error) {
	if s == "" {
		return 0, res.BadRequest("subject category is empty")
	}
	platforms, ok := vars.PlatformMap[stype]
	if !ok {
		return 0, res.BadRequest("bad subject type: " + strconv.Quote(s))
	}
	v, err := gstr.ParseUint16(s)
	if err != nil {
		return 0, res.BadRequest("bad subject category: " + strconv.Quote(s))
	}
	if _, ok := platforms[v]; !ok {
		return 0, res.BadRequest("bad subject category: " + strconv.Quote(s))
	}
	return v, nil
}

func ParseID(s string) (model.CharacterID, error) {
	if s == "" {
		return 0, errMissingID
	}

	v, err := gstr.ParseUint32(s)

	if err != nil || v == 0 {
		return 0, res.BadRequest(strconv.Quote(s) + " is not valid ID")
	}

	return v, nil
}

func ParseCollectionType(s string) (collection.SubjectCollection, error) {
	if s == "" {
		return collection.SubjectCollectionAll, nil
	}

	t, err := gstr.ParseUint8(s)
	if err != nil {
		return 0, res.BadRequest("bad collection type: " + strconv.Quote(s))
	}

	v := collection.SubjectCollection(t)
	switch v {
	case collection.SubjectCollectionAll,
		collection.SubjectCollectionWish,
		collection.SubjectCollectionDone,
		collection.SubjectCollectionDoing,
		collection.SubjectCollectionOnHold,
		collection.SubjectCollectionDropped:
		return v, nil
	}

	return 0, res.BadRequest(strconv.Quote(s) + "is not a valid collection type")
}

func ParseEpTypeOptional(s string) (*episode.Type, error) {
	if s == "" {
		return nil, nil //nolint:nilnil
	}

	v, err := gstr.ParseUint8(s)
	if err != nil {
		return nil, res.BadRequest("wrong value for query `type`")
	}

	switch v {
	case episode.TypeNormal, episode.TypeSpecial,
		episode.TypeOpening, episode.TypeEnding,
		episode.TypeMad, episode.TypeOther:
		return &v, nil
	}

	return nil, res.BadRequest(strconv.Quote(s) + " is not valid episode type")
}
