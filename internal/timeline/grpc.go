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

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal/query"
	pb "github.com/bangumi/server/generated/proto/go/api/v1"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger, cfg config.AppConfig) (Repo, error) {
	rpc, err := newGrpcClient(cfg)
	if err != nil {
		return nil, err
	}

	return mysqlRepo{q: q, log: log.Named("timeline.mysqlRepo"), rpc: rpc}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
	rpc pb.TimeLineServiceClient
}

func newGrpcClient(cfg config.AppConfig) (pb.TimeLineServiceClient, error) {
	if cfg.MicroServiceTimelineAddr == "" {
		logger.Info("no etcd, using nope timeline service")
		return noopClient{}, nil
	}

	logger.Info("using timeline service " + cfg.MicroServiceTimelineAddr)

	conn, err := grpc.Dial(
		cfg.MicroServiceTimelineAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errgo.Wrap(err, "grpc dail")
	}

	c := pb.NewTimeLineServiceClient(conn)

	return c, nil
}

var _ pb.TimeLineServiceClient = noopClient{}

type noopClient struct {
}

func (n noopClient) Hello(_ context.Context, _ *pb.HelloRequest, _ ...grpc.CallOption) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{}, nil
}

func (n noopClient) SubjectCollect(ctx context.Context, in *pb.SubjectCollectRequest,
	opts ...grpc.CallOption) (*pb.SubjectCollectResponse, error) {
	return &pb.SubjectCollectResponse{Ok: true}, nil
}

func (n noopClient) SubjectProgress(ctx context.Context, in *pb.SubjectProgressRequest,
	opts ...grpc.CallOption) (*pb.SubjectProgressResponse, error) {
	return &pb.SubjectProgressResponse{Ok: true}, nil
}

func (n noopClient) EpisodeCollect(ctx context.Context, in *pb.EpisodeCollectRequest,
	opts ...grpc.CallOption) (*pb.EpisodeCollectResponse, error) {
	return &pb.EpisodeCollectResponse{Ok: true}, nil
}

func (m mysqlRepo) ChangeSubjectCollection(
	ctx context.Context,
	u auth.Auth,
	sbj model.Subject,
	collect model.SubjectCollection,
	comment string,
	rate uint8,
) error {
	_, err := m.rpc.SubjectCollect(ctx, &pb.SubjectCollectRequest{
		UserId: uint64(u.ID),
		Subject: &pb.Subject{
			Id:        uint32(sbj.ID),
			Type:      uint32(sbj.TypeID),
			Name:      sbj.Name,
			NameCn:    sbj.NameCN,
			Image:     sbj.Image,
			Series:    false,
			VolsTotal: sbj.Volumes,
			EpsTotal:  sbj.Eps,
		},
		Collection: uint32(collect),
		Comment:    comment,
		Rate:       uint32(rate),
	})

	if err != nil {
		return errgo.Wrap(err, "grpc: timeline.SubjectCollect")
	}

	return nil
}
func (m mysqlRepo) ChangeEpisodeStatus(
	ctx context.Context,
	u auth.Auth,
	sbj model.Subject,
	episode episode.Episode,
	update collection.Update,
) error {
	// TODO
	return nil
}
