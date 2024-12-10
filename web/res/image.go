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

package res

const (
	imageSizeLarge  = "large"
	imageSizeMedium = "medium"
	imageSizeSmall  = "small"
	imageSizeGrid   = "grid"
	imageSizeCommon = "common"
)

type PersonImages struct {
	Small  string `json:"small"`
	Grid   string `json:"grid"`
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

func (i PersonImages) Select(t string) (string, bool) {
	switch t {
	case imageSizeSmall:
		return i.Small, true
	case imageSizeGrid:
		return i.Grid, true
	case imageSizeLarge:
		return i.Large, true
	case imageSizeMedium:
		return i.Medium, true
	default:
		return "", false
	}
}

type SubjectImages struct {
	Small  string `json:"small"`
	Grid   string `json:"grid"`
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Common string `json:"common"`
}

func (i SubjectImages) Select(t string) (string, bool) {
	switch t {
	case imageSizeSmall:
		return i.Small, true
	case imageSizeGrid:
		return i.Grid, true
	case imageSizeLarge:
		return i.Large, true
	case imageSizeMedium:
		return i.Medium, true
	case imageSizeCommon:
		return i.Common, true
	default:
		return "", false
	}
}

func SubjectImage(s string) SubjectImages {
	if s == "" {
		return SubjectImages{}
	}

	return SubjectImages{
		Large: "https://lain.bgm.tv/pic/cover/l/" + s,

		Grid:   "https://lain.bgm.tv/r/100/pic/cover/l/" + s,
		Small:  "https://lain.bgm.tv/r/200/pic/cover/l/" + s,
		Common: "https://lain.bgm.tv/r/400/pic/cover/l/" + s,
		Medium: "https://lain.bgm.tv/r/800/pic/cover/l/" + s,
	}
}

func PersonImage(s string) PersonImages {
	if s == "" {
		return PersonImages{}
	}

	return PersonImages{
		Large: "https://lain.bgm.tv/pic/crt/l/" + s,

		Small:  "https://lain.bgm.tv/r/100/pic/crt/l/" + s,
		Grid:   "https://lain.bgm.tv/pic/crt/g/" + s,
		Medium: "https://lain.bgm.tv/r/400/pic/crt/l/" + s,
	}
}

const (
	DefaultImageURL = "https://lain.bgm.tv/img/no_icon_subject.png"
)
