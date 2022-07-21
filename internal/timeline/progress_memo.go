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
	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

type ProgressMemo struct {
	EpName        *string            `php:"ep_name,omitempty"`
	VolsTotal     *string            `php:"vols_total,omitempty"`
	SubjectName   *string            `php:"subject_name,omitempty"`
	EpsUpdate     *int               `php:"eps_update,omitempty,string"`
	VolsUpdate    *int               `php:"vols_update,omitempty,string"`
	EpsTotal      *int               `php:"eps_total,omitempty,string"`
	EpSort        *int               `php:"ep_sort,omitempty,string"`
	EpID          *model.EpisodeID   `php:"ep_id,omitempty,string"`
	SubjectID     *model.SubjectID   `php:"subject_id,omitempty,string"`
	SubjectTypeID *model.SubjectType `php:"subject_type_id,omitempty,string"`
}

func (m *ProgressMemo) ToModel() *model.TimeLineProgressMemo {
	if m == nil {
		return nil
	}
	return &model.TimeLineProgressMemo{
		EpName:        m.EpName,
		VolsTotal:     m.VolsTotal,
		SubjectName:   m.SubjectName,
		EpsUpdate:     m.EpsUpdate,
		VolsUpdate:    m.VolsUpdate,
		EpsTotal:      m.EpsTotal,
		EpSort:        m.EpSort,
		EpID:          m.EpID,
		SubjectID:     m.SubjectID,
		SubjectTypeID: m.SubjectTypeID,
	}
}

func (m *ProgressMemo) FromModel(memo *model.TimeLineProgressMemo) {
	if memo == nil {
		return
	}
	m.EpName = memo.EpName
	m.VolsTotal = memo.VolsTotal
	m.SubjectName = memo.SubjectName
	m.EpsUpdate = memo.EpsUpdate
	m.VolsUpdate = memo.VolsUpdate
	m.EpsTotal = memo.EpsTotal
	m.EpSort = memo.EpSort
	m.EpID = memo.EpID
	m.SubjectID = memo.SubjectID
	m.SubjectTypeID = memo.SubjectTypeID
}

func unpackProgressMemo(tl *dao.TimeLine) (model.TimeLineContent, error) {
	var m ProgressMemo
	if err := phpserialize.Unmarshal(tl.Memo, &m); err != nil {
		return model.TimeLineContent{}, errgo.Wrap(err, "phpserialize.Unmarshal")
	}
	return model.TimeLineContent{TimeLineProgressMemo: m.ToModel()}, nil
}

func marshalProgressMemo(tl *model.TimeLine) ([]byte, error) {
	var pm ProgressMemo
	pm.FromModel(tl.Memo.TimeLineProgressMemo)
	result, err := phpserialize.Marshal(pm)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
}
