<https://bgm.tv/> 新后端服务器。

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Bangumi/server?style=flat-square)
[![Codecov](https://img.shields.io/codecov/c/github/Bangumi/server?style=flat-square)](https://app.codecov.io/gh/Bangumi/server)

## Requirements

- [Go 1.22](https://go.dev/)
- [go-task](https://taskfile.dev/installation/)，使用 `task` 查看所有的构建目标。
- [golangci-lint](https://golangci-lint.run/)，使用 `task lint` 运行 linter。

## Optional Requirements:

- [buf](https://buf.build/) 基于 ProtoBuffer 生成 grpc 相关文件。
- [mockery](https://github.com/vektra/mockery) 生成 mock
- nodejs: 用于生成 openapi 文件。

## Init

```bash
git clone --recursive https://github.com/bangumi/server.git bangumi-server
cd bangumi-server
task
```

### 设置

可设置的环境变量

- `APP_ENV` 不设置此变量默认为开发环境。在生产环境会被设置为 `production` 或 `stage`
- `MYSQL_HOST` 默认 `127.0.0.1`
- `MYSQL_PORT` 默认 `3306`
- `MYSQL_DB` 默认 `bangumi`
- `MYSQL_USER` 默认 `user`
- `MYSQL_PASS` 默认 `password`
- `REDIS_URI` 默认 `redis://127.0.0.1:6379/0`
- `HTTP_PORT` 默认 `3000`

#### 微服务相关

- `ETCD_ADDR` etcd, 用于微服务的服务发现，留空（默认）的情况下各个微服务会使用 noop mock。example：`http://127.0.0.1:2379`
- `ETCD_NAMESPACE` etcd 服务注册的 key 前缀。大多数情况下不需要设置。

https://github.com/bangumi/service-timeline

搜索功能相关的环境变量

- `MEILISEARCH_URL` meilisearch 地址，默认为空。不设置的话不会初始化搜索客户端。
- `MEILISEARCH_KEY` meilisearch key。
- `KAFKA_BROKER` kafka broker 地址，启动 canal 需要 kafka。

你也可以把配置放在 `.env` 文件中，`go-task` 会自动加载 `.env` 文件中的环境变量。

example:

```text
MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3306"
MYSQL_USER="user"
MYSQL_PASS="password"
MYSQL_DB="bangumi"
REDIS_URI="redis://:redis-pass@127.0.0.1:6379/1"
MEILISEARCH_URL="http://127.0.0.1:7700/"
```

或者使用 yaml 格式的配置文件：

查看 config.example.yaml 或者 [config/config.go](https://github.com/bangumi/server/blob/master/config/config.go)

## 开发

bangumi 相关项目整体说明 [bangumi/dev-docs](https://github.com/bangumi/dev-docs)

Web 框架: [echo](https://echo.labstack.com/)

ORM: [GORM](https://github.com/go-gorm/gorm) 和 [GORM Gen](https://github.com/go-gorm/gen)

启动 HTTP server

```shell
go run main.go --config config.yaml web
```

启动 kafka consumer

```shell
go run main.go canal --config config.yaml
```

### 后端环境

redis 和 mysql 都在此 docker-compose 内 <https://github.com/bangumi/dev-env> 。

如果你不使用 docker ，请自行启动 mysql 和 redis 并导入 `bangumi/dev-env` 仓库内的数据。

### 依赖的微服务

- 时间胶囊 https://github.com/bangumi/service-timeline

### 贡献指南

更多的细节介绍请查看[贡献指南](./.github/contributing.md)。

## 提交 Pull Request

如果你的 PR 是新功能，最好先发个 issue 讨论一下要不要实现，避免 PR 提交之后新功能本身被否决的问题。

如果已经存在相关的 issue，可以先在 issue 内回复一下自己的意向，或者创建一个 Draft PR 关联对应的 issue，避免撞车问题。

PR 标题需要遵守 [conventional commits 风格](https://www.conventionalcommits.org/en/v1.0.0/)。

## 测试

使用 [mock](./internal/mocks/) 进行部分测试。

```
task test       # Run mocked unit tests, need nothing.
task test-db    # Run mocked tests, and tests require mysql and redis. alias for `TEST_MYSQL=1 TEST_REDIS=1 task test`
task test-all   # Run all tests.
```

## 代码风格

```bash
task lint
```

### 配置文件

非 golang 文件 (yaml, json, markdown 等) 使用 [prettier](https://prettier.io/) 进行格式化。
如果你没有 nodejs 环境，也可以直接提交 PR。GitHub Actions 会对文件进行检查并提出修改建议。

## Go Mod

你不应当导入 `github.com/bangumi/server/pkg` 以外的任何路径。

具体可用的包见 [pkg/readme.md](./pkg)

## License

Copyright (C) 2021-2022 bangumi server contributors.

Source is licensed under the GNU AGPLv3 license that can be found in
the [LICENSE.txt](https://github.com/bangumi/server/blob/master/LICENSE.txt) file.
