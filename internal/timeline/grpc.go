/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package timeline

import (
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bangumi/server/config"
	pb "github.com/bangumi/server/generated/proto/go/api/v1"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func newGrpcClient(cfg config.AppConfig) (pb.TimeLineServiceClient, error) {
	cli, err := clientv3.NewFromURL(cfg.EtcdAddr)
	if err != nil {
		return nil, errgo.Wrap(err, "new etcd client")
	}

	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		return nil, errgo.Wrap(err, "new etcd resolver")
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("etcd://%s/timeline.v1", cfg.EtcdServiceNamespace),
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errgo.Wrap(err, "grpc dail")
	}

	c := pb.NewTimeLineServiceClient(conn)

	return c, nil
}
