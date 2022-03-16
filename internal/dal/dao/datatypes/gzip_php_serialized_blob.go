// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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

package datatypes

import (
	"bytes"
	"compress/flate"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/elliotchance/phpserialize"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/bangumi/server/internal/errgo"
)

type GzipPhpSerializedBlob map[string]interface{}

var errDefault = errors.New("")

func decompress(data []byte) ([]byte, error) {
	gr := flate.NewReader(bytes.NewBuffer(data))
	defer gr.Close()
	bytes, err := io.ReadAll(gr)
	if err != nil {
		return nil, errgo.Wrap(err, "decompress")
	}
	return bytes, nil
}

func toValidJSON(data interface{}) interface{} {
	t := reflect.TypeOf(data).Kind()
	switch t {
	case reflect.Array:
	case reflect.Slice:
		if arr, ok := data.([]interface{}); ok {
			for i, val := range arr {
				arr[i] = toValidJSON(val)
			}
			return arr
		}
	case reflect.Map:
		if m, ok := data.(map[interface{}]interface{}); ok {
			ret := map[string]interface{}{}
			for k, v := range m {
				ret[fmt.Sprint(k)] = toValidJSON(v)
			}
			return ret
		}
	default:
	}
	return data
}

func (b *GzipPhpSerializedBlob) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errgo.Wrap(errDefault, fmt.Sprint("Failed to unmarshal MEDIUMBLOB value:", value))
	}
	bytes, err := decompress(data)
	if err != nil {
		return err
	}
	result, err := phpserialize.UnmarshalAssociativeArray(bytes)
	if d, ok := toValidJSON(result).(map[string]interface{}); ok {
		*b = d
	}
	return errgo.Wrap(err, "phpserialize.UnmarshalAssociativeArray")
}

func (b GzipPhpSerializedBlob) Value() (driver.Value, error) {
	return nil, errgo.Wrap(errDefault, "write to db is not supported now")
}

func (b GzipPhpSerializedBlob) GormDataType() string {
	return "GzipPhpSerializedBlob"
}

func (b GzipPhpSerializedBlob) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "MEDIUMBLOB"
}
