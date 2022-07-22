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

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/trim21/go-phpserialize"
)

var ErrUnexpectedType = fmt.Errorf("unexpected type")

// phpSerializedMemo a glue layer between db repo data type and model
// memo is php serialized and stored as []byte in db(tml_memo column)
type phpSerializedMemo interface {
	// ToModel converts this glue to model,
	// usually following php.unmarshal() that unmarshal the repo data into this glue
	ToModel() *model.TimeLineMemo

	// FromModel converts the model to glue,
	// usually followed by php.marshal() that turns this glue into []byte repo format
	FromModel(tl *model.TimeLine)
}

func unmarshal[typ phpSerializedMemo](tl *dao.TimeLine, memo typ) (*model.TimeLineMemo, error) {
	if err := phpserialize.Unmarshal(tl.Memo, &memo); err != nil {
		return nil, errgo.Wrap(err, "phpserialize.Unmarshal")
	}
	return memo.ToModel(), nil
}

func marshal[typ phpSerializedMemo](tl *model.TimeLine, memo typ) ([]byte, error) {
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
	case tl.Cat == 1 && tl.Type == 2: // relation
		result, err = unmarshal(tl, &RelationMemo{})
	case tl.Cat == 1 && (tl.Type == 3 || tl.Type == 4): // group
		result, err = unmarshal(tl, &GroupMemo{})
	case tl.Cat == 2: // wiki
		result, err = unmarshal(tl, &WikiMemo{})
	case tl.Cat == 3: // Subject
		result, err = unmarshal(tl, &SubjectMemo{})
	case tl.Cat == 4: // progress
		result, err = unmarshal(tl, &ProgressMemo{})
	case tl.Cat == 5: // say
		if tl.Type == 2 {
			result, err = unmarshal(tl, &SayEditMemo{})
		} else {
			// SayMemo is not php serialized, thus convert it directly
			sayMemoString := string(tl.Memo)
			result.TimeLineSayMemo.TimeLineSay = (*model.TimeLineSay)(&sayMemoString)
		}
	case tl.Cat == 6: // blog
		result, err = unmarshal(tl, &BlogMemo{})
	case tl.Cat == 7: // index
		result, err = unmarshal(tl, &IndexMemo{})
	case tl.Cat == 8: // mono
		result, err = unmarshal(tl, &MonoMemo{})
	case tl.Cat == 9: // doujin
		result, err = unmarshal(tl, &DoujinMemo{})
	default:
		err = ErrUnexpectedType
	}
	return result, errgo.Wrap(err, fmt.Sprintf("(cat: %v, type: %v)", tl.Cat, tl.Type))
}

//nolint:gomnd,gocyclo
func ModelToDAO(tl *model.TimeLine) ([]byte, error) {
	var (
		result []byte
		err    error
	)
	switch {
	case tl.Cat == 1 && tl.Type == 2: // relation
		result, err = marshal(tl, &RelationMemo{})
	case tl.Cat == 1 && (tl.Type == 3 || tl.Type == 4): // group
		result, err = marshal(tl, &GroupMemo{})
	case tl.Cat == 2: // wiki
		result, err = marshal(tl, &WikiMemo{})
	case tl.Cat == 3: // Subject
		result, err = marshal(tl, &SubjectMemo{})
	case tl.Cat == 4: // progress
		result, err = marshal(tl, &ProgressMemo{})
	case tl.Cat == 5: // say
		if tl.Type == 2 {
			result, err = marshal(tl, &SayEditMemo{})
		} else {
			// SayMemo is not php serialized, thus convert it directly
			result = ([]byte)(*tl.Memo.TimeLineSayMemo.TimeLineSay)
		}
	case tl.Cat == 6: // blog
		result, err = marshal(tl, &BlogMemo{})
	case tl.Cat == 7: // index
		result, err = marshal(tl, &IndexMemo{})
	case tl.Cat == 8: // mono
		result, err = marshal(tl, &MonoMemo{})
	case tl.Cat == 9: // doujin
		result, err = marshal(tl, &DoujinMemo{})
	default:
		err = ErrUnexpectedType
	}
	return result, errgo.Wrap(err, fmt.Sprintf("(cat: %v, type: %v)", tl.Cat, tl.Type))
}
