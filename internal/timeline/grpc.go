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
	"fmt"
	"time"

	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal/query"
	pb "github.com/bangumi/server/generated/proto/go/api/v1"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/logger"
)

const defaultTimeout = time.Second * 30

func NewMysqlRepo(q *query.Query, log *zap.Logger, cfg config.AppConfig) (Service, error) {
	rpc, err := newGrpcClient(cfg)
	if err != nil {
		return nil, err
	}

	return grpcClient{q: q, log: log.Named("timeline.grpcClient"), rpc: rpc}, nil
}

type grpcClient struct {
	q   *query.Query
	log *zap.Logger
	rpc pb.TimeLineServiceClient
}

func (m grpcClient) ChangeSubjectProgress(ctx context.Context, u model.UserID, sbj model.Subject,
	epsUpdate uint32, volsUpdate uint32) error {
	ctx, canal := context.WithTimeout(ctx, defaultTimeout)
	defer canal()

	_, err := m.rpc.SubjectProgress(ctx, &pb.SubjectProgressRequest{
		UserId: uint64(u),
		Subject: &pb.Subject{
			Id:        sbj.ID,
			VolsTotal: sbj.Volumes,
			EpsTotal:  sbj.Eps,
		},
		EpsUpdate:  epsUpdate,
		VolsUpdate: volsUpdate,
	})

	return errgo.Wrap(err, "grpc")
}

func (m grpcClient) ChangeSubjectCollection(
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

	_, err := m.rpc.SubjectCollect(ctx, &pb.SubjectCollectRequest{
		UserId: uint64(u),
		Subject: &pb.Subject{
			Id:        sbj.ID,
			Type:      uint32(sbj.TypeID),
			VolsTotal: sbj.Volumes,
			EpsTotal:  sbj.Eps,
		},
		Collection:   uint32(collect),
		CollectionId: collectID,
		Comment:      comment,
		Rate:         uint32(rate),
	})

	if err != nil {
		return errgo.Wrap(err, "grpc: timeline.SubjectCollect")
	}

	return nil
}

func (m grpcClient) ChangeEpisodeStatus(
	ctx context.Context,
	u auth.Auth,
	sbj model.Subject,
	episode episode.Episode,
) error {
	ctx, canal := context.WithTimeout(ctx, defaultTimeout)
	defer canal()

	_, err := m.rpc.EpisodeCollect(ctx, &pb.EpisodeCollectRequest{
		UserId: uint64(u.ID),
		Last: &pb.Episode{
			Id: episode.ID,
		},
		Subject: &pb.Subject{
			Id:        sbj.ID,
			VolsTotal: sbj.Volumes,
			EpsTotal:  sbj.Eps,
		},
	})

	return errgo.Wrap(err, "grpc")
}

func newGrpcClient(cfg config.AppConfig) (pb.TimeLineServiceClient, error) {
	if cfg.SrvTimelineDomain == "" || cfg.SrvTimelinePort == 0 {
		logger.Info("no srv_timeline_domain and srv_timeline_port, using nope timeline service")
		return noopClient{}, nil
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("dns:///%s:%d", cfg.SrvTimelineDomain, cfg.SrvTimelinePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errgo.Wrap(err, "grpc dail")
	}

	c := pb.NewTimeLineServiceClient(conn)

	return c, nil
}
