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

package strparse

import (
	"errors"
	"strconv"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func Uint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)

	return uint8(v), errgo.Wrap(err, "strconv")
}

func Uint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)

	return uint32(v), errgo.Wrap(err, "strconv")
}

func uint32NonZero(s string) (uint32, error) {
	v, err := Uint32(s)
	if err != nil {
		return 0, errgo.Wrap(err, "strconv")
	}

	if v == 0 {
		return 0, errZeroValue
	}

	return v, nil
}

func UserID(s string) (model.UserID, error) {
	v, err := uint32NonZero(s)

	return model.UserID(v), err
}

func SubjectID(s string) (model.SubjectID, error) {
	v, err := uint32NonZero(s)

	return model.SubjectID(v), err
}

func PersonID(s string) (model.PersonID, error) {
	v, err := uint32NonZero(s)

	return model.PersonID(v), err
}

func CharacterID(s string) (model.CharacterID, error) {
	v, err := uint32NonZero(s)

	return model.CharacterID(v), err
}

func EpisodeID(s string) (model.EpisodeID, error) {
	v, err := uint32NonZero(s)

	return model.EpisodeID(v), err
}

func IndexID(s string) (model.IndexID, error) {
	return uint32NonZero(s)
}

func TopicID(s string) (model.TopicID, error) {
	v, err := uint32NonZero(s)

	return model.TopicID(v), err
}

var errZeroValue = errors.New("can't use 0 as ID")
