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

package memo

import (
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/util"
)

type ProgressMemo struct {
	EpName        *string            `php:"ep_name,omitempty"`
	VolsTotal     *string            `php:"vols_total,omitempty"`
	SubjectName   *string            `php:"subject_name,omitempty"`
	EpsUpdate     *int               `php:"eps_update,omitempty"`
	VolsUpdate    *int               `php:"vols_update,omitempty,string"`
	EpsTotal      *string            `php:"eps_total,omitempty"`
	EpSort        *float32           `php:"ep_sort,omitempty,string"`
	EpID          *model.EpisodeID   `php:"ep_id,omitempty,string"`
	SubjectID     *model.SubjectID   `php:"subject_id,omitempty,string"`
	SubjectTypeID *model.SubjectType `php:"subject_type_id,omitempty,string"`
}

func (m *ProgressMemo) ToModel() *model.TimeLineMemoContent {
	result := &model.TimeLineProgressMemo{}
	util.CopySameNameField(result, m)
	return &model.TimeLineMemoContent{TimeLineProgressMemo: result}
}

func (m *ProgressMemo) FromModel(mc *model.TimeLineMemoContent) {
	util.CopySameNameField(m, mc.TimeLineProgressMemo)
}
