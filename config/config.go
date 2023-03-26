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
		Gorm bool `yaml:"gorm"`
	} `yaml:"debug"`

	RedisURL      string `yaml:"redis_url" env:"REDIS_URI" env-default:"redis://127.0.0.1:6379/0"`
	RedisPassword string `yaml:"redis_pass" env:"REDIS_PASS" env-default:"redis-pass"`

	MySQLHost     string `yaml:"mysql_host" env:"MYSQL_HOST" env-default:"127.0.0.1"`
	MySQLPort     string `yaml:"mysql_port" env:"MYSQL_PORT" env-default:"3306"`
	MySQLUserName string `yaml:"mysql_user" env:"MYSQL_USER" env-default:"user"`
	MySQLPassword string `yaml:"mysql_pass" env:"MYSQL_PASS" env-default:"password"`
	MySQLDatabase string `yaml:"mysql_db" env:"MYSQL_DB" env-default:"bangumi"`
	MySQLMaxConn  int    `yaml:"mysql_max_connection" env:"MYSQL_MAX_CONNECTION" env-default:"4"`

	WebDomain string `yaml:"web_domain" env:"WEB_DOMAIN"` // new frontend web page domain
	HTTPHost  string `yaml:"http_host" env:"HTTP_HOST" env-default:"127.0.0.1"`
	HTTPPort  int    `yaml:"http_port" env:"HTTP_PORT" env-default:"3000"`

	KafkaBroker      string   `yaml:"kafka_broker" env:"KAFKA_BROKER"`
	KafkaCanalTopics []string `yaml:"kafka_canal_topics"`

	MeiliSearchURL string `yaml:"meilisearch_url" env:"MEILISEARCH_URL"`
	MeiliSearchKey string `yaml:"meilisearch_key" env:"MEILISEARCH_KEY"`

	TurnstileSecretKey string `env:"TURNSTILE_SECRET_KEY" env-default:"1x0000000000000000000000000000000AA"`
	TurnstileSiteKey   string `env:"TURNSTILE_SITE_KEY" env-default:"1x00000000000000000000AA"`

	SearchBatchSize     int           `env:"SEARCH_BATCH_SIZE" yaml:"search_batch_size" env-default:"100"`
	SearchBatchInterval time.Duration `env:"SEARCH_BATCH_INTERVAL" yaml:"search_batch_interval" env-default:"10m"`

	SlowSQLDuration time.Duration `yaml:"slow_sql_duration" env:"SLOW_SQL_DURATION"`

	NsfwWord     string `yaml:"nsfw_word"`
	DisableWords string `yaml:"disable_words"`
	BannedDomain string `yaml:"banned_domain"`

	// "http://localhost:2379"
	EtcdAddr      string `yaml:"etcd_addr" env:"ETCD_ADDR"`
	EtcdNamespace string `yaml:"etcd_namespace" env:"ETCD_NAMESPACE"  env-default:"/chii/services"`

	S3EntryPoint        string `yaml:"s3_entry_point" env:"S3_ENTRY_POINT"`
	S3AccessKey         string `yaml:"s3_access_key" env:"S3_ACCESS_KEY"`
	S3SecretKey         string `yaml:"s3_secret_key" env:"S3_SECRET_KEY"`
	S3ImageResizeBucket string `yaml:"s3_image_resize_bucket" env:"S3_IMAGE_RESIZE_BUCKET" env-default:"img-resize"`

	AppType string
}

func (c AppConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}
