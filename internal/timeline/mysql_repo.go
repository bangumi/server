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
	"strconv"
	"time"

	"github.com/trim21/go-phpserialize"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic"
	"github.com/bangumi/server/internal/timeline/image"
	"github.com/bangumi/server/internal/timeline/memo"
)

func (m mysqlRepo) ChangeSubjectCollection(
	ctx context.Context,
	u auth.Auth,
	sbj model.Subject,
	collect model.SubjectCollection,
	comment string,
	rate uint8,
) error {
	tlType := convSubjectType(collect, sbj.TypeID)

	if comment != "" {
		return m.changeSubjectCollection(ctx, u, sbj, tlType, comment, rate)
	}

	previous, err := m.q.TimeLine.WithContext(ctx).Where(m.q.TimeLine.UID.Eq(u.ID)).
		Limit(1).Order(m.q.TimeLine.ID.Desc()).First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return m.changeSubjectCollection(ctx, u, sbj, tlType, "", rate)
		}

		return errgo.Wrap(err, "dal")
	}

	if previous.Cat == model.TimeLineCatProgress && previous.Type == tlType {
		return m.updatePreviousSubjectCollection(ctx, previous, sbj)
	}

	return m.changeSubjectCollection(ctx, u, sbj, tlType, "", rate)
}

/*
有两种情况，更新收藏进度.
*/
func (m mysqlRepo) changeSubjectCollection(
	ctx context.Context,
	u auth.Auth,
	sbj model.Subject,
	tlType uint16,
	comment string,
	rate uint8,
) error {
	img, err := phpserialize.Marshal(image.Subject{
		SubjectID: sbj.ID,
		Images:    sbj.Image,
	})
	if err != nil {
		return errgo.Wrap(err, "marshal")
	}

	serializedMemo, err := phpserialize.Marshal(memo.SubjectMemo{
		ID:             strconv.FormatUint(uint64(sbj.ID), 10),
		TypeID:         strconv.FormatUint(uint64(sbj.TypeID), 10),
		Name:           sbj.Name,
		NameCN:         sbj.NameCN,
		Series:         strconv.Itoa(generic.BtoI(sbj.Series)),
		CollectComment: comment,
		CollectRate:    int(rate),
	})
	if err != nil {
		return errgo.Wrap(err, "marshal")
	}

	err = m.q.TimeLine.WithContext(ctx).Create(
		&dao.TimeLine{
			UID:      u.ID,
			Cat:      model.TimeLineCatSubject,
			Type:     tlType,
			Related:  strconv.FormatUint(uint64(sbj.ID), 10),
			Memo:     serializedMemo,
			Img:      img,
			Dateline: uint32(time.Now().Unix()),
		})

	return errgo.Wrap(err, "dal.create")
}

func (m mysqlRepo) updatePreviousSubjectCollection(
	ctx context.Context,
	p *dao.TimeLine,
	sbj model.Subject,
) error {
	serializedMemo, imgBytes, err := mergePreviousTimeline(p, sbj)
	if err != nil {
		return err
	}

	_, err = m.q.TimeLine.WithContext(ctx).Where(m.q.TimeLine.ID.Eq(p.ID)).UpdateSimple(
		m.q.TimeLine.Memo.Value(serializedMemo),
		m.q.TimeLine.Img.Value(imgBytes),
		m.q.TimeLine.Batch.Value(1),
	)

	return errgo.Wrap(err, "dal.create")
}

func mergePreviousTimeline(p *dao.TimeLine, sbj model.Subject) ([]byte, []byte, error) {
	var err error
	var batch = p.Batch != 0
	var batchImage = make(map[string]image.Subject, 2)
	var batchMemo = make(map[string]memo.SubjectMemo, 2)

	if batch {
		var img image.Subject
		err = phpserialize.Unmarshal(p.Img, &img)
		batchImage[strconv.FormatUint(uint64(img.SubjectID), 10)] = img
	} else {
		err = phpserialize.Unmarshal(p.Img, &batchImage)
	}
	if err != nil {
		return nil, nil, errgo.Wrap(err, "php.unmarshal")
	}

	batchImage[strconv.FormatUint(uint64(sbj.ID), 10)] = image.Subject{SubjectID: sbj.ID, Images: sbj.Image}

	if batch {
		var mm memo.SubjectMemo
		err = phpserialize.Unmarshal(p.Memo, &mm)
		batchMemo[mm.ID] = mm
	} else {
		err = phpserialize.Unmarshal(p.Memo, &batchMemo)
	}

	if err != nil {
		return nil, nil, errgo.Wrap(err, "php.unmarshal")
	}

	batchMemo[strconv.FormatUint(uint64(sbj.ID), 10)] = memo.SubjectMemo{
		ID:     strconv.FormatUint(uint64(sbj.ID), 10),
		TypeID: strconv.FormatUint(uint64(sbj.TypeID), 10),
		Name:   sbj.Name,
		NameCN: sbj.NameCN,
		Series: strconv.Itoa(generic.BtoI(sbj.Series)),
	}

	imgBytes, err := phpserialize.Marshal(batchImage)
	if err != nil {
		return nil, nil, errgo.Wrap(err, "marshal")
	}

	serializedMemo, err := phpserialize.Marshal(batchMemo)
	if err != nil {
		return nil, nil, errgo.Wrap(err, "marshal")
	}

	return serializedMemo, imgBytes, nil
}

func convSubjectType(collection model.SubjectCollection, st model.SubjectType) uint16 {
	original := collection

	l, ok := subjectTypeMap()[st]

	if !ok {
		return uint16(original)
	}

	if original < 1 || original > 5 {
		return uint16(original)
	}

	return l[original]
}

// 想看 看过 在看 搁置 抛弃.
func subjectTypeMap() map[model.SubjectType][]uint16 {
	return map[model.SubjectType][]uint16{
		model.SubjectTypeBook:  {0, 1, 5, 9, 13, 14},
		model.SubjectTypeAnime: {0, 2, 6, 10, 13, 14},
		model.SubjectTypeMusic: {0, 3, 7, 11, 13, 14},
		model.SubjectTypeGame:  {0, 4, 8, 12, 13, 14},
		model.SubjectTypeReal:  {0, 2, 6, 10, 13, 14},
	}
}
