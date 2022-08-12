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

package canal

import (
	"context"
	"fmt"

	"github.com/gookit/event"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/session"
)

const (
	EventUserChangePassword = "user-change-password" // 用户在旧站修改密码，吊销全部 session
	EventSubjectCreate      = "subject.create"
	EventSubjectUpdate      = "subject.update"
	EventSubjectDelete      = "subject.delete"
)

func NewUserChangePassword(userID model.UserID) event.Event {
	return event.NewBasic(EventUserChangePassword, event.M{"user_id": userID})
}

func GetUserChangePasswordEventPayload(e event.Event) model.UserID {
	return e.Data()["user_id"].(model.UserID)
}

func NewSubjectEvent(name string, subjectID model.SubjectID) event.Event {
	return event.NewBasic(name, event.M{"subject_id": subjectID})
}

func GetSubjectEventPayload(e event.Event) model.SubjectID {
	return e.Data()["subject_id"].(model.SubjectID)
}

func eventManager(
	logger *zap.Logger,
	session session.Manager,
) *event.Manager {
	e := event.NewManager("chii")
	// e.On(EventSubjectCreate, event.ListenerFunc(func(e event.Event) error {
	// 	id := GetSubjectEventPayload(e)
	// 	fmt.Println(e.Name(), id)
	// 	return nil
	// }))

	logger = logger.Named("event")
	e.On(EventUserChangePassword, event.ListenerFunc(func(e event.Event) error {
		id := GetUserChangePasswordEventPayload(e)
		fmt.Println(e.Name(), id)
		err := session.RevokeUser(context.Background(), id)
		if err != nil {
			logger.Error("failed to revoke user", log.UserID(id), zap.Error(err))
		}
		return nil
	}))

	return e
}
