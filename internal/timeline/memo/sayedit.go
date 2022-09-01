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

type SayEditMemo struct {
	Before string `php:"before"`
	After  string `php:"after"`
}

func (m *SayEditMemo) ToModel() *model.TimeLineMemoContent {
	result := &model.TimeLineSayEdit{}
	util.CopySameNameField(result, m)
	return &model.TimeLineMemoContent{TimeLineSayMemo: &model.TimeLineSayMemo{TimeLineSayEdit: result}}
}

func (m *SayEditMemo) FromModel(mc *model.TimeLineMemoContent) {
	util.CopySameNameField(m, mc.TimeLineSayMemo.TimeLineSayEdit)
}
