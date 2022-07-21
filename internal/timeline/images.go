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

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/trim21/go-phpserialize"
)

type Image struct {
	Cat       *int64  `json:"cat,omitempty" php:"cat,omitempty"`
	GroupID   *string `json:"grp_id,omitempty" php:"grp_id,omitempty"`
	GroupName *string `json:"grp_name,omitempty" php:"grp_name,omitempty"`
	Name      *string `json:"name,omitempty" php:"name,omitempty"`
	Title     *string `json:"title,omitempty" php:"title,omitempty"`
	ID        *string `json:"id,omitempty" php:"id,omitempty"`
	UserID    *string `json:"uid,omitempty" php:"uid,omitempty"`
	SubjectID *string `json:"subject_id,omitempty" php:"subject_id,omitempty"`
	Images    *string `json:"images,omitempty" php:"images,omitempty"`
}

func (i *Image) ToModel() *model.TimeLineImage {
	if i == nil {
		return nil
	}
	return &model.TimeLineImage{
		Cat:       i.Cat,
		GroupID:   i.GroupID,
		GroupName: i.GroupName,
		Name:      i.Name,
		Title:     i.Title,
		ID:        i.ID,
		UserID:    i.UserID,
		SubjectID: i.SubjectID,
		Images:    i.Images,
	}
}

func (i *Image) FromModel(image *model.TimeLineImage) {
	if image == nil {
		return
	}
	i.Cat = image.Cat
	i.GroupID = image.GroupID
	i.GroupName = image.GroupName
	i.Name = image.Name
	i.Title = image.Title
	i.ID = image.ID
	i.UserID = image.UserID
	i.SubjectID = image.SubjectID
	i.Images = image.Images
}

func marshalImage(image *model.TimeLineImage) ([]byte, error) {
	var i Image
	i.FromModel(image)
	result, err := phpserialize.Marshal(i)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
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

func marshalImages(tl *model.TimeLine) ([]byte, error) {
	images := tl.Image
	if len(images) == 1 {
		return marshalImage(&images[0])
	}
	var is Images
	for i := range images {
		var img Image
		img.FromModel(&images[i])
		is = append(is, img)
	}
	result, err := phpserialize.Marshal(is)
	return result, errgo.Wrap(err, "phpserialize.Marshal")
}

func unpackImages(b []byte) (model.TimeLineImages, error) {
	// the image []byte in db is in the type: Union[Image, Images]
	// handle the Image case first, fast path
	var image Image
	var errImage error
	if errImage = phpserialize.Unmarshal(b, &image); errImage == nil {
		return model.TimeLineImages{*image.ToModel()}, nil
	}

	// Unmarshal Image errors, try the Images unmarshalling path
	// wraps error raised in Image handling path
	var images Images
	if err := phpserialize.Unmarshal(b, &images); err != nil {
		return nil, errgo.Wrap(err,
			fmt.Sprintf("phpserialize.Unmarshal(with unmarshal as Image failed too with error: %v)", errImage),
		)
	}
	return images.ToModel(), nil
}
