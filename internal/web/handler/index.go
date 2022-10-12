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

package handler

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/compat"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/web/handler/internal/cachekey"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/pkg/wiki"
)

func modelToResponse(i *model.Index, u model.User) res.Index {
	return res.Index{
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Creator: res.Creator{
			Username: u.UserName,
			Nickname: u.NickName,
		},
		Title:       i.Title,
		Description: i.Description,
		Total:       i.Total,
		ID:          i.ID,
		Stat: res.Stat{
			Comments: i.Comments,
			Collects: i.Collects,
		},
		Ban:  i.Ban,
		NSFW: i.NSFW,
	}
}

func (h Handler) getIndexWithCache(c context.Context, id uint32) (res.Index, bool, error) {
	var key = cachekey.Index(id)

	var r res.Index
	ok, err := h.cache.Get(c, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	i, err := h.i.Get(c, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.Index{}, false, nil
		}

		return res.Index{}, false, errgo.Wrap(err, "Index.Get")
	}

	u, err := h.ctrl.GetUser(c, i.CreatorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.log.Error("index missing creator", zap.Uint32("index_id", id), i.CreatorID.Zap())
		}
		return res.Index{}, false, errgo.Wrap(err, "failed to get creator: user.GetByID")
	}

	r = modelToResponse(&i, u)

	if e := h.cache.Set(c, key, r, time.Hour); e != nil {
		h.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}

func (h Handler) GetIndex(c *fiber.Ctx) error {
	user := h.GetHTTPAccessor(c)

	id, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getIndexWithCache(c.UserContext(), id)
	if err != nil {
		return err
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

	r, ok, err := h.getIndexWithCache(c.UserContext(), id)
	if err != nil {
		return err
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
	if err := json.UnmarshalNoEscape(c.Body(), &reqData); err != nil {
		return errgo.Wrap(err, "request data is invalid")
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
	resp := modelToResponse(i, u)
	return c.JSON(resp)
}

// 确保目录存在, 并且当前请求的用户持有权限.
func (h Handler) ensureIndexPermission(c *fiber.Ctx) (*model.Index, error) {
	id, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return nil, err
	}
	accessor := h.GetHTTPAccessor(c)
	// TODO: 是否走 redis 缓存
	index, err := h.i.Get(c.UserContext(), id)
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
	var reqData req.IndexBasicInfo
	if err := json.UnmarshalNoEscape(c.Body(), &reqData); err != nil {
		return errgo.Wrap(err, "request data is invalid")
	}

	if reqData.Title == "" && reqData.Description == "" {
		return res.BadRequest("request data is empty")
	}

	index, err := h.ensureIndexPermission(c)
	if err != nil {
		return err
	}
	if err = h.i.Update(c.UserContext(), index.ID, reqData.Title, reqData.Description); err != nil {
		return errgo.Wrap(err, "update index failed")
	}
	return nil
}

func (h Handler) DeleteIndex(c *fiber.Ctx) error {
	index, err := h.ensureIndexPermission(c)
	if err != nil {
		return err
	}

	if err = h.i.Delete(c.UserContext(), index.ID); err != nil {
		return errgo.Wrap(err, "failed to delete index from db")
	}

	return nil
}

func (h Handler) AddIndexSubject(c *fiber.Ctx) error {
	var reqData req.IndexAddSubject
	if err := json.UnmarshalNoEscape(c.Body(), &reqData); err != nil {
		return errgo.Wrap(err, "request data is invalid")
	}
	index, err := h.ensureIndexPermission(c)
	if err != nil {
		return err
	}
	indexSubject, err := h.i.AddIndexSubject(c.UserContext(),
		index.ID, reqData.SubjectID, reqData.SortKey, reqData.Comment)
	if err != nil || indexSubject == nil {
		return errgo.Wrap(err, "failed to edit subject in the index")
	}
	return c.JSON(indexSubjectToResp(*indexSubject))
}

func (h Handler) UpdateIndexSubject(c *fiber.Ctx) error {
	var reqData req.IndexSubjectInfo
	if err := json.UnmarshalNoEscape(c.Body(), &reqData); err != nil {
		return errgo.Wrap(err, "request data is invalid")
	}
	index, err := h.ensureIndexPermission(c)
	if err != nil {
		return err
	}
	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return errgo.Wrap(err, "subject id is invalid")
	}
	err = h.i.UpdateIndexSubject(c.UserContext(),
		index.ID, subjectID, reqData.SortKey, reqData.Comment)
	if err != nil {
		return errgo.Wrap(err, "failed to edit subject in the index")
	}
	return nil
}

func indexSubjectToResp(s domain.IndexSubject) res.IndexSubjectV0 {
	return res.IndexSubjectV0{
		AddedAt: s.AddedAt,
		Date:    null.NilString(s.Subject.Date),
		Image:   res.SubjectImage(s.Subject.Image),
		Name:    s.Subject.Name,
		NameCN:  s.Subject.NameCN,
		Comment: s.Comment,
		Infobox: compat.V0Wiki(wiki.ParseOmitError(s.Subject.Infobox).NonZero()),
		ID:      s.Subject.ID,
		TypeID:  s.Subject.TypeID,
	}
}
