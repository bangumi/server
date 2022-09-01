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

type RelationMemo struct {
	UserID   string `php:"uid"`
	Username string `php:"username"`
	Nickname string `php:"nickname"`
}

func (m *RelationMemo) ToModel() *model.TimeLineMemoContent {
	result := &model.TimeLineRelationMemo{}
	util.CopySameNameField(result, m)
	return &model.TimeLineMemoContent{TimeLineRelationMemo: result}
}

func (m *RelationMemo) FromModel(mc *model.TimeLineMemoContent) {
	util.CopySameNameField(m, mc.TimeLineRelationMemo)
}
