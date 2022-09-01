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

package image

import (
	"fmt"
	"strconv"

	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/util"
)

type Image struct {
	Cat       *int64  `php:"cat,omitempty"`
	GroupID   *string `php:"grp_id,omitempty"`
	GroupName *string `php:"grp_name,omitempty"`
	Name      *string `php:"name,omitempty"`
	Title     *string `php:"title,omitempty"`
	ID        *int    `php:"id,omitempty"`
	UserID    *string `php:"uid,omitempty"`
	SubjectID *string `php:"subject_id,omitempty"`
	Images    *string `php:"images,omitempty"`
}

func (i *Image) ToModel() *model.TimeLineImage {
	result := &model.TimeLineImage{}
	util.CopySameNameField(result, i)
	return result
}

func (i *Image) UnmarshalPHP(b []byte) error {
	m := make(map[string]any)
	if err := phpserialize.Unmarshal(b, &m); err != nil {
		return errgo.Wrap(err, "phpserialize.Unmarshal")
	}

	i.Cat = extractAs[int64](m, "cat")
	i.GroupID = extractAs[string](m, "grp_id")
	i.GroupName = extractAs[string](m, "grp_name")
	i.Name = extractAs[string](m, "name")
	i.Title = extractAs[string](m, "title")
	i.UserID = extractAs[string](m, "uid")
	i.SubjectID = extractAs[string](m, "subject_id")
	i.Images = extractAs[string](m, "images")

	i.ID = extractAs[int](m, "id")
	if i.ID == nil {
		// for some cases, the ID is stored as string in db
		// try to extractAs string if failed as int
		i.ID = strPtrToIntPtr(extractAs[string](m, "id"))
	}
	return nil
}

func extractAs[T interface{ int64 | string | int }](m map[string]any, key string) *T {
	d, ok := m[key]
	if !ok {
		return nil
	}
	d2, ok := d.(T)
	if !ok {
		return nil
	}
	return &d2
}

func strPtrToIntPtr(s *string) *int {
	if s == nil {
		return nil
	}
	result, err := strconv.Atoi(*s)
	if err != nil {
		return nil
	}
	return &result
}

func intPtrToStrPtr(i *int) *string {
	if i == nil {
		return nil
	}
	result := strconv.Itoa(*i)
	return &result
}

func (i *Image) FromModel(image *model.TimeLineImage) {
	util.CopySameNameField(i, image)
}

//nolint:gocyclo
func imageToMap(image *model.TimeLineImage, cat uint16) map[string]any {
	var i Image
	i.FromModel(image)

	m := make(map[string]any)
	if i.Cat != nil {
		m["cat"] = i.Cat
	}
	if i.GroupID != nil {
		m["grp_id"] = i.GroupID
	}
	if i.GroupName != nil {
		m["grp_name"] = i.GroupName
	}
	if i.Name != nil {
		m["name"] = i.Name
	}
	if i.Title != nil {
		m["title"] = i.Title
	}
	if i.UserID != nil {
		m["uid"] = i.UserID
	}
	if i.SubjectID != nil {
		m["subject_id"] = i.SubjectID
	}
	if i.Images != nil {
		m["images"] = i.Images
	}

	if i.ID != nil {
		if cat == 9 {
			// yes, the cat 9 is the case that image id is stored as string
			m["id"] = intPtrToStrPtr(i.ID)
		} else {
			m["id"] = i.ID
		}
	}
	return m
}

type Images []Image

func (is Images) ToModel() model.TimeLineImages {
	if is == nil {
		return nil
	}
	var images model.TimeLineImages
	for _, i := range is {
		images = append(images, *i.ToModel())
	}
	return images
}

func DAOToModel(tl *dao.TimeLine) (model.TimeLineImages, error) {
	b := tl.Img

	// the image []byte in db is in the type: Union[Image, Images]
	// handle the single-Image case first, fast path
	var resultFP Image // result fast path
	errFP := phpserialize.Unmarshal(b, &resultFP)
	if errFP == nil {
		return []model.TimeLineImage{*resultFP.ToModel()}, nil
	}

	// unmarshal Image errors, try the Images unmarshalling path
	// wraps error raised in single-Image handling path
	var images Images
	if err := phpserialize.Unmarshal(b, &images); err != nil {
		return nil, errgo.Wrap(err,
			fmt.Sprintf("phpserialize.unmarshal(with unmarshal as single-Image failed too with error: %v)", errFP),
		)
	}
	return images.ToModel(), nil
}

func ModelToDAO(tl *model.TimeLine) ([]byte, error) {
	images := tl.Image

	// handle the single-Image case first, fast path
	if len(images) == 1 {
		result, err := phpserialize.Marshal(imageToMap(&images[0], tl.Cat))
		return result, errgo.Wrap(err, "phpserialize.Marshal")
	}

	var is []map[string]any
	for i := range images {
		is = append(is, imageToMap(&images[i], tl.Cat))
	}
	result, err := phpserialize.Marshal(is)
	return result, errgo.Wrap(err, "phpserialize.marshal")
}
