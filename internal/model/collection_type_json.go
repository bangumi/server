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

package model

import (
	"reflect"
	"strconv"

	"github.com/goccy/go-json"
)

var _ json.Unmarshaler = (*SubjectCollection)(nil)
var rtSubjectCollection = reflect.TypeOf(SubjectCollection(0))

func (s *SubjectCollection) UnmarshalJSON(bytes []byte) error {
	raw := string(bytes)

	if raw == "null" {
		return &json.UnmarshalTypeError{Value: raw, Type: rtSubjectCollection}
	}

	v, err := strconv.ParseUint(string(bytes), 10, 8)
	if err != nil {
		return &json.UnmarshalTypeError{Value: raw, Type: rtSubjectCollection}
	}

	var n = SubjectCollection(v)

	switch n { //nolint:exhaustive
	case SubjectCollectionWish,
		SubjectCollectionDone,
		SubjectCollectionDoing,
		SubjectCollectionOnHold,
		SubjectCollectionDropped:
		*s = n
		return nil
	}

	return &json.UnmarshalTypeError{Value: raw, Type: rtSubjectCollection}
}

var _ json.Unmarshaler = (*EpisodeCollection)(nil)
var rtEpisodeCollection = reflect.TypeOf(EpisodeCollection(0))

func (s *EpisodeCollection) UnmarshalJSON(bytes []byte) error {
	raw := string(bytes)

	if raw == "null" {
		return &json.UnmarshalTypeError{Value: raw, Type: rtEpisodeCollection}
	}

	v, err := strconv.ParseUint(raw, 10, 8)
	if err != nil {
		return &json.UnmarshalTypeError{Value: "null", Type: rtEpisodeCollection}
	}

	var n = EpisodeCollection(v)

	switch n { //nolint:exhaustive
	case EpisodeCollectionWish,
		EpisodeCollectionDone,
		EpisodeCollectionDropped:
		*s = n
		return nil
	}

	return &json.UnmarshalTypeError{Value: "null", Type: rtEpisodeCollection}
}
