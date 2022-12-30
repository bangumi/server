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
	"strconv"
	"time"

	"github.com/samber/lo"
	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/timeline/image"
	"github.com/bangumi/server/internal/timeline/memo"
)

func (m mysqlRepo) ChangeEpisodeStatus(
	ctx context.Context,
	u auth.Auth,
	sbj model.Subject,
	episode episode.Episode,
	update collection.Update,
) error {
	// TODO: merge previous timeline
	if !update.EpStatus.Set && !update.VolStatus.Set {
		return nil
	}

	var VolsTotal = "??"
	if sbj.Volumes != 0 {
		VolsTotal = strconv.FormatUint(uint64(sbj.Volumes), 10)
	}
	var EpsTotal = "??"
	if sbj.Eps != 0 {
		EpsTotal = strconv.FormatUint(uint64(sbj.Eps), 10)
	}

	var VolsUpdate *int
	if update.VolStatus.Set {
		VolsUpdate = lo.ToPtr[int](int(update.VolStatus.Value))
	}

	var EpsUpdate *int
	if update.EpStatus.Set {
		EpsUpdate = lo.ToPtr[int](int(update.EpStatus.Value))
	}

	mm := memo.ProgressMemo{
		VolsUpdate: VolsUpdate,
		EpsUpdate:  EpsUpdate,

		VolsTotal: &VolsTotal,
		EpsTotal:  &EpsTotal,

		EpName:        &episode.Name,
		EpSort:        &episode.Sort,
		EpID:          &episode.ID,
		SubjectID:     &sbj.ID,
		SubjectName:   &sbj.Name,
		SubjectTypeID: &sbj.TypeID,
	}

	b, err := phpserialize.Marshal(mm)
	if err != nil {
		return errgo.Wrap(err, "marshal")
	}

	img, err := phpserialize.Marshal(image.Subject{
		SubjectID: sbj.ID,
		Images:    sbj.Image,
	})
	if err != nil {
		return errgo.Wrap(err, "marshal")
	}

	err = m.q.TimeLine.WithContext(ctx).Create(&dao.TimeLine{
		UID:      u.ID,
		Cat:      model.TimeLineCatProgress,
		Type:     0,
		Related:  sbj.ID.String(),
		Memo:     b,
		Img:      img,
		Batch:    0,
		Source:   0,
		Dateline: uint32(time.Now().Unix()),
	})

	return errgo.Wrap(err, "dal.create")
}
