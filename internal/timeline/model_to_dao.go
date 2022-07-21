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
	"fmt"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func convertModelToDAO(tl *model.TimeLine) (*dao.TimeLine, error) {
	img, err := marshalImages(tl)
	if err != nil {
		return nil, errgo.Wrap(err, "modelImageToDAO")
	}
	memo, err := memoToDAO(tl)
	if err != nil {
		return nil, errgo.Wrap(err, "memoToDAO")
	}

	return &dao.TimeLine{
		ID:       tl.ID,
		Related:  tl.Related,
		Img:      img,
		Memo:     memo,
		UID:      tl.UID,
		Replies:  tl.Replies,
		Dateline: tl.Dateline,
		Cat:      tl.Cat,
		Type:     tl.Type,
		Batch:    tl.Batch,
		Source:   tl.Source,
	}, nil
}

//nolint:gomnd
func memoToDAO(tl *model.TimeLine) ([]byte, error) {
	switch {
	case tl.Cat == 4: // Progress
		result, err := marshalProgressMemo(tl)
		return result, errgo.Wrap(err, "marshalProgressMemo")
	case tl.Cat == 5: // Say
		result, err := marshalSayMemo(tl)
		return result, errgo.Wrap(err, "sayMemoToBytes")
	default:
		return nil, errgo.Wrap(ErrUnexpectedType, fmt.Sprintf("(cat: %v, type: %v)", tl.Cat, tl.Type))
	}
}
