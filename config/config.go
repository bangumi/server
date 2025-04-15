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
		Gorm bool `yaml:"gorm" toml:"gorm"`
	} `yaml:"debug" toml:"debug"`

	RedisURL string `toml:"redis-url" yaml:"redis_url" env:"REDIS_URI" env-default:"redis://127.0.0.1:6379/0"`

	Mysql struct {
		Host        string        `toml:"host" yaml:"host" env:"MYSQL_HOST" env-default:"127.0.0.1"`
		Port        string        `toml:"port" yaml:"port" env:"MYSQL_PORT" env-default:"3306"`
		UserName    string        `toml:"user-name" yaml:"user" env:"MYSQL_USER" env-default:"user"`
		Password    string        `toml:"password" yaml:"password" env:"MYSQL_PASS" env-default:"password"`
		Database    string        `toml:"database" yaml:"db" env:"MYSQL_DB" env-default:"bangumi"`
		MaxConn     int           `toml:"max-conn" yaml:"max_connection" env:"MYSQL_MAX_CONNECTION" env-default:"4"`
		MaxIdleTime time.Duration `toml:"max-idle-time" yaml:"conn_max_idle_time" env-default:"4h"`
		MaxLifeTime time.Duration `toml:"max-life-time" yaml:"conn_max_life_time" env-default:"6h"`

		SlowSQLDuration time.Duration `toml:"slow-sql-duration" yaml:"slow_sql_duration" env:"SLOW_SQL_DURATION"`
	} `yaml:"mysql" toml:"mysql"`

	HTTP struct {
		Host string `yaml:"host" toml:"host" env:"HTTP_HOST" env-default:"127.0.0.1"`
		Port int    `yaml:"port" toml:"port" env:"HTTP_PORT" env-default:"3000"`
	} `toml:"http" yaml:"http"`

	RateLimit struct {
		LimitLongTime time.Duration `yaml:"long_time" toml:"long-time" env:"RATE_LIMIT_LONG_TIME" env-default:"1h"`
		LimitWindow   time.Duration `yaml:"window" toml:"window" env:"RATE_LIMIT_WINDOW" env-default:"10m"`
		LimitCount    uint          `yaml:"count" toml:"count" env:"RATE_LIMIT_COUNT" env-default:"3000"`
	} `yaml:"rate_limit" toml:"rate-limit"`

	Canal struct {
		Broker string `yaml:"broker" toml:"broker"`

		KafkaBroker string   `yaml:"kafka_broker" toml:"kafka-broker" env:"KAFKA_BROKER"`
		Topics      []string `yaml:"topics" toml:"topics"`
	} `yaml:"canal" toml:"canal"`

	Search struct {
		MeiliSearch struct {
			URL     string        `yaml:"url" toml:"url" env:"MEILISEARCH_URL"`
			Key     string        `yaml:"key" toml:"key" env:"MEILISEARCH_KEY"`
			Timeout time.Duration `yaml:"timeout" toml:"timeout" env:"MEILISEARCH_REQUEST_TIMEOUT" env-default:"2s"`
		} `yaml:"meilisearch" toml:"meilisearch"`

		SearchBatchSize     int           `env:"SEARCH_BATCH_SIZE" yaml:"batch_size" toml:"batch-size" env-default:"100"`
		SearchBatchInterval time.Duration `env:"SEARCH_BATCH_INTERVAL" yaml:"batch_interval" toml:"batch-interval" env-default:"10m"`
	} `yaml:"search" toml:"search"`

	NsfwWord     string `yaml:"nsfw_word" toml:"nsfw-word"`
	DisableWords string `yaml:"disable_words" toml:"disable-words"`
	BannedDomain string `yaml:"banned_domain" toml:"banned-domain"`

	S3EntryPoint        string `yaml:"s3_entry_point" toml:"s3-entry-point" env:"S3_ENTRY_POINT"`
	S3AccessKey         string `yaml:"s3_access_key" toml:"s3-access-key" env:"S3_ACCESS_KEY"`
	S3SecretKey         string `yaml:"s3_secret_key" toml:"s3-secret-key" env:"S3_SECRET_KEY"`
	S3ImageResizeBucket string `yaml:"s3_image_resize_bucket" toml:"s3-image-resize-bucket" env:"S3_IMAGE_RESIZE_BUCKET" env-default:"img-resize"`

	AppType string `toml:"app-type"`
}

func (c AppConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTP.Host, c.HTTP.Port)
}
