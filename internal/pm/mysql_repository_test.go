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

package pm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/pm"
)

func getRepo(t *testing.T) domain.PrivateMessageRepo {
	t.Helper()
	repo, err := pm.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func mapToID(msg model.PrivateMessage) model.PrivateMessageID {
	return msg.ID
}

func mockMessage(
	ctx context.Context,
	t *testing.T,
	repo domain.PrivateMessageRepo,
	relatedID *model.PrivateMessageID,
	senderID model.UserID,
	receiverID model.UserID,
) model.PrivateMessage {
	t.Helper()
	m, err := repo.Create(
		ctx,
		senderID,
		[]model.UserID{receiverID},
		domain.PrivateMessageIDFilter{Type: null.NewFromPtr(relatedID)},
		"title",
		"content",
	)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	t.Cleanup(func() {
		err = repo.Delete(ctx, senderID, slice.Map(m, mapToID))
		require.NoError(t, err)
		err = repo.Delete(ctx, receiverID, slice.Map(m, mapToID))
		require.NoError(t, err)
	})
	return m[0]
}

func TestListInbox(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	m := mockMessage(ctx, t, repo, nil, 1, 382951)

	list, err := repo.List(ctx, 382951, model.PrivateMessageFolderTypeInbox, 0, 10)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	require.LessOrEqual(t, m.ID, list[0].Self.ID)
}

func TestListOutbox(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	m := mockMessage(ctx, t, repo, nil, 1, 382951)

	list, err := repo.List(ctx, 1, model.PrivateMessageFolderTypeOutbox, 0, 10)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	require.LessOrEqual(t, m.ID, list[0].Self.ID)
}

func TestListRelated(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()
	msg := mockMessage(ctx, t, repo, nil, 1, 382951)

	msg2 := mockMessage(ctx, t, repo, &msg.ID, 382951, 1)

	list, err := repo.ListRelated(ctx, 1, msg.ID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, msg2.ID, list[len(list)-1].ID)

	// 使用非首条信息作查询
	list, err = repo.ListRelated(ctx, 382951, msg2.ID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, msg.ID, list[0].ID)
}

func TestCountTypes(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	ctx := context.Background()
	prevCounts, err := repo.CountTypes(ctx, 5)
	require.NoError(t, err)
	prevCounts2, err := repo.CountTypes(ctx, 4)
	require.NoError(t, err)
	mockMessage(ctx, t, repo, nil, 5, 4)
	counts, err := repo.CountTypes(ctx, 5)
	require.NoError(t, err)
	counts2, err := repo.CountTypes(ctx, 4)
	require.NoError(t, err)
	require.Less(t, prevCounts.Outbox, counts.Outbox)
	require.Less(t, prevCounts2.Unread, counts2.Unread)
	require.Less(t, prevCounts2.Inbox, counts2.Inbox)
}

func TestListRecentContact(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	mockMessage(ctx, t, repo, nil, 1, 382951)

	list, err := repo.ListRecentContact(ctx, 1)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	require.Contains(t, list, model.UserID(382951))
}

func TestMarkRead(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()
	msg := mockMessage(ctx, t, repo, nil, 1, 382951)
	err := repo.MarkRead(ctx, 382951, msg.ID)
	require.NoError(t, err)
	msgs, err := repo.ListRelated(ctx, 1, msg.ID)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.Equal(t, false, msgs[0].New)
}

func TestCreate(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	mainMsgs, err := repo.Create(
		ctx,
		1,
		[]model.UserID{382951, 2},
		domain.PrivateMessageIDFilter{Type: null.NewFromPtr[model.PrivateMessageID](nil)},
		"私信",
		"内容",
	)
	require.NoError(t, err)
	require.Len(t, mainMsgs, 2)

	msgs := mainMsgs

	// reply
	replyMsgs, err := repo.Create(ctx, 382951, []model.UserID{1}, domain.PrivateMessageIDFilter{
		Type: null.New(mainMsgs[0].ID),
	}, "私信回复", "内容")

	require.NoError(t, err)
	require.Len(t, replyMsgs, 1)
	require.Equal(t, mainMsgs[0].ID, replyMsgs[0].RelatedMessageID)

	msgs = append(msgs, replyMsgs...)

	_, err = repo.Create(ctx, 382951, []model.UserID{2}, domain.PrivateMessageIDFilter{
		Type: null.New(mainMsgs[1].ID),
	}, "私信回复", "发给错误的人")

	require.Error(t, err)

	msgsNotMain, err := repo.Create(ctx, 1, []model.UserID{382951}, domain.PrivateMessageIDFilter{
		Type: null.New(replyMsgs[0].ID),
	}, "私信回复", "使用非首条信息的id作为related id")

	require.NoError(t, err)
	require.Len(t, msgsNotMain, 1)
	require.Equal(t, mainMsgs[0].ID, msgsNotMain[0].RelatedMessageID)
	msgs = append(msgs, msgsNotMain...)

	t.Cleanup(func() {
		for _, msg := range msgs {
			err = repo.Delete(ctx, msg.SenderID, slice.Map([]model.PrivateMessage{msg}, mapToID))
			require.NoError(t, err)
			err = repo.Delete(ctx, msg.ReceiverID, slice.Map([]model.PrivateMessage{msg}, mapToID))
			require.NoError(t, err)
		}
	})
}

func TestDelete(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	ctx := context.Background()
	res, err := repo.Create(
		ctx,
		1,
		[]model.UserID{382951},
		domain.PrivateMessageIDFilter{Type: null.NewFromPtr[model.PrivateMessageID](nil)},
		"私信",
		"内容",
	)
	require.NoError(t, err)
	require.Len(t, res, 1)
	err = repo.Delete(ctx, 1, []model.PrivateMessageID{res[0].ID})
	require.NoError(t, err)
	_, err = repo.ListRelated(ctx, 1, res[0].ID)
	require.Error(t, err)
	res, err = repo.ListRelated(ctx, 382951, res[0].ID)
	require.NoError(t, err)
	require.Len(t, res, 1)
	t.Cleanup(func() {
		err := repo.Delete(ctx, 382951, []model.PrivateMessageID{res[0].ID})
		require.NoError(t, err)
	})
}
