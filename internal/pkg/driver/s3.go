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

package driver

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	transport "github.com/aws/smithy-go/endpoints"
	"github.com/samber/lo"

	"github.com/bangumi/server/config"
)

type resolver struct {
	URL *url.URL
}

func (r *resolver) ResolveEndpoint(_ context.Context, params s3.EndpointParameters) (transport.Endpoint, error) {
	u := *r.URL
	u.Path += "/" + *params.Bucket
	return transport.Endpoint{URI: u}, nil
}

func NewS3(c config.AppConfig) (*s3.Client, error) {
	if c.S3EntryPoint == "" {
		return nil, nil //nolint:nilnil
	}

	svc := s3.New(s3.Options{
		EndpointResolverV2: &resolver{URL: lo.Must(url.Parse(c.S3EntryPoint))},
		Region:             "us-east-1",
		UsePathStyle:       true,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     c.S3AccessKey,
				SecretAccessKey: c.S3SecretKey,
			}, nil
		}),
	})

	return svc, nil
}
