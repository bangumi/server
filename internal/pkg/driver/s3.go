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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/config"
)

func NewS3(c config.AppConfig) (*minio.Client, error) {
	if c.S3EntryPoint == "" {
		return nil, nil //nolint:nilnil
	}

	// Initialize minio client object.
	minioClient, err := minio.New(c.S3EntryPoint, &minio.Options{
		Creds: credentials.NewStaticV4(c.S3AccessKey, c.S3SecretKey, ""),
	})
	if err != nil {
		return nil, errgo.Wrap(err, "s3: failed to connect to s3")
	}

	return minioClient, nil
}
