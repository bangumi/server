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

import "reflect"

type PersonImages struct {
	Small  string `json:"small"`
	Grid   string `json:"grid"`
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

type SubjectImages struct {
	Small  string `json:"small"`
	Grid   string `json:"grid"`
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Common string `json:"common"`
}

func SubjectImage(s string) SubjectImages {
	if s == "" {
		return SubjectImages{}
	}

	return SubjectImages{
		Grid:   "https://lain.bgm.tv/pic/cover/g/" + s,
		Small:  "https://lain.bgm.tv/pic/cover/s/" + s,
		Common: "https://lain.bgm.tv/pic/cover/c/" + s,
		Large:  "https://lain.bgm.tv/pic/cover/l/" + s,
		Medium: "https://lain.bgm.tv/pic/cover/m/" + s,
	}
}

func PersonImage(s string) PersonImages {
	if s == "" {
		return PersonImages{}
	}

	return PersonImages{
		Grid:   "https://lain.bgm.tv/pic/crt/g/" + s,
		Small:  "https://lain.bgm.tv/pic/crt/s/" + s,
		Large:  "https://lain.bgm.tv/pic/crt/l/" + s,
		Medium: "https://lain.bgm.tv/pic/crt/m/" + s,
	}
}

const (
	DefaultImageURL = "https://lain.bgm.tv/img/no_icon_subject.png"
)

func SelectImageByType(in interface{}, key string) (string, bool) {
	if in == nil {
		return "", false
	}
	t := reflect.TypeOf(in)
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get("json") == key {
			return reflect.ValueOf(in).Field(i).String(), true
		}
	}
	return "", false
}
