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
	"context"
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
)

func (r mysqlRepo) CountEpisodeRelated(ctx context.Context, episodeID model.EpisodeID) (int64, error) {
	c, err := r.q.RevisionHistory.WithContext(ctx).
		Where(
			r.q.RevisionHistory.Mid.Eq(episodeID),
			r.q.RevisionHistory.Type.In(model.EpisodeRevisionTypes()...),
		).Count()
	return c, wrapGORMError(err)
}

func (r mysqlRepo) ListEpisodeRelated(
	ctx context.Context,
	episodeID model.EpisodeID,
	limit int, offset int,
) ([]model.EpisodeRevision, error) {
	revisions, err := r.q.RevisionHistory.WithContext(ctx).
		Where(
			r.q.RevisionHistory.Mid.Eq(episodeID),
			r.q.RevisionHistory.Type.In(model.EpisodeRevisionTypes()...),
		).
		Order(r.q.RevisionHistory.ID.Desc()).
		Limit(limit).
		Offset(offset).Find()
	if err != nil {
		return nil, wrapGORMError(err)
	}

	result := make([]model.EpisodeRevision, 0, len(revisions))
	for _, revision := range revisions {
		result = append(result, convertEpisodeRevisionDao(revision, nil))
	}
	return result, nil
}

func (r mysqlRepo) GetEpisodeRelated(ctx context.Context, id model.RevisionID) (model.EpisodeRevision, error) {
	revision, err := r.q.RevisionHistory.WithContext(ctx).
		Where(r.q.RevisionHistory.ID.Eq(id), r.q.RevisionHistory.Type.In(model.EpisodeRevisionTypes()...)).
		First()
	if err != nil {
		return model.EpisodeRevision{}, wrapGORMError(err)
	}

	data, err := r.q.RevisionText.WithContext(ctx).Where(r.q.RevisionText.TextID.Eq(revision.TextID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Error("can't find revision text", zap.Uint32("id", revision.TextID))
			return model.EpisodeRevision{}, gerr.ErrNotFound
		}
		r.log.Error("unexpected error happened", zap.Error(err))
		return model.EpisodeRevision{}, errgo.Wrap(err, "dal")
	}

	return convertEpisodeRevisionDao(revision, data), nil
}

func castEpisodeData(raw map[string]any) model.EpisodeRevisionData {
	if raw == nil {
		return nil
	}

	result := make(map[string]model.EpisodeRevisionDataItem, len(raw))
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       safeDecodeExtra,
		Result:           &result,
		WeaklyTypedInput: true,
	})
	if err != nil || decoder.Decode(raw) != nil {
		return nil
	}
	return result
}

func convertEpisodeRevisionDao(r *dao.RevisionHistory, text *dao.RevisionText) model.EpisodeRevision {
	var data model.EpisodeRevisionData
	if text != nil {
		data = castEpisodeData(convertRevisionText(text.Text))
	}

	return model.EpisodeRevision{
		RevisionCommon: model.RevisionCommon{
			ID:        r.ID,
			Type:      r.Type,
			Summary:   r.Summary,
			CreatorID: r.CreatorID,
			CreatedAt: time.Unix(int64(r.CreatedTime), 0),
		},
		Data: data,
	}
}
