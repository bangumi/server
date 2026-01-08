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

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/internal/cachekey"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

type indexCacheValue struct {
	Index     res.Index          `json:"index"`
	Privacy   model.IndexPrivacy `json:"privacy"`
	CreatorID model.UserID       `json:"creator_id"`
}

func (h Handler) GetIndex(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	resp, ok, err := h.getIndexWithCache(c.Request().Context(), user, id)
	if err != nil {
		return errgo.Wrap(err, "failed to get index")
	}

	if !ok {
		return res.NotFound("index not found")
	}

	return c.JSON(http.StatusOK, resp)
}

func (h Handler) getIndexWithCache(ctx context.Context, user *accessor.Accessor, id uint32) (res.Index, bool, error) {
	key := cachekey.Index(id)

	userID, allowNSFW := h.extractUserPrefs(user)

	if cached, ok, err := h.getIndexFromCache(ctx, key, userID, allowNSFW); err != nil || ok {
		return cached, ok, err
	}

	item, ok, err := h.buildIndexResponse(ctx, id, userID, allowNSFW)
	if err != nil || !ok {
		return item.Index, ok, err
	}

	if item.Privacy == model.IndexPrivacyPublic {
		_ = h.cache.Set(ctx, key, item, time.Hour)
	}

	return item.Index, true, nil
}

func (h Handler) getIndexFromCache(
	ctx context.Context, key string, userID model.UserID, allowNSFW bool,
) (res.Index, bool, error) {
	var cached indexCacheValue
	ok, err := h.cache.Get(ctx, key, &cached)
	if err != nil {
		return res.Index{}, ok, errgo.Wrap(err, "cache.Get")
	}

	if !ok {
		return res.Index{}, false, nil
	}

	if !isIndexVisible(cached.Privacy, cached.CreatorID, userID) {
		return res.Index{}, false, nil
	}
	if cached.Index.NSFW && !allowNSFW {
		return res.Index{}, false, nil
	}

	return cached.Index, true, nil
}

func (h Handler) buildIndexResponse(
	ctx context.Context, id uint32, userID model.UserID, allowNSFW bool,
) (indexCacheValue, bool, error) {
	i, err := h.i.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return indexCacheValue{}, false, nil
		}

		return indexCacheValue{}, false, errgo.Wrap(err, "Index.Get")
	}

	if !isIndexVisible(i.Privacy, i.CreatorID, userID) {
		return indexCacheValue{}, false, nil
	}

	u, err := h.u.GetByID(ctx, i.CreatorID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			h.log.Error("index missing creator", zap.Uint32("index_id", id), zap.Uint32("creator", i.CreatorID))
		}
		return indexCacheValue{}, false, errgo.Wrap(err, "failed to get creator: user.GetByID")
	}

	r := res.IndexModelToResponse(&i, u)
	if r.NSFW && !allowNSFW {
		return indexCacheValue{}, false, nil
	}

	return indexCacheValue{Index: r, Privacy: i.Privacy, CreatorID: i.CreatorID}, true, nil
}

func isIndexVisible(privacy model.IndexPrivacy, creatorID, userID model.UserID) bool {
	if privacy == model.IndexPrivacyDeleted {
		return false
	}
	if privacy == model.IndexPrivacyPrivate && creatorID != userID {
		return false
	}
	return true
}

func (h Handler) extractUserPrefs(user *accessor.Accessor) (model.UserID, bool) {
	if user == nil {
		return 0, false
	}

	return user.ID, user.AllowNSFW()
}

func (h Handler) GetIndexSubjects(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
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

	_, ok, err := h.getIndexWithCache(c.Request().Context(), user, id)
	if err != nil {
		return errgo.Wrap(err, "failed to get index")
	}

	if !ok {
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
	if err := c.Echo().JSONSerializer.Deserialize(c, &reqData); err != nil {
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
		if errors.Is(err, gerr.ErrNotFound) {
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
	indexID, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	var reqData req.IndexBasicInfo
	if err = c.Echo().JSONSerializer.Deserialize(c, &reqData); err != nil {
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
