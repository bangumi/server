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
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
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

const defaultTimeout = time.Second * 5

func NewMysqlRepo(q *query.Query, log *zap.Logger, cfg config.AppConfig) (Service, error) {
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

func (m mysqlRepo) ChangeSubjectProgress(ctx context.Context, u model.UserID, sbj model.Subject,
	epsUpdate uint32, volsUpdate uint32) error {
	ctx, canal := context.WithTimeout(ctx, defaultTimeout)
	defer canal()

	_, err := m.rpc.SubjectProgress(ctx, &pb.SubjectProgressRequest{
		UserId: uint64(u),
		Subject: &pb.Subject{
			Id:        sbj.ID,
			Type:      uint32(sbj.TypeID),
			Name:      sbj.Name,
			NameCn:    sbj.NameCN,
			Image:     sbj.Image,
			Series:    sbj.Series,
			VolsTotal: sbj.Volumes,
			EpsTotal:  sbj.Eps,
		},
		EpsUpdate:  epsUpdate,
		VolsUpdate: volsUpdate,
	})

	return errgo.Wrap(err, "grpc")
}

func (m mysqlRepo) ChangeSubjectCollection(
	ctx context.Context,
	u model.UserID,
	sbj model.Subject,
	collect collection.SubjectCollection,
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
) error {
	ctx, canal := context.WithTimeout(ctx, defaultTimeout)
	defer canal()

	_, err := m.rpc.EpisodeCollect(ctx, &pb.EpisodeCollectRequest{
		UserId: uint64(u.ID),
		Last: &pb.Episode{
			Id:     episode.ID,
			Type:   uint32(episode.Type),
			Name:   episode.Name,
			NameCn: episode.NameCN,
			Sort:   float64(episode.Sort),
		},
		Subject: &pb.Subject{
			Id:        sbj.ID,
			Type:      uint32(sbj.TypeID),
			Name:      sbj.Name,
			NameCn:    sbj.Name,
			Image:     sbj.Image,
			Series:    sbj.Series,
			VolsTotal: sbj.Volumes,
			EpsTotal:  sbj.Eps,
		},
	})

	return errgo.Wrap(err, "grpc")
}

func newGrpcClient(cfg config.AppConfig) (pb.TimeLineServiceClient, error) {
	if cfg.EtcdAddr == "" {
		logger.Info("no etcd, using nope timeline service")
		return noopClient{}, nil
	}

	logger.Info("using etcd to discovery timeline services " + cfg.EtcdAddr)

	cli, err := clientv3.NewFromURL(cfg.EtcdAddr)
	if err != nil {
		return nil, errgo.Wrap(err, "etcd new client")
	}

	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		return nil, errgo.Wrap(err, "etcd grpc resolver")
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("etcd:///%s/timeline", cfg.EtcdNamespace),
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errgo.Wrap(err, "grpc dail")
	}

	c := pb.NewTimeLineServiceClient(conn)

	return c, nil
}
