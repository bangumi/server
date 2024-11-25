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

	count, err := h.subject.Count(c.Request().Context(), *filter)
	if err != nil {
		return errgo.Wrap(err, "failed to count subjects")
	}

	if count == 0 {
		return c.JSON(http.StatusOK, res.Paged{
			Data: []res.SubjectV0{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.Check(count); err != nil {
		return err
	}

	subjects, err := h.subject.Browse(c.Request().Context(), *filter, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to browse subjects")
	}
	ids := make([]model.SubjectID, 0, len(subjects))
	for _, s := range subjects {
		ids = append(ids, s.ID)
	}
	tags, err := h.tag.GetByIDs(c.Request().Context(), ids)
	if err != nil {
		return errgo.Wrap(err, "failed to get tags")
	}

	data := make([]res.SubjectV0, 0, len(subjects))
	for _, s := range subjects {
		metaTags := tags[s.ID]
		data = append(data, res.ToSubjectV0(s, 0, metaTags))
	}

	return c.JSON(http.StatusOK, res.Paged{Data: data, Total: count, Limit: page.Limit, Offset: page.Offset})
}

func parseBrowseQuery(c echo.Context) (*subject.BrowseFilter, error) {
	filter := subject.BrowseFilter{}
	u := accessor.GetFromCtx(c)
	filter.NSFW = null.Bool{Value: false, Set: !u.AllowNSFW()}
	if stype, err := req.ParseSubjectType(c.QueryParam("type")); err != nil {
		return nil, res.BadRequest(err.Error())
	} else {
		filter.Type = stype
	}
	if catStr := c.QueryParam("cat"); catStr != "" {
		if cat, err := req.ParseSubjectCategory(filter.Type, catStr); err != nil {
			return nil, res.BadRequest(err.Error())
		} else {
			filter.Category = null.Uint16{Value: cat, Set: true}
		}
	}
	if filter.Type == model.SubjectTypeBook {
		if seriesStr := c.QueryParam("series"); seriesStr != "" {
			if series, err := gstr.ParseBool(seriesStr); err != nil {
				return nil, res.BadRequest(err.Error())
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
	if sort := c.QueryParam("sort"); sort != "" {
		switch sort {
		case "rank", "date":
			filter.Sort = null.String{Value: sort, Set: true}
		default:
			return nil, res.BadRequest("unknown sort: " + sort)
		}
	}
	if year, err := GetYearQuery(c); err != nil {
		return nil, err
	} else {
		filter.Year = year
	}
	if month, err := GetMonthQuery(c); err != nil {
		return nil, err
	} else {
		filter.Month = month
	}

	return &filter, nil
}

func GetYearQuery(c echo.Context) (null.Int32, error) {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		return null.Int32{}, nil
	}
	if year, err := gstr.ParseInt32(yearStr); err != nil {
		return null.Int32{}, res.BadRequest(err.Error())
	} else {
		if year < 1900 || year > 3000 {
			return null.Int32{}, res.BadRequest("invalid year: " + yearStr)
		}
		return null.Int32{Value: year, Set: true}, nil
	}
}

func GetMonthQuery(c echo.Context) (null.Int8, error) {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return null.Int8{}, nil
	}
	if month, err := gstr.ParseInt8(monthStr); err != nil {
		return null.Int8{}, res.BadRequest(err.Error())
	} else {
		if month < 1 || month > 12 {
			return null.Int8{}, res.BadRequest("invalid month: " + monthStr)
		}
		return null.Int8{Value: month, Set: true}, nil
	}
}
