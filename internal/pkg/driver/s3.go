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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/samber/lo"

	"github.com/bangumi/server/config"
)

func NewS3(c config.AppConfig) (*s3.S3, error) {
	if c.S3EntryPoint == "" {
		return nil, nil //nolint:nilnil
	}

	cred := credentials.NewStaticCredentials(c.S3AccessKey, c.S3SecretKey, "")
	s := lo.Must(session.NewSession(&aws.Config{
		Credentials:      cred,
		Endpoint:         &c.S3EntryPoint,
		Region:           lo.ToPtr("us-east-1"),
		DisableSSL:       lo.ToPtr(true),
		S3ForcePathStyle: lo.ToPtr(true),
	}))

	svc := s3.New(s)

	return svc, nil
}
