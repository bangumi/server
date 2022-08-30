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
	"fmt"

	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

var ErrUnexpectedType = fmt.Errorf("unexpected type")

// phpSerializedMemo a glue layer between db repo data type and model
// memo is php serialized and stored as []byte in db(tml_memo column).
type phpSerializedMemo interface {
	// ToModel converts this glue to model,
	// usually following php.unmarshal() that unmarshal the repo data into this glue
	ToModel() *model.TimeLineMemo

	// FromModel converts the model to glue,
	// usually followed by php.marshal() that turns this glue into []byte repo format
	FromModel(tl *model.TimeLine)
}

func unmarshal(tl *dao.TimeLine, memo phpSerializedMemo) (*model.TimeLineMemo, error) {
	if err := phpserialize.Unmarshal(tl.Memo, memo); err != nil {
		return nil, errgo.Wrap(err, "phpserialize.Unmarshal")
	}
	return memo.ToModel(), nil
}

func marshal(tl *model.TimeLine, memo phpSerializedMemo) ([]byte, error) {
	memo.FromModel(tl)
	result, err := phpserialize.Marshal(memo)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
}

//nolint:gomnd,gocyclo
func DAOToModel(tl *dao.TimeLine) (*model.TimeLineMemo, error) {
	var (
		result *model.TimeLineMemo
		err    error
	)
	switch {
	case tl.Cat == model.TimeLineCatRelation && tl.Type == 2: // relation
		result, err = unmarshal(tl, &RelationMemo{})
	case tl.Cat == model.TimeLineCatGroup && (tl.Type == 3 || tl.Type == 4): // group
		result, err = unmarshal(tl, &GroupMemo{})
	case tl.Cat == model.TimeLineCatWiki: // wiki
		result, err = unmarshal(tl, &WikiMemo{})
	case tl.Cat == model.TimeLineCatSubject: // Subject
		result, err = unmarshal(tl, &SubjectMemo{})
	case tl.Cat == model.TimeLineCatProgress: // progress
		result, err = unmarshal(tl, &ProgressMemo{})
	case tl.Cat == model.TimeLineCatSay: // say
		if tl.Type == 2 {
			result, err = unmarshal(tl, &SayEditMemo{})
		} else {
			// SayMemo is not php serialized, thus convert it directly
			sayMemoString := string(tl.Memo)
			result = &model.TimeLineMemo{
				TimeLineSayMemo: &model.TimeLineSayMemo{TimeLineSay: (*model.TimeLineSay)(&sayMemoString)},
			}
		}
	case tl.Cat == model.TimeLineCatBlog: // blog
		result, err = unmarshal(tl, &BlogMemo{})
	case tl.Cat == model.TimeLineCatIndex: // index
		result, err = unmarshal(tl, &IndexMemo{})
	case tl.Cat == model.TimeLineCatMono: // mono
		result, err = unmarshal(tl, &MonoMemo{})
	case tl.Cat == model.TimeLineCatDoujin: // doujin
		result, err = unmarshal(tl, &DoujinMemo{})
	default:
		err = ErrUnexpectedType
	}
	if err != nil {
		return result, errgo.Wrap(err, fmt.Sprintf("unmarshal(cat: %v, type: %v)", tl.Cat, tl.Type))
	}
	return result, nil
}

//nolint:gomnd,gocyclo
func ModelToDAO(tl *model.TimeLine) ([]byte, error) {
	var (
		result []byte
		err    error
	)
	switch {
	case tl.Cat == model.TimeLineCatRelation && tl.Type == 2: // relation
		result, err = marshal(tl, &RelationMemo{})
	case tl.Cat == model.TimeLineCatGroup && (tl.Type == 3 || tl.Type == 4): // group
		result, err = marshal(tl, &GroupMemo{})
	case tl.Cat == model.TimeLineCatWiki: // wiki
		result, err = marshal(tl, &WikiMemo{})
	case tl.Cat == model.TimeLineCatSubject: // Subject
		result, err = marshal(tl, &SubjectMemo{})
	case tl.Cat == model.TimeLineCatProgress: // progress
		result, err = marshal(tl, &ProgressMemo{})
	case tl.Cat == model.TimeLineCatSay: // say
		if tl.Type == 2 {
			result, err = marshal(tl, &SayEditMemo{})
		} else {
			// SayMemo is not php serialized, thus convert it directly
			result = ([]byte)(*tl.Memo.TimeLineSayMemo.TimeLineSay)
		}
	case tl.Cat == model.TimeLineCatBlog: // blog
		result, err = marshal(tl, &BlogMemo{})
	case tl.Cat == model.TimeLineCatIndex: // index
		result, err = marshal(tl, &IndexMemo{})
	case tl.Cat == model.TimeLineCatMono: // mono
		result, err = marshal(tl, &MonoMemo{})
	case tl.Cat == model.TimeLineCatDoujin: // doujin
		result, err = marshal(tl, &DoujinMemo{})
	default:
		err = ErrUnexpectedType
	}
	if err != nil {
		return nil, fmt.Errorf("marshal(cat: %v, type: %v): %w", tl.Cat, tl.Type, err)
	}
	return result, nil
}
