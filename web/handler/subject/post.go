package subject

import (
	"errors"
	"net/http"
	"strconv"
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

func (h Subject) GetPost(c echo.Context) error {
	postID, err := req.ParseID(c.Param("post_id"))
	if err != nil {
		return err
	}

	result, err := h.subject.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return res.BadRequest("cannot found subject post")
	}
	resp := res.ConventSubjectComment2Resp(result)

	return c.JSON(http.StatusOK, resp)
}

func (h Subject) GetPostReplies(c echo.Context) error {
	postID, err := req.ParseID(c.Param("post_id"))
	if err != nil {
		return err
	}

	pq, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return res.BadRequest("cannot get offset and limit")
	}

	replies, err := h.subject.GetPaginatedRepliesByPostID(c.Request().Context(), postID, pq.Offset, pq.Limit)
	if err != nil {
		return res.BadRequest("cannot to get comment replies")
	}
	resp := make([]res.SubjectPost, 0)
	for _, v := range replies {
		resp = append(resp, res.ConventSubjectComment2Resp(v))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h Subject) GetPaginatedPosts(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	s, err := h.subject.Get(c.Request().Context(), subjectID, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject")
	}

	pq, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return res.BadRequest("cannot get offset and limit")
	}

	result, err := h.subject.GetPaginatedPostsBySubjectID(c.Request().Context(), s.ID, pq.Offset, pq.Limit)
	if err != nil {
		return res.BadRequest("cannot found comment")
	}
	resp := make([]res.SubjectPost, 0)
	for _, v := range result {
		resp = append(resp, res.ConventSubjectComment2Resp(v))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h Subject) GetPostsWithReplies(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	repliesLimitRaw := c.Param("replies_limit")
	repliesLimit, err := strconv.Atoi(repliesLimitRaw)
	if err != nil {
		return res.BadRequest("can't parse query args replies_limit as int: " + strconv.Quote(repliesLimitRaw))
	}

	s, err := h.subject.Get(c.Request().Context(), subjectID, subject.Filter{
		NSFW: null.Bool{Value: false, Set: !u.AllowNSFW()},
	})
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject")
	}

	pq, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return res.BadRequest("cannot get offset and limit")
	}

	topPosts, err := h.subject.GetPaginatedTopLevelPostsBySubjectID(c.Request().Context(), s.ID, pq.Offset, pq.Limit)
	if err != nil {
		return res.BadRequest("cannot found subject post")
	}
	resp := make([]model.SubjectPost, 0, len(topPosts))
	for _, post := range topPosts {
		replies, err := h.subject.GetPaginatedRepliesByPostID(c.Request().Context(), post.ID, 0, repliesLimit)
		if err != nil {
			return res.BadRequest("failed to get post replies")
		}
		post.Replies = replies
		resp = append(resp, post)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h Subject) AddComment(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	subjectID, err := req.ParseID(c.Param("subject_id"))
	if err != nil {
		return err
	}

	s, err := h.subject.Get(c.Request().Context(), subjectID, subject.Filter{
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
		_, err = h.subject.GetPostByID(c.Request().Context(), reqBody.FieldID)
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

func (h Subject) RemovePost(c echo.Context) error {
	u := accessor.GetFromCtx(c)

	postID, err := req.ParseID(c.Param("post_id"))
	if err != nil {
		return err
	}

	// 校验消息是否存在以及是否为本人发送
	comment, err := h.subject.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return res.NotFound("cannot find post")
	}

	if comment.User != u.ID {
		return res.Forbidden("no permission to delete post not sent by oneself")
	}

	err = h.subject.DeletePostByID(c.Request().Context(), postID)

	if err != nil {
		return res.BadRequest("cannot remove post")
	}

	return c.NoContent(http.StatusOK)
}
