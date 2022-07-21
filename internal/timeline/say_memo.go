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

type SayMemo string

func (m *SayMemo) ToModel() *model.TimeLineSay {
	return (*model.TimeLineSay)(m)
}

// SayMemo2 cat=5, type=2
// TODO: looking for a better name kk.
type SayMemo2 struct {
	Before string `php:"before"`
	After  string `php:"after"`
}

func (m *SayMemo2) ToModel() *model.TimeLineSayEdit {
	if m == nil {
		return nil
	}
	return &model.TimeLineSayEdit{
		Before: m.Before,
		After:  m.After,
	}
}

func (m *SayMemo2) FromModel(memo *model.TimeLineSayEdit) {
	if memo == nil {
		return
	}
	m.Before = memo.Before
	m.After = memo.After
}

func unpackSayMemo(tl *dao.TimeLine) (model.TimeLineContent, error) {
	var say model.TimeLineSayMemo
	if tl.Type == 2 {
		var m SayMemo2
		if err := phpserialize.Unmarshal(tl.Memo, &m); err != nil {
			return model.TimeLineContent{}, errgo.Wrap(err, "phpserialize.Unmarshal")
		}
		say.TimeLineSayEdit = m.ToModel()
	} else {
		var m SayMemo
		if err := phpserialize.Unmarshal(tl.Memo, &m); err != nil {
			return model.TimeLineContent{}, errgo.Wrap(err, "phpserialize.Unmarshal")
		}
		say.TimeLineSay = m.ToModel()
	}
	return model.TimeLineContent{TimeLineSayMemo: &say}, nil
}

func marshalSayMemo(tl *model.TimeLine) ([]byte, error) {
	memo := tl.Memo.TimeLineSayMemo
	if memo == nil {
		return nil, ErrEmptyMemo
	}

	if tl.Type == 2 {
		var m SayMemo2
		m.FromModel(memo.TimeLineSayEdit)
		result, err := phpserialize.Marshal(m)
		return result, errgo.Wrap(err, "phpserialize.Marshal")
	}
	return []byte(*memo.TimeLineSay), nil
}
