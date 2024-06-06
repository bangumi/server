package index

import (
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"
	"net/http"
	"strconv"
	"time"
)

func (h Handler) GetComment(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))

	if err != nil {
		return err
	}

	commentID, err := req.ParseID(c.Param("id"))

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

	result, err := h.i.GetIndexComment(c.Request().Context(), commentID)
	if err != nil {
		return res.NotFound("comment not found")
	}

	resp := res.ConventIndexCommit2Resp(*result)

	return c.JSON(http.StatusOK, resp)
}

func (h Handler) GetComments(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
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
	result, err := h.i.GetIndexComments(c.Request().Context(), id, offset, limit)

	if err != nil {
		return res.NotFound("comment not found")
	}
	var resp = make([]res.IndexCommentResp, len(result))

	for _, v := range result {
		resp = append(resp, res.ConventIndexCommit2Resp(v))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h Handler) AddComment(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
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

	var comment req.IndexComment

	if err := c.Echo().JSONSerializer.Deserialize(c, &comment); err != nil {
		return res.JSONError(c, err)
	}
	if err := h.ensureValidStrings(comment.Comment); err != nil {
		return err
	}

	if comment.FieldID != 0 {
		// 验证回复的消息是否存在
		_, err := h.i.GetIndexComment(c.Request().Context(), comment.FieldID)
		if err != nil {
			return res.NotFound("comment to reply is not found")
		}
	}

	err = h.i.AddIndexComment(c.Request().Context(), model.IndexComment{
		Field:     id,
		User:      user.ID,
		CreatedAt: time.Now(),
		Related:   comment.FieldID,
		Content:   comment.Comment,
	})
	if err != nil {
		return res.BadRequest("cannot add comment to index")
	}

	return c.NoContent(http.StatusOK)
}

func (h Handler) RemoveComment(c echo.Context) error {
	user := accessor.GetFromCtx(c)

	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	commentID, err := req.ParseID(c.Param("comment_id"))
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

	// 验证消息是否存在，且是否为当前用户所发
	cmt, err := h.i.GetIndexComment(c.Request().Context(), commentID)
	if err != nil {
		return res.NotFound("comment not found")
	}
	if cmt.User != user.ID {
		return res.Unauthorized("cannot remove comment from other user")
	}

	err = h.i.DeleteIndexComment(c.Request().Context(), commentID)

	if err != nil {
		return res.BadRequest("cannot remove comment from index")
	}
	return c.NoContent(http.StatusOK)
}
