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

package collection

// handle php serialization

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/elliotchance/phpserialize"
	ms "github.com/mitchellh/mapstructure"

	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

type mysqlEpCollectionItem struct {
	EpisodeID model.EpisodeID      `ms:"eid" php:"eid"`
	Type      model.CollectionType `ms:"type" php:"type"`
}

var errEpisodeInvalid = errors.New("number is not valid as episode ID")

type mysqlEpCollection = map[model.EpisodeID]mysqlEpCollectionItem

func deserializePhpEpStatus(phpSerialized []byte) (mysqlEpCollection, error) {
	var e map[interface{}]interface{}
	if err := phpserialize.Unmarshal(phpSerialized, &e); err != nil {
		return nil, errgo.Wrap(err, "php deserialize")
	}

	var ep = make(mysqlEpCollection, len(e))
	for key, value := range e {
		iKey, ok := key.(int64)
		if !ok {
			//nolint:goerr113
			return nil, fmt.Errorf("failed to convert type %s to int64, value %v", reflect.TypeOf(key).String(), key)
		}
		if iKey <= 0 || iKey > math.MaxUint32 {
			return nil, errgo.Wrap(errEpisodeInvalid, strconv.FormatInt(iKey, 10))
		}

		e, err := decodePhpItem(value)
		if err != nil {
			return nil, err
		}

		ep[model.EpisodeID(iKey)] = e
	}

	return ep, nil
}

func serializePhpEpStatus(data mysqlEpCollection) ([]byte, error) {
	var e = make(map[interface{}]interface{}, len(data))
	// have to convert struct back to map, so phpserialize marshal it to php array instead of php object.
	for key, value := range data {
		e[int64(key)] = map[interface{}]interface{}{
			"eid":  strconv.FormatUint(uint64(value.EpisodeID), 10),
			"type": value.Type,
		}
	}

	b, err := phpserialize.Marshal(e, nil)
	return b, errgo.Wrap(err, "phpserialize.Marshal")
}

// convert map[any]any to mysqlEpCollectionItem.
func decodePhpItem(value interface{}) (mysqlEpCollectionItem, error) {
	var e mysqlEpCollectionItem
	decoder, err := ms.NewDecoder(&ms.DecoderConfig{
		ErrorUnused:          true,
		ErrorUnset:           true,
		WeaklyTypedInput:     true,
		Result:               &e,
		TagName:              "ms",
		IgnoreUntaggedFields: true,
	})
	if err != nil {
		return e, errgo.Wrap(err, "mapstructure.MewDecoder")
	}

	err = decoder.Decode(value)
	return e, errgo.Wrap(err, "mapstructure.Decode")
}
