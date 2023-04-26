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
	"encoding"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/samber/lo"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger/log"
)

func (e *eventHandler) OnUserChange(key json.RawMessage, payload payload) error {
	var k UserKey
	if err := json.Unmarshal(key, &k); err != nil {
		e.log.Error("failed to unmarshal json", zap.Error(err))
		return errgo.Wrap(err, "json.Unmarshal")
	}

	switch payload.Op {
	case opCreate, opSnapshot, opDelete:
		return nil
	case opUpdate:
		var before userPayload
		if err := json.Unmarshal(payload.Before, &before); err != nil {
			return errgo.Wrap(err, "json")
		}
		var after userPayload
		if err := json.Unmarshal(payload.After, &after); err != nil {
			return errgo.Wrap(err, "json")
		}

		if before.Password != after.Password {
			err := e.OnUserPasswordChange(k.ID)
			if err != nil {
				e.log.Error("failed to clear cache", zap.Error(err))
			}
		}

		if before.NewNotify != after.NewNotify {
			e.redis.Publish(context.Background(), fmt.Sprintf("event-user-notify-%d", k.ID), redisUserChannel{
				UserID:    k.ID,
				NewNotify: after.NewNotify,
			})
		}

		if before.Avatar != after.Avatar {
			if e.s3 == nil {
				break
			}

			e.log.Debug("clear user avatar cache", log.User(k.ID))
			go e.clearImageCache(after.Avatar)
		}
	}

	return nil
}

func (e *eventHandler) clearImageCache(avatar string) {
	p, q, ok := strings.Cut(avatar, "?")
	if !ok {
		p = avatar
	}

	p = "/pic/user/l/" + p

	if strings.Contains(q, "hd=1") {
		p = "/hd" + p
	}

	e.log.Debug("clear image for prefix", zap.String("avatar", avatar), zap.String("prefix", p))

	err := e.s3.ListObjectsV2PagesWithContext(context.Background(),
		&s3.ListObjectsV2Input{Bucket: &e.config.S3ImageResizeBucket, Prefix: &p},
		func(output *s3.ListObjectsV2Output, b bool) bool {
			_, err := e.s3.DeleteObjects(&s3.DeleteObjectsInput{
				Bucket: &e.config.S3ImageResizeBucket,
				Delete: &s3.Delete{
					Objects: lo.Map(output.Contents, func(item *s3.Object, index int) *s3.ObjectIdentifier {
						return &s3.ObjectIdentifier{
							Key: item.Key,
						}
					}),
				},
			})

			if err != nil {
				e.log.Error("failed to clear s3 cached image", zap.Error(err))
			}

			return true
		},
	)

	if err != nil {
		e.log.Error("failed to clear s3 cached image", zap.Error(err))
	}
}

var _ encoding.BinaryMarshaler = redisUserChannel{}

type redisUserChannel struct {
	UserID    model.UserID `json:"user_id"`
	NewNotify uint16       `json:"new_notify"`
}

func (r redisUserChannel) MarshalBinary() ([]byte, error) {
	return json.Marshal(r) //nolint:wrapcheck
}

type UserKey struct {
	ID model.UserID `json:"uid"`
}

type userPayload struct {
	Password  string `json:"password_crypt"`
	NewNotify uint16 `json:"new_notify"`
	Avatar    string `json:"avatar"`
}
