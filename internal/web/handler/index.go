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
		data[i] = res.IndexSubjectV0{
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
	i := &model.Index{
		ID:          0,
		CreatedAt:   time.Now(),
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

func (h Handler) UpdateIndex(c *fiber.Ctx) error {
	return nil
}

func (h Handler) DeleteIndex(c *fiber.Ctx) error {
	return nil
}

func (h Handler) AddIndexSubject(c *fiber.Ctx) error {
	return nil
}

func (h Handler) UpdateIndexSubject(c *fiber.Ctx) error {
	return nil
}

func (h Handler) RemoveIndexSubject(c *fiber.Ctx) error {
	return nil
}
