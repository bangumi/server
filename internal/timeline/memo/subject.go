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

type SubjectMemo struct {
	ID             string `php:"subject_id"`
	TypeID         string `php:"subject_type_id"`
	Name           string `php:"subject_name"`
	NameCN         string `php:"subject_name_cn"`
	Series         string `php:"subject_series"`
	CollectComment string `php:"collect_comment"`
	CollectRate    int    `php:"collect_rate"`
}

func (m *SubjectMemo) ToModel() *model.TimeLineMemo {
	result := &model.TimeLineSubjectMemo{}
	util.CopySameNameField(result, m)
	return &model.TimeLineMemo{TimeLineSubjectMemo: result}
}

func (m *SubjectMemo) FromModel(tl *model.TimeLine) {
	util.CopySameNameField(m, tl.Memo.TimeLineSubjectMemo)
}
