// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
//
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

package compat

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/elliotchance/phpserialize"

	"github.com/bangumi/server/internal/errgo"
)

type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

var ErrTypeCase = errors.New("can't cast to expected type")

func ParseTags(s []byte) (t []Tag, err error) {
	if len(s) == 0 {
		return []Tag{}, nil
	}

	in, err := phpserialize.UnmarshalIndexedArray(s)
	if err != nil {
		return nil, errgo.Wrap(err, "php unmarshal")
	}

	t = make([]Tag, 0, len(in))

	for _, tag := range in {
		v, ok := tag.(map[interface{}]interface{})
		if !ok {
			return nil, errgo.Msg(ErrTypeCase,
				fmt.Sprintf("failed to cast %v to map[interface{}]interface{} ", tag))
		}

		name, ok := v["tag_name"].(string)
		if !ok {
			// v["tag_name"] maybe nil, just skip this case
			continue
		}

		countRaw, ok := v["result"].(string)
		if !ok {
			return nil, errgo.Msg(ErrTypeCase, fmt.Sprintf("failed to cast %v to string ", v["result"]))
		}

		count, err := strconv.Atoi(countRaw)
		if err != nil {
			return nil, errgo.Wrap(err, "atoi")
		}

		t = append(t, Tag{
			Name:  name,
			Count: count,
		})
	}

	sort.Slice(t, func(i, j int) bool {
		if t[i].Count == t[j].Count {
			return t[i].Name <= t[j].Name
		}

		return t[i].Count > t[j].Count
	})

	return t, nil
}
