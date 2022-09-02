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
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

const groupIconPrefix = "https://lain.bgm.tv/pic/icon/l/"

func (h Handler) GetGroupProfileByNamePrivate(c *fiber.Ctx) error {
	groupName := c.Params("name")
	if groupName == "" {
		return res.BadRequest("group name is required")
	}

	g, err := h.g.GetByName(c.Context(), groupName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("group not found")
		}

		return errgo.Wrap(err, "failed to get group")
	}

	members, err := h.listGroupMembers(c.Context(), g.ID, domain.GroupMemberAll, 10, 0)
	if err != nil {
		return errgo.Wrap(err, "failed to list recent members")
	}

	return res.JSON(c, res.PrivateGroupProfile{
		ID:           g.ID,
		CreatedAt:    g.CreatedAt,
		Name:         g.Name,
		Title:        g.Title,
		Description:  g.Description,
		Icon:         groupIconPrefix + g.Icon,
		TotalMembers: g.MemberCount,
		NewMembers:   members,
	})
}

func (h Handler) ListGroupMembersPrivate(c *fiber.Ctx) error {
	// parse request info
	groupName := c.Params("name")
	if groupName == "" {
		return res.BadRequest("group name is required")
	}

	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	memberType, err := parseGroupMemberType(c)
	if err != nil {
		return err
	}

	// end parse query, fetch data
	return h.listGroupMembersPrivate(c, groupName, page, memberType)
}

// fetch data.
func (h Handler) listGroupMembersPrivate(
	c *fiber.Ctx,
	groupName string,
	page req.PageQuery,
	memberType domain.GroupMemberType,
) error {
	g, err := h.g.GetByName(c.Context(), groupName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("group not found")
		}

		return errgo.Wrap(err, "failed to get group")
	}

	memberCount, err := h.g.CountMembersByID(c.Context(), g.ID, memberType)
	if err != nil {
		return errgo.Wrap(err, "failed to count group member")
	}

	if memberCount == 0 {
		return res.JSON(c, res.Paged{Data: res.EmptySlice(), Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.Check(memberCount); err != nil {
		return err
	}

	data, err := h.listGroupMembers(c.Context(), g.ID, memberType, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to list group members")
	}

	return res.JSON(c, res.Paged{
		Data:   data,
		Total:  memberCount,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (h Handler) listGroupMembers(
	ctx context.Context,
	groupID model.GroupID,
	memberType domain.GroupMemberType,
	limit, offset int,
) ([]res.PrivateGroupMember, error) {
	members, err := h.g.ListMembersByID(ctx, groupID, memberType, limit, offset)
	if err != nil {
		h.log.Error("failed to list recent members", zap.Error(err), log.GroupID(groupID))
		return nil, errgo.Wrap(err, "groupRepo。ListMembersByID")
	}

	userIDs := make([]model.UserID, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID
	}
	userMap, err := h.ctrl.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, errgo.Wrap(err, "userRepo.GetByIDs")
	}

	return convertGroupMembers(members, userMap), nil
}

func convertGroupMembers(members []model.GroupMember, userMap map[model.UserID]model.User) []res.PrivateGroupMember {
	s := make([]res.PrivateGroupMember, len(members))
	for i, member := range members {
		u := userMap[member.UserID]
		s[i] = res.PrivateGroupMember{
			ID:       member.UserID,
			UserName: u.UserName,
			NickName: u.NickName,
			JoinedAt: member.JoinAt,
			Avatar:   res.UserAvatar(u.Avatar),
		}
	}

	return s
}

func parseGroupMemberType(c *fiber.Ctx) (domain.GroupMemberType, error) {
	var memberType = domain.GroupMemberAll

	memberQuery := c.Query("type")
	switch memberQuery {
	case "mod":
		memberType = domain.GroupMemberMod
	case "normal":
		memberType = domain.GroupMemberNormal
	case "all", "":
	default:
		return 0, res.BadRequest(strconv.Quote(memberQuery) +
			` is not a valid group member type, allowed: "mod", "normal", "all"(default)`)
	}

	return memberType, nil
}

func (h Handler) ListGroupTopics(c *fiber.Ctx) error {
	groupName := c.Params("name")
	if groupName == "" {
		return res.BadRequest("group name is required")
	}

	g, err := h.g.GetByName(c.Context(), groupName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("group not found")
		}

		return errgo.Wrap(err, "failed to get group")
	}

	return h.listTopics(c, domain.TopicTypeGroup, uint32(g.ID))
}
