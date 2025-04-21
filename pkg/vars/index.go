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

// Package vars provide some pre-defined variable from old codebase.
//
//nolint:gochecknoglobals
package vars

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bangumi/server/internal/model"
	"sigs.k8s.io/yaml"
)

//go:embed common/subject_staffs.yml
var staffRaw []byte

//go:embed platform.go.json
var platformRaw []byte

//go:embed common/subject_relations.yml
var relationRaw []byte

// StaffID ...
type StaffID = uint16

// PlatformID ...
type PlatformID = uint16

// RelationID ...
type RelationID = uint16

var (
	// StaffMap ...
	StaffMap map[model.SubjectType]map[StaffID]Staff
	// PlatformMap ...
	PlatformMap map[model.SubjectType]map[PlatformID]model.Platform
	// RelationMap ...
	RelationMap map[model.SubjectType]map[RelationID]Relation
)

//nolint:gochecknoinits
func init() {
	if err := json.Unmarshal(platformRaw, &PlatformMap); err != nil {
		log.Panicln("can't unmarshal raw staff json to go type", err)
	}
	platformRaw = nil

	var staffsYaml struct {
		Staffs map[model.SubjectType]map[StaffID]Staff `yaml:"staffs"`
	}
	if err := json.Unmarshal(staffRaw, &staffsYaml); err != nil {
		log.Panicln("can't unmarshal raw staff yaml to go type", err)
	}
	staffRaw = nil
	StaffMap = staffsYaml.Staffs

	var relationYAML struct {
		Relations map[model.SubjectType]map[RelationID]Relation `yaml:"relations"`
	}
	if err := yaml.Unmarshal(relationRaw, &relationYAML); err != nil {
		log.Panicln("can't unmarshal raw relation yaml to go type", err)
	}
	relationRaw = nil
	RelationMap = relationYAML.Relations
}

type Staff struct {
	CN  string
	JP  string
	EN  string
	RDF string
}

func (s Staff) String() string {
	switch {
	case s.CN != "":
		return s.CN
	case s.JP != "":
		return s.JP
	case s.EN != "":
		return s.EN
	case s.RDF != "":
		return s.RDF
	default:
		return "unknown"
	}
}

type Relation struct {
	CN          string `json:"cn"`
	EN          string `json:"en"`
	JP          string `json:"jp"`
	Description string `json:"description"`
}

func (r Relation) String(id uint16) string {
	switch {
	case r.CN != "":
		return r.CN
	case r.JP != "":
		return r.JP
	case r.EN != "":
		return r.EN
	default:
		return fmt.Sprintf("unknown(%d)", id)
	}
}
