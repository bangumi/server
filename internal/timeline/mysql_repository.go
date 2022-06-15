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
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/elliotchance/phpserialize"
	"github.com/gookit/goutil/dump"
	ms "github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.TimeLineRepo, error) {
	return mysqlRepo{q: q, log: log.Named("timeline.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetByID(ctx context.Context, id model.TimeLineID) (model.TimeLine, error) {
	record, err := m.q.TimeLine.WithContext(ctx).Where(m.q.TimeLine.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.TimeLine{}, domain.ErrNotFound
		}

		m.log.Error("unexpected happened", zap.Error(err))
		return model.TimeLine{}, errgo.Wrap(err, "dal")
	}

	return convertDAO(record)
}

//nolint:gomnd,gocyclo
func convertBatchDao(record *dao.TimeLine) (model.TimeLine, error) {
	var memo model.TimeLineContent
	var err error

	switch {
	case record.Cat == 3: // Subject
	case record.Cat == 1 && record.Type == 2: // relation
		memo, err = parseRelationMemo(record.Memo)
	case record.Cat == 1 && (record.Type == 3 || record.Type == 4): // group
		memo, err = parseGroupMemo(record.Memo)
	case record.Cat == 8: // mono
		memo, err = parseMenoMemo(record.Memo)
	case record.Cat == 9: // doujin
		memo, err = parseDoujinMemo(record.Memo)
	default:
		panic(dump.Format(record.Cat, record.Type))
	}

	if err != nil {
		panic(err)
	}

	img, err := parseImages(record.Img)
	if err != nil {
		return model.TimeLine{}, err
	}

	return model.TimeLine{
		Image:    img,
		ID:       record.ID,
		Related:  record.Related,
		Memo:     memo,
		UID:      record.UID,
		Replies:  record.Replies,
		Dateline: record.Dateline,
		Cat:      record.Cat,
		Type:     record.Type,
		Batch:    record.Batch,
		Source:   record.Source,
	}, nil
}

//nolint:gomnd
func convertDAO(record *dao.TimeLine) (model.TimeLine, error) {
	if record.Batch != 0 {
		return convertBatchDao(record)
	}

	var memo model.TimeLineContent
	var err error

	switch {
	case record.Cat == 3: // Subject
		memo, err = parseSubjectMemo(record.Memo)
	case record.Cat == 4: // progress
		memo, err = parseProgressMeno(record.Memo)
	case record.Cat == 1 && record.Type == 2: // relation
		memo, err = parseRelationMemo(record.Memo)
	case record.Cat == 1 && (record.Type == 3 || record.Type == 4): // group
		memo, err = parseGroupMemo(record.Memo)
	case record.Cat == 5: // say
		memo = parseSayMemo(record.Memo)
	case record.Cat == 2: // wiki
		memo, err = parseWikiMemo(record.Memo)
	case record.Cat == 6: // blog
		memo, err = parseBlogMemo(record.Memo)
	case record.Cat == 7: // index
		memo, err = parseIndexMemo(record.Memo)
	case record.Cat == 8: // mono
		memo, err = parseMenoMemo(record.Memo)
	case record.Cat == 9: // doujin
		memo, err = parseDoujinMemo(record.Memo)
	default:
		panic(dump.Format(record.Cat, record.Type))
	}

	if err != nil {
		panic(err)
	}

	img, err := parseImage(record.Img)
	if err != nil {
		return model.TimeLine{}, err
	}

	return model.TimeLine{
		Image:    img,
		ID:       record.ID,
		Related:  record.Related,
		Memo:     memo,
		UID:      record.UID,
		Replies:  record.Replies,
		Dateline: record.Dateline,
		Cat:      record.Cat,
		Type:     record.Type,
		Batch:    record.Batch,
		Source:   record.Source,
	}, nil
}

func parseImage(b []byte) ([]model.TimelineImage, error) {
	if len(b) == 0 {
		return []model.TimelineImage{}, nil
	}

	data, err := phpserialize.UnmarshalAssociativeArray(b)
	if err != nil {
		return nil, fmt.Errorf("phpserialize.UnmarshalAssociativeArray: %w: %s", err, string(b))
	}

	var o model.TimelineImage
	err = decodeMap(data, &o)

	return []model.TimelineImage{o}, err
}

func parseImages(b []byte) ([]model.TimelineImage, error) {
	if len(b) == 0 {
		return []model.TimelineImage{}, nil
	}

	data, err := phpserialize.UnmarshalAssociativeArray(b)
	if err != nil {
		return nil, fmt.Errorf("phpserialize.UnmarshalAssociativeArray: %w: %s", err, string(b))
	}

	var result = make([]model.TimelineImage, 0, len(data))

	for _, d := range data {
		var o model.TimelineImage
		err = decodeMap(d, &o)
		if err != nil {
			return nil, err
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

func parseSayMemo(b []byte) model.TimeLineContent {
	return model.TimeLineContent{
		Text: string(b),
	}
}

func parseProgressMeno(b []byte) (model.TimeLineContent, error) {
	var o ProgressMemo
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

func parseMenoMemo(b []byte) (model.TimeLineContent, error) {
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
		err := decodeMap(d, &o)
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

	decoder, err := ms.NewDecoder(&ms.DecoderConfig{
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
	decoder, err := ms.NewDecoder(&ms.DecoderConfig{
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
