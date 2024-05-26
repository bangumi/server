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
	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) Browse(c echo.Context) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	filter, err := parseBrowseQuery(c)
	if err != nil {
		return err
	}

	count, err := h.subject.Count(c.Request().Context(), filter)
	if err != nil {
		return errgo.Wrap(err, "failed to count subjects")
	}

	if count == 0 {
		return c.JSON(http.StatusOK, res.Paged{Data: []res.SubjectV0{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.Check(count); err != nil {
		return err
	}

	subjects, err := h.subject.Browse(c.Request().Context(), filter, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to browse subjects")
	}
	data := make([]res.SubjectV0, 0, len(subjects))
	for _, s := range subjects {
		data = append(data, convertModelSubject(s, 0))
	}

	return c.JSON(http.StatusOK, res.Paged{Data: data, Total: count, Limit: page.Limit, Offset: page.Offset})
}

func parseBrowseQuery(c echo.Context) (filter subject.BrowseFilter, err error) {
	filter = subject.BrowseFilter{}

	u := accessor.GetFromCtx(c)
	filter.NSFW = null.Bool{Value: !u.AllowNSFW(), Set: true}

	if stype, e := req.ParseSubjectType(c.QueryParam("type")); e != nil {
		err = res.BadRequest(e.Error())
		return
	} else {
		filter.Type = stype
	}

	if catStr := c.QueryParam("cat"); catStr != "" {
		if cat, e := req.ParseSubjectCategory(filter.Type, catStr); e != nil {
			err = res.BadRequest(e.Error())
			return
		} else {
			filter.Category = null.Uint16{Value: cat, Set: true}
		}
	}

	if filter.Type == model.SubjectTypeBook {
		if seriesStr := c.QueryParam("series"); seriesStr != "" {
			if series, e := gstr.ParseBool(seriesStr); e != nil {
				err = res.BadRequest(e.Error())
				return
			} else {
				filter.Series = null.Bool{Value: series, Set: true}
			}
		}
	}

	if filter.Type == model.SubjectTypeGame {
		if platform := c.QueryParam("platform"); platform != "" {
			// TODO: check if platform is valid
			filter.Platform = null.String{Value: platform, Set: true}
		}
	}

	if order := c.QueryParam("order"); order != "" {
		switch order {
		case "rank", "date":
			filter.Order = null.String{Value: order, Set: true}
		default:
			err = res.BadRequest("unknown order: " + order)
			return
		}
	}

	if yearStr := c.QueryParam("year"); yearStr != "" {
		if year, e := gstr.ParseInt32(yearStr); e != nil {
			err = res.BadRequest(e.Error())
			return
		} else {
			filter.Year = null.Int32{Value: year, Set: true}
		}
	}

	if monthStr := c.QueryParam("month"); monthStr != "" {
		if month, e := gstr.ParseInt8(monthStr); e != nil {
			err = res.BadRequest(e.Error())
			return
		} else {
			filter.Month = null.Int8{Value: month, Set: true}
		}
	}

	return
}
