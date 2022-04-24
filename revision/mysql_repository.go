// Copyright (c) 2022 TWT <TWT2333@outlook.com>
// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
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

package revision

import (
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/elliotchance/phpserialize"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.RevisionRepo, error) {
	return mysqlRepo{q: q, log: log.Named("revision.mysqlRepo")}, nil
}

func (r mysqlRepo) CountPersonRelated(ctx context.Context, id model.PersonIDType) (int64, error) {
	c, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.Mid.Eq(id), r.q.RevisionHistory.Type.In(model.PersonRevisionTypes()...)).Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}
	return c, nil
}

func (r mysqlRepo) ListPersonRelated(
	ctx context.Context, personID model.PersonIDType, limit int, offset int,
) ([]model.Revision, error) {
	revisions, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.Mid.Eq(personID), r.q.RevisionHistory.Type.In(model.PersonRevisionTypes()...)).
		Order(r.q.RevisionHistory.ID.Desc()).
		Limit(limit).
		Offset(offset).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	result := make([]model.Revision, len(revisions))
	for i, revision := range revisions {
		result[i] = convertRevisionDao(revision, nil)
	}
	return result, nil
}

func (r mysqlRepo) GetPersonRelated(ctx context.Context, id model.IDType) (model.Revision, error) {
	revision, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.ID.Eq(id),
			r.q.RevisionHistory.Type.In(model.PersonRevisionTypes()...)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Revision{}, domain.ErrNotFound
		}
		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Revision{}, errgo.Wrap(err, "dal")
	}
	data, err := r.q.RevisionText.WithContext(ctx).
		Where(r.q.RevisionText.TextID.Eq(revision.TextID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("can't find revision text", zap.Uint32("id", revision.TextID))
			return model.Revision{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Revision{}, errgo.Wrap(err, "dal")
	}
	return convertRevisionDao(revision, data), nil
}

func (r mysqlRepo) CountCharacterRelated(ctx context.Context, characterID model.CharacterIDType) (int64, error) {
	c, err := r.q.RevisionHistory.WithContext(ctx).
		Where(
			r.q.RevisionHistory.Mid.Eq(characterID),
			r.q.RevisionHistory.Type.In(model.CharacterRevisionTypes()...),
		).Count()
	return c, wrapGORMError(err)
}

func (r mysqlRepo) ListCharacterRelated(
	ctx context.Context, characterID model.CharacterIDType, limit int, offset int,
) ([]model.CharacterRevision, error) {
	revisions, err := r.q.RevisionHistory.WithContext(ctx).
		Where(
			r.q.RevisionHistory.Mid.Eq(characterID),
			r.q.RevisionHistory.Type.In(model.CharacterRevisionTypes()...),
		).
		Order(r.q.RevisionHistory.ID.Desc()).
		Limit(limit).
		Offset(offset).Find()
	if err != nil {
		return nil, wrapGORMError(err)
	}

	result := make([]model.CharacterRevision, 0, len(revisions))
	for _, revision := range revisions {
		result = append(result, convertCharacterRevisionDao(revision, nil))
	}
	return result, nil
}

func (r mysqlRepo) GetCharacterRelated(ctx context.Context, id model.IDType) (model.CharacterRevision, error) {
	revision, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.ID.Eq(id),
			r.q.RevisionHistory.Type.In(model.CharacterRevisionTypes()...)).
		First()
	if err != nil {
		return model.CharacterRevision{}, wrapGORMError(err)
	}

	data, err := r.q.RevisionText.WithContext(ctx).
		Where(r.q.RevisionText.TextID.Eq(revision.TextID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("can't find revision text", zap.Uint32("id", revision.TextID))
			return model.CharacterRevision{}, domain.ErrNotFound
		}
		r.log.Error("unexpected error happened", zap.Error(err))
		return model.CharacterRevision{}, errgo.Wrap(err, "dal")
	}
	return convertCharacterRevisionDao(revision, data), nil
}

func wrapGORMError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	return errgo.Wrap(err, "dal")
}

func (r mysqlRepo) CountSubjectRelated(ctx context.Context, id model.SubjectIDType) (int64, error) {
	c, err := r.q.SubjectRevision.WithContext(ctx).
		Where(r.q.SubjectRevision.SubjectID.Eq(id)).Count()
	if err != nil {
		return 0, errgo.Wrap(err, "dal")
	}
	return c, nil
}

