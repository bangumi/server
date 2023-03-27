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

	"google.golang.org/grpc"

	pb "github.com/bangumi/server/generated/proto/go/api/v1"
)

var _ pb.TimeLineServiceClient = noopClient{}

type noopClient struct {
}

func (n noopClient) Hello(_ context.Context, _ *pb.HelloRequest, _ ...grpc.CallOption) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{}, nil
}

func (n noopClient) SubjectCollect(_ context.Context, _ *pb.SubjectCollectRequest,
	_ ...grpc.CallOption) (*pb.SubjectCollectResponse, error) {
	return &pb.SubjectCollectResponse{Ok: true}, nil
}

func (n noopClient) SubjectProgress(_ context.Context, _ *pb.SubjectProgressRequest,
	_ ...grpc.CallOption) (*pb.SubjectProgressResponse, error) {
	return &pb.SubjectProgressResponse{Ok: true}, nil
}

func (n noopClient) EpisodeCollect(_ context.Context, _ *pb.EpisodeCollectRequest,
	_ ...grpc.CallOption) (*pb.EpisodeCollectResponse, error) {
	return &pb.EpisodeCollectResponse{Ok: true}, nil
}
