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

package ctrl

import (
	"context"
	"errors"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

var errTypeBlocked = errors.New("have been blocked")
var errTypeReceiverRejectPrivateMessage = errors.New("some receivers reject private message")
var errTypeNotAFriend = errors.New("not a friend to some receivers")

func (ctl Ctrl) checkNeedFriendshipReceivers(
	ctx context.Context,
	senderID model.UserID,
	receiverIDs []model.UserID,
	fieldsMap map[model.UserID]model.UserFields) error {
	checkFriendshipList := slice.MapFilter(receiverIDs, func(id model.UserID) (model.UserID, bool) {
		if fields, ok := fieldsMap[id]; ok {
			if fields.Privacy.ReceivePrivateMessage == model.UserReceiveFilterFriends {
				return id, true
			}
		}
		return 0, false
	})
	if len(checkFriendshipList) != 0 {
		ok, checkErr := ctl.user.CheckIsFriendToOthers(ctx, senderID, checkFriendshipList...)
		if checkErr != nil {
			return errgo.Wrap(checkErr, "dal")
		}
		if !ok {
			return errTypeNotAFriend
		}
	}
	return nil
}

func (ctl Ctrl) checkReceivers(ctx context.Context,
	senderID model.UserID,
	receiverIDs []model.UserID) error {
	receivers, err := ctl.user.GetByIDs(ctx, receiverIDs...)
	if err != nil {
		return errgo.Wrap(err, "dal")
	}
	if len(receivers) != len(receiverIDs) {
		return errgo.Wrap(err, "some receivers not exist")
	}
	fieldsMap, err := ctl.user.GetFieldsByIDs(ctx, receiverIDs)
	if err != nil {
		return errgo.Wrap(err, "dal")
	}
	for _, id := range receiverIDs {
		if fields, ok := fieldsMap[id]; ok {
			i := slice.FindIndex(fields.Blocklist, func(v model.UserID) bool { return v == senderID })
			if i != -1 {
				return errTypeBlocked
			}
			if fields.Privacy.ReceivePrivateMessage == model.UserReceiveFilterNone {
				return errTypeReceiverRejectPrivateMessage
			}
		}
	}

	err = ctl.checkNeedFriendshipReceivers(ctx, senderID, receiverIDs, fieldsMap)
	return err
}

func (ctl Ctrl) CreatePrivateMessage(
	ctx context.Context,
	senderID model.UserID,
	receiverIDs []model.UserID,
	relatedIDFilter domain.PrivateMessageIDFilter,
	title string,
	content string) ([]model.PrivateMessage, error) {
	emptyList := make([]model.PrivateMessage, 0)
	err := ctl.checkReceivers(ctx, senderID, receiverIDs)
	if err != nil {
		return emptyList, errgo.Wrap(err, "dal")
	}
	res, err := ctl.privateMessage.Create(ctx, senderID, receiverIDs, relatedIDFilter, title, content)
	if err != nil {
		return emptyList, errgo.Wrap(err, "dal")
	}
	return res, nil
}
