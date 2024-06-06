package subject

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) GetComment(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	commentID, err := req.ParseID(c.Param("post_id"))
	if err != nil {
		return err
	}

	_, err = h.subject.Get(c.Request().Context(), id, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})

	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "failed to get subject")
	}

	result, err := h.subject.GetPost(c.Request().Context(), commentID)
	if err != nil {
		return res.BadRequest("cannot found comment")
	}
	resp := res.ConventSubjectComment2Resp(result)

	return c.JSON(http.StatusOK, resp)
}

func (h Subject) GetComments(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	s, err := h.subject.Get(c.Request().Context(), id, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject")
	}

	var offset, limit int

	offsetStr := c.QueryParam("offset")
	limitStr := c.QueryParam("limit")

	if offsetStr == "" {
		offset = 0 // 默认为0
	}
	if limitStr == "" {
		limit = 25 // 默认25
	}

	result, err := h.subject.GetAllPost(c.Request().Context(), s.ID, offset, limit)
	if err != nil {
		return res.BadRequest("cannot found comment")
	}
	resp := make([]res.SubjectPost, 0)
	for _, v := range result {
		resp = append(resp, res.ConventSubjectComment2Resp(v))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h Subject) AddComment(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	s, err := h.subject.Get(c.Request().Context(), id, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject")
	}

	var reqBody = req.SubjectComment{}
	if err = c.Echo().JSONSerializer.Deserialize(c, &reqBody); err != nil {
		return res.JSONError(c, err)
	}

	// 校验回复消息是否存在
	if reqBody.FieldID != 0 {
		_, err = h.subject.GetPost(c.Request().Context(), reqBody.FieldID)
		if err != nil {
			return res.NotFound("cannot find comment to reply")
		}
	}

	err = h.subject.NewPost(c.Request().Context(), model.SubjectPost{
		Field:     s.ID,
		User:      u.ID,
		Related:   reqBody.FieldID,
		CreatedAt: time.Now(),
		Content:   reqBody.Content,
		State:     0,
	})
	if err != nil {
		return res.BadRequest("cannot add comment to subject")
	}
	return c.NoContent(http.StatusOK)
}

func (h Subject) RemoveComment(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	commentID, err := req.ParseID(c.Param("post_id"))
	if err != nil {
		return err
	}

	_, err = h.subject.Get(c.Request().Context(), id, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject")
	}

	// 校验消息是否存在以及是否为本人发送
	comment, err := h.subject.GetPost(c.Request().Context(), commentID)
	if err != nil {
		return res.NotFound("cannot find comment")
	}

	if comment.User != u.ID {
		return res.Forbidden("cannot remove comment")
	}

	err = h.subject.DeletePost(c.Request().Context(), commentID)

	if err != nil {
		return res.BadRequest("cannot remove comment")
	}

	return c.NoContent(http.StatusOK)
}
