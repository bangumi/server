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
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Handler) GetIndex(c *fiber.Ctx) error {
	user := h.GetHTTPAccessor(c)

	id, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.ctrl.GetIndexWithCache(c.UserContext(), id)
	if err != nil {
		return errgo.Wrap(err, "failed to get index")
	}

	if !ok || r.NSFW && !user.AllowNSFW() {
		return res.NotFound("index not found")
	}

	return c.JSON(r)
}

func (h Handler) GetIndexSubjects(c *fiber.Ctx) error {
	user := h.GetHTTPAccessor(c)

	id, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}

	subjectType, err := req.ParseSubjectType(c.Query("type"))
	if err != nil {
		return errgo.Wrap(err, "invalid query `type` for subject type")
	}

	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	r, ok, err := h.ctrl.GetIndexWithCache(c.UserContext(), id)
	if err != nil {
		return errgo.Wrap(err, "failed to get index")
	}

	if !ok || (r.NSFW && !user.AllowNSFW()) {
		return res.ErrNotFound
	}

	return h.getIndexSubjects(c, id, subjectType, page)
}

func (h Handler) getIndexSubjects(
	c *fiber.Ctx, id model.IndexID, subjectType uint8, page req.PageQuery,
) error {
	count, err := h.i.CountSubjects(c.UserContext(), id, subjectType)
	if err != nil {
		return errgo.Wrap(err, "Index.CountSubjects")
	}

	if count == 0 {
		return c.JSON(res.Paged{
			Data:   []int{},
			Total:  0,
			Limit:  page.Limit,
			Offset: page.Offset,
		})
	}

	if err = page.Check(count); err != nil {
		return err
	}

	subjects, err := h.i.ListSubjects(c.UserContext(), id, subjectType, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "Index.ListSubjects")
	}

	var data = make([]res.IndexSubjectV0, len(subjects))
	for i, s := range subjects {
		data[i] = indexSubjectToResp(s)
	}

	return c.JSON(res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (h Handler) NewIndex(c *fiber.Ctx) error {
	var reqData req.IndexBasicInfo
	if err := sonic.Unmarshal(c.Body(), &reqData); err != nil {
		return res.JSONError(c, err)
	}
	if err := h.ensureValidStrings(reqData.Description, reqData.Title); err != nil {
		return err
	}
	accessor := h.GetHTTPAccessor(c)
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
	ctx := c.UserContext()
	if err := h.i.New(ctx, i); err != nil {
		return errgo.Wrap(err, "failed to create a new index")
	}
	u, err := h.ctrl.GetUser(ctx, i.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "failed to get user info")
	}
	resp := res.IndexModelToResponse(i, u)
	return c.JSON(resp)
}

// 确保目录存在, 并且当前请求的用户持有权限.
func (h Handler) ensureIndexPermission(c *fiber.Ctx, indexID uint32) (*model.Index, error) {
	accessor := h.GetHTTPAccessor(c)
	index, err := h.i.Get(c.UserContext(), indexID)
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

func (h Handler) UpdateIndex(c *fiber.Ctx) error {
	indexID, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}
	var reqData req.IndexBasicInfo
	if err = sonic.Unmarshal(c.Body(), &reqData); err != nil {
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
	if err = h.i.Update(c.UserContext(), index.ID, reqData.Title, reqData.Description); err != nil {
		return errgo.Wrap(err, "update index failed")
	}
	h.invalidateIndexCache(c.UserContext(), index.ID)
	return nil
}
