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

type GroupMemo struct {
	ID          string `php:"grp_id"`
	Name        string `php:"grp_name"`
	Title       string `php:"grp_title"`
	Description string `php:"grp_desc"`
}

func (g *GroupMemo) ToModel() *model.TimeLineMemo {
	result := &model.TimeLineGroupMemo{}
	util.CopySameNameField(result, g)
	return &model.TimeLineMemo{TimeLineGroupMemo: result}
}

func (g *GroupMemo) FromModel(tl *model.TimeLine) {
	util.CopySameNameField(g, tl.Memo.TimeLineGroupMemo)
}
