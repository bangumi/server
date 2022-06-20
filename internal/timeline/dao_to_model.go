// Copyright (c) 2022 TWT <TWT2333@outlook.com>
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

package timeline

import (
	"fmt"
	"reflect"

	"github.com/elliotchance/phpserialize"
	"github.com/mitchellh/mapstructure"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

//nolint:gomnd
func daoToModel(tl *dao.TimeLine) (model.TimeLine, error) {
	memo, err := memoToModel(tl)
	if err != nil {
		return model.TimeLine{}, err
	}

	img, err := imageToModel(tl.Img)
	if err != nil {
		return model.TimeLine{}, err
	}

	return model.TimeLine{
		ID:       tl.ID,
		Related:  tl.Related,
		Memo:     memo,
		Image:    img,
		UID:      tl.UID,
		Replies:  tl.Replies,
		Dateline: tl.Dateline,
		Cat:      tl.Cat,
		Type:     tl.Type,
		Batch:    tl.Batch,
		Source:   tl.Source,
	}, nil
}

func memoToModel(tl *dao.TimeLine) (model.TimeLineContent, error) {
	var memo model.TimeLineContent
	var err error

	switch {
	case tl.Cat == 1 && tl.Type == 2: // relation
		memo, err = parseRelationMemo(tl.Memo)
	case tl.Cat == 1 && (tl.Type == 3 || tl.Type == 4): // group
		memo, err = parseGroupMemo(tl.Memo)
	case tl.Cat == 2: // wiki
		memo, err = parseWikiMemo(tl.Memo)
	case tl.Cat == 3: // Subject
		memo, err = parseSubjectMemo(tl.Memo)
	case tl.Cat == 4: // progress
		memo.TimeLineProgressMemo = &model.TimeLineProgressMemo{}
		err = memo.TimeLineProgressMemo.FromBytes(tl.Memo)
	case tl.Cat == 5: // say
		memo.TimeLineSayMemo = (*model.TimeLineSayMemo)(new(string))
		err = memo.TimeLineSayMemo.FromBytes(tl.Memo)
	case tl.Cat == 6: // blog
		memo, err = parseBlogMemo(tl.Memo)
	case tl.Cat == 7: // index
		memo, err = parseIndexMemo(tl.Memo)
	case tl.Cat == 8: // mono
		memo, err = parseMonoMemo(tl.Memo)
	case tl.Cat == 9: // doujin
		memo, err = parseDoujinMemo(tl.Memo)
	default:
		err = fmt.Errorf("unexpected cat:%d type:%d", tl.Cat, tl.Type)
	}
	return memo, err
}

func imageToModel(b []byte) (model.TimeLineImages, error) {
	if len(b) == 0 {
		return model.TimeLineImages{}, nil
	}

	data, err := phpserialize.UnmarshalAssociativeArray(b)
	if err != nil {
		return nil, fmt.Errorf("phpserialize.UnmarshalAssociativeArray: %w: %s", err, string(b))
	}

	var o model.TimeLineImage
	err = decodeMap(data, &o)

	return model.TimeLineImages{o}, err
}

func imagesToModel(b []byte) (model.TimeLineImages, error) {
	if len(b) == 0 {
		return model.TimeLineImages{}, nil
	}

	data, err := phpserialize.UnmarshalAssociativeArray(b)
	if err != nil {
		return model.TimeLineImages{}, fmt.Errorf("phpserialize.UnmarshalAssociativeArray: %w: %s", err, string(b))
	}

	var result = make(model.TimeLineImages, 0, len(data))
	for _, d := range data {
		var o model.TimeLineImage
		if err = decodeMap(d, &o); err != nil {
			return nil, errgo.Wrap(err, "decodeMap")
		}
		result = append(result, o)
	}

	return result, nil
}

func parseWikiMemo(b []byte) (model.TimeLineContent, error) {
	var o WikiMemo
	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseBlogMemo(b []byte) (model.TimeLineContent, error) {
	var o BlogMemo
	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseGroupMemo(b []byte) (model.TimeLineContent, error) {
	var o GroupMemo
	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseSubjectMemo(b []byte) (model.TimeLineContent, error) {
	var o SubjectMemo
	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseRelationMemo(b []byte) (model.TimeLineContent, error) {
	var o RelationMemo
	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseIndexMemo(b []byte) (model.TimeLineContent, error) {
	var o IndexMemo
	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseMonoMemo(b []byte) (model.TimeLineContent, error) {
	var o MenoMemo
	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseDoujinMemo(b []byte) (model.TimeLineContent, error) {
	var o DoujinMemo

	err := decodeBytes(b, &o)

	return model.TimeLineContent{}, err
}

func parseDoujinMemoBatch(b []byte) ([]DoujinMemo, error) {
	data, err := phpserialize.UnmarshalAssociativeArray(b)
	if err != nil {
		return nil, fmt.Errorf("phpserialize.UnmarshalAssociativeArray: %w: %s", err, string(b))
	}

	var r = make([]DoujinMemo, 0, len(data))

	for _, d := range data {
		var o DoujinMemo
		err = decodeMap(d, &o)
		if err != nil {
			return nil, err
		}

		r = append(r, o)
	}

	return r, err
}

func decodeBytes(b []byte, output interface{}) error {
	data, err := phpserialize.UnmarshalAssociativeArray(b)
	if err != nil {
		return fmt.Errorf("phpserialize.UnmarshalAssociativeArray: %w: %s", err, string(b))
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused:          true,
		ErrorUnset:           false,
		WeaklyTypedInput:     true,
		Result:               output,
		TagName:              "json",
		IgnoreUntaggedFields: true,
	})

	if err != nil {
		return errgo.Wrap(err, "mapstructure.NewDecoder")
	}

	err = decoder.Decode(data)
	if err != nil {
		return errgo.Wrap(err, "mapstructure.Decode: "+reflect.TypeOf(output).String())
	}

	return nil
}

func decodeMap(m interface{}, output interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused:          true,
		ErrorUnset:           false,
		WeaklyTypedInput:     true,
		Result:               output,
		TagName:              "json",
		IgnoreUntaggedFields: true,
	})

	if err != nil {
		return errgo.Wrap(err, "mapstructure.NewDecoder")
	}

	err = decoder.Decode(m)
	if err != nil {
		return errgo.Wrap(err, "mapstructure.Decode: "+reflect.TypeOf(output).String())
	}

	return nil
}
