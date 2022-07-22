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

package image

import (
	"fmt"

	"github.com/bangumi/server/internal/pkg/util"
	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

type Image struct {
	Cat       *int64  `php:"cat,omitempty"`
	GroupID   *string `php:"grp_id,omitempty"`
	GroupName *string `php:"grp_name,omitempty"`
	Name      *string `php:"name,omitempty"`
	Title     *string `php:"title,omitempty"`
	ID        *string `php:"id,omitempty"`
	UserID    *string `php:"uid,omitempty"`
	SubjectID *string `php:"subject_id,omitempty"`
	Images    *string `php:"images,omitempty"`
}

func (i *Image) ToModel() *model.TimeLineImage {
	result := &model.TimeLineImage{}
	util.CopySameNameField(result, i)
	return result
}

func (i *Image) FromModel(image *model.TimeLineImage) {
	util.CopySameNameField(i, image)
}

func marshalImage(image *model.TimeLineImage) ([]byte, error) {
	var i Image
	i.FromModel(image)
	result, err := phpserialize.Marshal(i)
	return result, errgo.Wrap(err, "phpserialize.marshal")
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

func DAOToModel(b []byte) (model.TimeLineImages, error) {
	// the image []byte in db is in the type: Union[Image, Images]
	// handle the single-Image case first, fast path
	var image Image
	var errImage error
	if errImage = phpserialize.Unmarshal(b, &image); errImage == nil {
		return model.TimeLineImages{*image.ToModel()}, nil
	}

	// unmarshal Image errors, try the Images unmarshalling path
	// wraps error raised in Image handling path
	var images Images
	if err := phpserialize.Unmarshal(b, &images); err != nil {
		return nil, errgo.Wrap(err,
			fmt.Sprintf("phpserialize.unmarshal(with unmarshal as Image failed too with error: %v)", errImage),
		)
	}
	return images.ToModel(), nil
}

func ModelToDAO(tl *model.TimeLine) ([]byte, error) {
	images := tl.Image

	// handle the single-Image case first, fast path
	if len(images) == 1 {
		result, err := marshalImage(&images[0])
		return result, errgo.Wrap(err, "marshalImage")
	}

	var is Images
	for i := range images {
		var img Image
		img.FromModel(&images[i])
		is = append(is, img)
	}
	result, err := phpserialize.Marshal(is)
	return result, errgo.Wrap(err, "phpserialize.marshal")
}
