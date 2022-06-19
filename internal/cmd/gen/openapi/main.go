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

package main

import (
	"os"
	"reflect"

	"github.com/danielgtaylor/huma/schema"
	"github.com/ghodss/yaml"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strutil"
	"github.com/bangumi/server/internal/web/req"
)

func main() {
	s, err := schema.Generate(reflect.TypeOf(req.PutSubjectCollection{}))
	if err != nil {
		logger.Fatal("failed to generate schema", zap.Error(err))
	}

	b, err := yaml.Marshal(WithExampleValue{
		Schema: s,
		Example: req.PutSubjectCollection{
			Comment:   "看看",
			Tags:      strutil.Split("柯南 万年小学生 推理 青山刚昌 TV", " "),
			EpStatus:  1041,
			VolStatus: 0,
			Type:      model.CollectionDoing,
			Rate:      8,
			Private:   false,
		},
	})
	if err != nil {
		logger.Fatal("failed to marshal yaml", zap.Error(err))
	}

	if err = os.WriteFile("./openapi/components/put_subject_collection.yaml", b, 0600); err != nil { //nolint:gomnd
		logger.Fatal("failed to write file", zap.Error(err))
	}
}

type WithExampleValue struct {
	Example interface{} `json:"example"`
	*schema.Schema
}