func (r mysqlRepo) ListSubjectRelated(
	ctx context.Context, id model.SubjectIDType, limit int, offset int,
) ([]model.Revision, error) {
	revisions, err := r.q.SubjectRevision.WithContext(ctx).
		Where(r.q.SubjectRevision.SubjectID.Eq(id)).
		Order(r.q.SubjectRevision.ID.Desc()).
		Limit(limit).
		Offset(offset).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	result := make([]model.Revision, len(revisions))
	for i, revision := range revisions {
		result[i] = convertSubjectRevisionDao(revision, false)
	}
	return result, nil
}

func (r mysqlRepo) GetSubjectRelated(ctx context.Context, id model.IDType) (model.Revision, error) {
	revision, err := r.q.SubjectRevision.WithContext(ctx).
		Where(r.q.SubjectRevision.ID.Eq(id)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Revision{}, domain.ErrNotFound
		}

		r.log.Error("unexpected error happened", zap.Error(err))
		return model.Revision{}, errgo.Wrap(err, "dal")
	}
	return convertSubjectRevisionDao(revision, true), nil
}

func toValidJSON(data interface{}) interface{} {
	if data == nil {
		return nil
	}
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

func convertRevisionText(text []byte) map[string]interface{} {
	gr := flate.NewReader(bytes.NewBuffer(text))
	defer gr.Close()
	b, err := io.ReadAll(gr)
	if err != nil {
		return nil
	}
	result, err := phpserialize.UnmarshalAssociativeArray(b)
	if err != nil {
		return nil
	}
	if d, ok := toValidJSON(result).(map[string]interface{}); ok {
		return d
	}
	return nil
}

func safeDecodeExtra(k1 reflect.Type, k2 reflect.Type, input interface{}) (interface{}, error) {
	if k2.Name() == "Extra" && k1.Kind() != reflect.Map {
		return map[string]string{}, nil
	}
	return input, nil
}

func castCharacterData(raw map[string]interface{}) model.CharacterRevisionData {
	if raw == nil {
		return model.CharacterRevisionData{}
	}
	result := model.CharacterRevisionData{
		CharacterRevisionEdit: make(map[string]model.CharacterRevisionEditItem, len(raw)),
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: safeDecodeExtra,
		Result:     &result.CharacterRevisionEdit,
	})
	if err != nil {
		return model.CharacterRevisionData{}
	}
	if decoder.Decode(raw) != nil {
		return model.CharacterRevisionData{}
	}
	return result
}

func castPersonData(raw map[string]interface{}) map[string]model.PersonRevisionDataItem {
	if raw == nil {
		return nil
	}
	items := make(map[string]model.PersonRevisionDataItem, len(raw))
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "json",
		DecodeHook: safeDecodeExtra,
		Result:     &items,
	})
	if err != nil {
		return nil
	}
	if err := decoder.Decode(raw); err != nil {
		return nil
	}
	return items
}

func convertRevisionDao(r *dao.RevisionHistory, data *dao.RevisionText) model.Revision {
	var text map[string]model.PersonRevisionDataItem
	if data != nil {
		text = castPersonData(convertRevisionText(data.Text))
	}

	return model.Revision{
		ID:        r.ID,
		Type:      r.Type,
		Summary:   r.Summary,
		CreatorID: r.CreatorID,
		CreatedAt: time.Unix(int64(r.CreatedAt), 0),
		Data:      text,
	}
}

func convertCharacterRevisionDao(r *dao.RevisionHistory, text *dao.RevisionText) model.CharacterRevision {
	var data model.CharacterRevisionData
	if text != nil {
		data = castCharacterData(convertRevisionText(text.Text))
	}

	return model.CharacterRevision{
		RevisionCommon: model.RevisionCommon{
			ID:        r.ID,
			Type:      r.Type,
			Summary:   r.Summary,
			CreatorID: r.CreatorID,
			CreatedAt: time.Unix(int64(r.CreatedAt), 0),
		},
		Data: data,
	}
}

func convertSubjectRevisionDao(r *dao.SubjectRevision, isDetailed bool) model.Revision {
	var data *model.SubjectRevisionData
	if isDetailed {
		data = &model.SubjectRevisionData{
			SubjectID:    r.SubjectID,
			Name:         r.Name,
			NameCN:       r.NameCN,
			VoteField:    r.VoteField,
			Type:         r.Type,
			TypeID:       r.TypeID,
			FieldInfobox: r.FieldInfobox,
			FieldSummary: r.FieldSummary,
			FieldEps:     r.FieldEps,
			Platform:     r.Platform,
		}
	}

	return model.Revision{
		ID:        r.ID,
		Type:      r.Type,
		Summary:   r.EditSummary,
		CreatorID: r.Creator,
		CreatedAt: time.Unix(int64(r.Dateline), 0),
		Data:      data,
	}
}
