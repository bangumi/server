package handler

import (
	"errors"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"
	"net/http"
	"strconv"
	"time"
)

func (h Handler) GetEpisodeComments(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	e, err := h.episode.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get episode")
	}

	_, err = h.subject.Get(c.Request().Context(), e.SubjectID, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to find subject of episode")
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

	offset, err = strconv.Atoi(offsetStr)
	if err != nil {
		return res.BadRequest(err.Error())
	}
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		return res.BadRequest(err.Error())
	}

	r, err := h.episode.GetAllComment(c.Request().Context(), id, offset, limit)
	if err != nil {
		return res.NotFound("cannot get episode comments")
	}

	result := make([]res.EpisodeCommentResp, len(r))

	for _, v := range r {
		result = append(result, res.ConventEpisodeComment2Resp(v))
	}
	return c.JSON(http.StatusOK, result)
}

func (h Handler) PostEpisodeComment(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	e, err := h.episode.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get episode")
	}

	_, err = h.subject.Get(c.Request().Context(), e.SubjectID, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to find subject of episode")
	}

	var comment req.EpisodeComment
	if err := c.Echo().JSONSerializer.Deserialize(c, comment); err != nil {
		return res.BadRequest(err.Error())
	}

	err = h.episode.AddNewComment(c.Request().Context(), model.EpisodeComment{
		ID:        0,
		Field:     comment.FieldID,
		User:      u.ID,
		Related:   comment.FieldID,
		CreatedAt: time.Now(),
		Content:   comment.Comment,
	})

	if err != nil {
		return res.BadRequest("cannot add new comment")
	}
	return c.NoContent(http.StatusOK)
}
