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

package index

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bytedance/sonic/decoder"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/internal/cachekey"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Handler) GetIndex(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseIndexID(c.Param("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getIndexWithCache(c.Request().Context(), id)
	if err != nil {
		return errgo.Wrap(err, "failed to get index")
	}

	if !ok || r.NSFW && !user.AllowNSFW() {
		return res.NotFound("index not found")
	}

	return c.JSON(http.StatusOK, r)
}

func (h Handler) getIndexWithCache(c context.Context, id uint32) (res.Index, bool, error) {
	var key = cachekey.Index(id)

	var r res.Index
	ok, err := h.cache.Get(c, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, ok, nil
	}

	i, err := h.i.Get(c, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.Index{}, false, nil
		}

		return res.Index{}, false, errgo.Wrap(err, "Index.Get")
	}

	u, err := h.u.GetByID(c, i.CreatorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.log.Error("index missing creator", zap.Uint32("index_id", id), i.CreatorID.Zap())
		}
		return res.Index{}, false, errgo.Wrap(err, "failed to get creator: user.GetByID")
	}

	r = res.IndexModelToResponse(&i, u)

	_ = h.cache.Set(c, key, r, time.Hour)

	return r, true, nil
}

func (h Handler) GetIndexSubjects(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseIndexID(c.Param("id"))
	if err != nil {
		return err
	}

	subjectType, err := req.ParseSubjectType(c.QueryParam("type"))
	if err != nil {
		return errgo.Wrap(err, "invalid query `type` for subject type")
	}

	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	r, ok, err := h.getIndexWithCache(c.Request().Context(), id)
	if err != nil {
		return errgo.Wrap(err, "failed to get index")
	}

	if !ok || (r.NSFW && !user.AllowNSFW()) {
		return res.ErrNotFound
	}

	return h.getIndexSubjects(c, id, subjectType, page)
}

func (h Handler) getIndexSubjects(
	c echo.Context, id model.IndexID, subjectType uint8, page req.PageQuery,
) error {
	count, err := h.i.CountSubjects(c.Request().Context(), id, subjectType)
	if err != nil {
		return errgo.Wrap(err, "Index.CountSubjects")
	}

	if count == 0 {
		return c.JSON(http.StatusOK, res.Paged{
			Data:   []int{},
			Total:  0,
			Limit:  page.Limit,
			Offset: page.Offset,
		})
	}

	if err = page.Check(count); err != nil {
		return err
	}

	subjects, err := h.i.ListSubjects(c.Request().Context(), id, subjectType, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "Index.ListSubjects")
	}

	var data = make([]res.IndexSubjectV0, len(subjects))
	for i, s := range subjects {
		data[i] = indexSubjectToResp(s)
	}

	return c.JSON(http.StatusOK, res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (h Handler) NewIndex(c echo.Context) error {
	var reqData req.IndexBasicInfo
	if err := decoder.NewStreamDecoder(c.Request().Body).Decode(&reqData); err != nil {
		return res.JSONError(c, err)
	}
	if err := h.ensureValidStrings(reqData.Description, reqData.Title); err != nil {
		return err
	}
	accessor := accessor.GetFromCtx(c)
	now := time.Now()
	i := &model.Index{
		ID:          0,
		CreatedAt:   now,
		UpdatedAt:   now,
		Title:       reqData.Title,
		Description: reqData.Description,
		CreatorID:   accessor.ID,
		Total:       0,
		Comments:    0,
		Collects:    0,
		Ban:         false,
		NSFW:        false,
	}
	ctx := c.Request().Context()
	if err := h.i.New(ctx, i); err != nil {
		return errgo.Wrap(err, "failed to create a new index")
	}
	u, err := h.u.GetByID(ctx, i.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "failed to get user info")
	}
	resp := res.IndexModelToResponse(i, u)
	return c.JSON(http.StatusOK, resp)
}

// 确保目录存在, 并且当前请求的用户持有权限.
func (h Handler) ensureIndexPermission(c echo.Context, indexID uint32) (*model.Index, error) {
	accessor := accessor.GetFromCtx(c)
	index, err := h.i.Get(c.Request().Context(), indexID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, res.NotFound("index not found")
		}
		return nil, res.InternalError(c, err, "failed to get index")
	}
	if index.CreatorID != accessor.ID {
		return nil, res.Unauthorized("you are not the creator of this index")
	}
	return &index, nil
}

func (h Handler) UpdateIndex(c echo.Context) error {
	indexID, err := req.ParseIndexID(c.Param("id"))
	if err != nil {
		return err
	}
	var reqData req.IndexBasicInfo
	if err = decoder.NewStreamDecoder(c.Request().Body).Decode(&reqData); err != nil {
		return res.JSONError(c, err)
	}

	if reqData.Title == "" && reqData.Description == "" {
		return res.BadRequest("request data is empty")
	}

	if err = h.ensureValidStrings(reqData.Description, reqData.Title); err != nil {
		return err
	}

	index, err := h.ensureIndexPermission(c, indexID)
	if err != nil {
		return err
	}
	if err = h.i.Update(c.Request().Context(), index.ID, reqData.Title, reqData.Description); err != nil {
		return errgo.Wrap(err, "update index failed")
	}
	h.invalidateIndexCache(c.Request().Context(), index.ID)
	return nil
}
