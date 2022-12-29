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

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

var ErrUnexpectedType = fmt.Errorf("unexpected type")

// phpSerializedMemo a glue layer between db repo data type and model
// memo is php serialized and stored as []byte in db(tml_memo column).
type phpSerializedMemo interface {
	// ToModel converts this glue to model,
	// usually following php.unmarshal() that unmarshal the repo data into this glue
	ToModel() *model.TimeLineMemoContent

	// FromModel converts the model to glue,
	// usually followed by php.marshal() that turns this glue into []byte repo format
	FromModel(mc *model.TimeLineMemoContent)
}

func unmarshal(tl *dao.TimeLine, psm phpSerializedMemo) (*model.TimeLineMemoContent, error) {
	if err := phpserialize.Unmarshal(tl.Memo, psm); err != nil {
		return nil, errgo.Wrap(err, "phpserialize.Unmarshal")
	}
	return psm.ToModel(), nil
}

func marshal(tl *model.TimeLineMemoContent, psm phpSerializedMemo) ([]byte, error) {
	psm.FromModel(tl)
	result, err := phpserialize.Marshal(psm)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
}

func DAOToModel(tl *dao.TimeLine) (*model.TimeLineMemo, error) {
	content, err := ContentDAOToModel(tl)
	if err != nil {
		return nil, errgo.Wrap(err, "ContentDAOToModel")
	}
	return &model.TimeLineMemo{
		Cat:     tl.Cat,
		Type:    tl.Type,
		Content: content,
	}, nil
}

//nolint:gomnd,gocyclo
func ContentDAOToModel(tl *dao.TimeLine) (*model.TimeLineMemoContent, error) {
	var (
		result *model.TimeLineMemoContent
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
			result = &model.TimeLineMemoContent{
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
func ModelToDAO(memo *model.TimeLineMemo) ([]byte, error) {
	var (
		result []byte
		err    error
	)
	switch {
	case memo.Cat == model.TimeLineCatRelation && memo.Type == 2: // relation
		result, err = marshal(memo.Content, &RelationMemo{})
	case memo.Cat == model.TimeLineCatGroup && (memo.Type == 3 || memo.Type == 4): // group
		result, err = marshal(memo.Content, &GroupMemo{})
	case memo.Cat == model.TimeLineCatWiki: // wiki
		result, err = marshal(memo.Content, &WikiMemo{})
	case memo.Cat == model.TimeLineCatSubject: // Subject
		result, err = marshal(memo.Content, &SubjectMemo{})
	case memo.Cat == model.TimeLineCatProgress: // progress
		result, err = marshal(memo.Content, &ProgressMemo{})
	case memo.Cat == model.TimeLineCatSay: // say
		if memo.Type == 2 {
			result, err = marshal(memo.Content, &SayEditMemo{})
		} else {
			// SayMemo is not php serialized, thus convert it directly
			result = ([]byte)(*memo.Content.TimeLineSayMemo.TimeLineSay)
		}
	case memo.Cat == model.TimeLineCatBlog: // blog
		result, err = marshal(memo.Content, &BlogMemo{})
	case memo.Cat == model.TimeLineCatIndex: // index
		result, err = marshal(memo.Content, &IndexMemo{})
	case memo.Cat == model.TimeLineCatMono: // mono
		result, err = marshal(memo.Content, &MonoMemo{})
	case memo.Cat == model.TimeLineCatDoujin: // doujin
		result, err = marshal(memo.Content, &DoujinMemo{})
	default:
		err = ErrUnexpectedType
	}
	if err != nil {
		return nil, fmt.Errorf("marshal(cat: %v, type: %v): %w", memo.Cat, memo.Type, err)
	}
	return result, nil
}
