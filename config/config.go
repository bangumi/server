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

package config

import (
	"fmt"
	"time"
)

const AppTypeCanal = "canal"
const AppTypeHTTP = "http"

type AppConfig struct {
	Debug struct {
		Gorm bool `toml:"gorm"`
	} `toml:"debug"`

	RedisURL string `toml:"redis-url" env:"REDIS_URI" env-default:"redis://127.0.0.1:6379/0"`

	Mysql struct {
		Host        string        `toml:"host" env:"MYSQL_HOST" env-default:"127.0.0.1"`
		Port        string        `toml:"port" env:"MYSQL_PORT" env-default:"3306"`
		UserName    string        `toml:"user-name" env:"MYSQL_USER" env-default:"user"`
		Password    string        `toml:"password" env:"MYSQL_PASS" env-default:"password"`
		Database    string        `toml:"database" env:"MYSQL_DB" env-default:"bangumi"`
		MaxConn     int           `toml:"max-conn" env:"MYSQL_MAX_CONNECTION" env-default:"4"`
		MaxIdleTime time.Duration `toml:"max-idle-time" env-default:"4h"`
		MaxLifeTime time.Duration `toml:"max-life-time" env-default:"6h"`

		SlowSQLDuration time.Duration `toml:"slow-sql-duration" env:"SLOW_SQL_DURATION"`
	} `toml:"mysql"`

	HTTP struct {
		Host string `toml:"host" env:"HTTP_HOST" env-default:"127.0.0.1"`
		Port int    `toml:"port" env:"HTTP_PORT" env-default:"3000"`
	} `toml:"http"`

	RateLimit struct {
		LimitLongTime time.Duration `toml:"long-time" env:"RATE_LIMIT_LONG_TIME" env-default:"1h"`
		LimitWindow   time.Duration `toml:"window" env:"RATE_LIMIT_WINDOW" env-default:"10m"`
		LimitCount    uint          `toml:"count" env:"RATE_LIMIT_COUNT" env-default:"3000"`
	} `toml:"rate-limit"`

	Kafka struct {
		Broker string   `toml:"broker" env:"KAFKA_BROKER"`
		Topics []string `toml:"topics"`
	} `toml:"kafka"`

	Search struct {
		MeiliSearch struct {
			URL     string        `toml:"url" env:"MEILISEARCH_URL"`
			Key     string        `toml:"key" env:"MEILISEARCH_KEY"`
			Timeout time.Duration `toml:"timeout" env:"MEILISEARCH_REQUEST_TIMEOUT" env-default:"2s"`
		} `toml:"meilisearch"`
	} `toml:"search"`

	NsfwWord     string `toml:"nsfw-word"`
	DisableWords string `toml:"disable-words"`
	BannedDomain string `toml:"banned-domain"`

	S3EntryPoint        string `toml:"s3-entry-point" env:"S3_ENTRY_POINT"`
	S3AccessKey         string `toml:"s3-access-key" env:"S3_ACCESS_KEY"`
	S3SecretKey         string `toml:"s3-secret-key" env:"S3_SECRET_KEY"`
	S3ImageResizeBucket string `toml:"s3-image-resize-bucket" env:"S3_IMAGE_RESIZE_BUCKET" env-default:"img-resize"`

	AppType string `toml:"app-type"`
}

func (c AppConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTP.Host, c.HTTP.Port)
}
