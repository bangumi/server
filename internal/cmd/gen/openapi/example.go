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

// 作为生成 openapi component 的脚本文件示例。
// 不要 commit 你的修改。
package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/schema"
	"github.com/ghodss/yaml"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/web/res"
)

func main() {
	s, err := schema.Generate(reflect.TypeOf(res.PrivateGroupProfile{}))
	if err != nil {
		logger.Fatal("failed to generate schema", zap.Error(err))
	}

	b, err := yaml.Marshal(WithExampleValue{
		Schema: s,
		Example: res.PrivateGroupProfile{
			Name:        "a",
			Title:       "～技术宅真可怕～",
			Description: "本小组欢迎对各种技术有一定了解的人，\n比如像橘花热衷杀人技术……\n\n不过、本组主要谈论ＰＣ软硬件方面，\n想了解相关知识，结识可怕的技术宅，请进。",
			Icon:        "https://lain.bgm.tv/pic/icon/l/000/00/00/11.jpg",
			CreatedAt:   time.Unix(1217542289, 0),
		},
	})
	if err != nil {
		logger.Fatal("failed to marshal yaml", zap.Error(err))
	}

	fmt.Println(string(b)) //nolint:forbidigo
}

type WithExampleValue struct {
	Example any `json:"example"`
	*schema.Schema
}
