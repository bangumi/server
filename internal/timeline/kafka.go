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

package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/segmentio/kafka-go"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
)

const timelineSourceAPI = 5
const timelineTopic = "timeline"
const defaultTimeout = time.Second * 5

func NewSrv(kafka *kafka.Writer) (Service, error) {
	return kafkaClient{kafka: kafka}, nil
}

type kafkaClient struct {
	kafka *kafka.Writer
}

func (m kafkaClient) ChangeSubjectProgress(ctx context.Context, u model.UserID, sbj model.Subject,
	epsUpdate uint32, volsUpdate uint32) error {
	ctx, canal := context.WithTimeout(ctx, defaultTimeout)
	defer canal()

	return m.writeMessage(ctx, u, timelineValue{
		Op: "progressSubject",
		Message: progressSubject{
			UID:       u,
			CreatedAt: time.Now().Unix(),
			Source:    timelineSourceAPI,
			Subject: tlSubjectCollect{
				ID:      int(sbj.ID),
				Type:    sbj.TypeID,
				Eps:     int(sbj.Eps),
				Volumes: int(sbj.Volumes),
			},
			Collect: tlCollect{
				EpsUpdate:  &epsUpdate,
				VolsUpdate: &volsUpdate,
			},
		},
	})
}

func (m kafkaClient) ChangeSubjectCollection(
	ctx context.Context,
	u model.UserID,
	sbj model.Subject,
	collect collection.SubjectCollection,
	collectID uint64,
	comment string,
	rate uint8,
) error {
	ctx, canal := context.WithTimeout(ctx, defaultTimeout)
	defer canal()

	return m.writeMessage(ctx, u, timelineValue{
		Op: "subject",
		Message: subject{
			UID:       u,
			CreatedAt: time.Now().Unix(),
			Source:    timelineSourceAPI,
			Subject: tlSubject{
				ID:   sbj.ID,
				Type: sbj.TypeID,
			},
			Collect: tlCollectRating{
				ID:      collectID,
				Type:    collect,
				Rate:    rate,
				Comment: comment,
			},
		},
	})
}

func (m kafkaClient) ChangeEpisodeStatus(
	ctx context.Context,
	u auth.Auth,
	sbj model.Subject,
	episode episode.Episode,
	t collection.EpisodeCollection,
) error {
	ctx, canal := context.WithTimeout(ctx, defaultTimeout)
	defer canal()

	return m.writeMessage(
		ctx,
		u.ID,
		timelineValue{
			Op: "progressEpisode",
			Message: progressEpisode{
				UID: u.ID,
				Subject: tlSubject{
					ID:   sbj.ID,
					Type: sbj.TypeID,
				},
				Episode: tlEpisode{
					ID:     episode.ID,
					Status: t,
				},
				CreatedAt: time.Now().Unix(),
				Source:    timelineSourceAPI,
			},
		},
	)
}

func (m kafkaClient) writeMessage(ctx context.Context, uid model.UserID, value timelineValue) error {
	err := m.kafka.WriteMessages(ctx, kafka.Message{
		Topic: timelineTopic,
		Key:   fmt.Appendf(nil, "%d", uid),
		Value: lo.Must(json.Marshal(value)),
	})

	return errgo.Wrap(err, "kafka")
}
