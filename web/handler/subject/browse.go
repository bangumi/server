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

package subject

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) Browse(c echo.Context) error {
	// u := accessor.GetFromCtx(c)
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res.Paged{Data: []res.SubjectV0{}, Total: 0, Limit: page.Limit, Offset: page.Offset})
}

func parseBrowseQuery(c echo.Context) (query req.BrowseSubjects, err error) {
	query = req.BrowseSubjects{}

	query.SubjectType, err = req.ParseSubjectType(c.QueryParam("type"))
	if err != nil {
		err = res.BadRequest(err.Error())
		return
	}

	return
}

// - name: cat
//   in: query
//   description: 条目分类，参照 `SubjectCategory` enum
//   required: false
//   schema:
// 	$ref: "#/components/schemas/SubjectCategory"
// - name: series
//   in: query
//   description: 是否系列，仅对书籍类型的条目有效
//   required: false
//   schema:
// 	type: boolean
// - name: platform
//   in: query
//   description: 平台，仅对游戏类型的条目有效
//   required: false
//   schema:
// 	type: string
// - name: order
//   in: query
//   description: 排序，枚举值 {date|rank}
//   required: false
//   schema:
// 	title: Sort Order
// 	type: string
// - name: year
//   in: query
//   description: 年份
//   required: false
//   schema:
// 	type: integer
// - name: month
//   in: query
//   description: 月份
//   required: false
//   schema:
// 	type: integer
