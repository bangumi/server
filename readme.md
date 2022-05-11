<https://bgm.tv/> 新后端服务器。

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Bangumi/server?style=flat-square)
[![Codecov](https://img.shields.io/codecov/c/github/Bangumi/server?style=flat-square)](https://app.codecov.io/gh/Bangumi/server)

## Requirements

- [Go 1.17+](https://go.dev/)
- [GNU make](https://www.gnu.org/software/make/)

## Optional Requirements:

- nodejs and npm: 用于生成 openapi 文件。
- [mockery](https://github.com/vektra/mockery#installation): 用于生成测试用的 mock 文件。

## Init

```bash
git clone --recursive https://github.com/bangumi/server.git bangumi-server
cd bangumi-server
make install
```

### 设置

可设置的环境变量

- `MYSQL_HOST` 默认 `127.0.0.1`
- `MYSQL_PORT` 默认 `3306`
- `MYSQL_DB` 默认 `bangumi`
- `MYSQL_USER` 默认 `user`
- `MYSQL_PASS` 默认 `password`
- `REDIS_URI` 默认 `redis://127.0.0.1:6379/0`

你也可以把配置放在 `.env` 文件中。

example:

```text
MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3306"
MYSQL_USER="user"
MYSQL_PASS="password"
MYSQL_DB="bangumi"
REDIS_URI="redis://:redis-pass@127.0.0.1:6379/1"
```

## 开发

bangumi 相关项目整体说明 [bangumi/dev-docs](https://github.com/bangumi/dev-docs)

Web 框架: [fiber](https://github.com/gofiber/fiber)

ORM: [GORM](https://github.com/go-gorm/gorm) 和 [GORM Gen](https://github.com/go-gorm/gen)

### 后端环境

redis 和 mysql 都在此 docker-compose 内 <https://github.com/bangumi/dev-env> 。

如果你不使用 docker ，请自行启动 mysql 和 redis 并导入 `bangumi/dev-env` 仓库内的数据。

## 提交 Pull Request

如果你的 PR 是新功能，最好先发个 issue 讨论一下要不要实现，避免 PR 提交之后新功能本身被否决的问题。

如果已经存在相关的 issue，可以先在 issue 内回复一下自己的意向，或者创建一个 Draft PR 关联对应的 issue，避免撞车问题。

## 测试

运行部分测试。

```
make test
```

运行全部测试，需要数据库环境。

```
make test-all
```

## 代码风格

```bash
make lint
```

### 配置文件

非 golang 文件(yaml, json, markdown 等)使用 [prettier](https://prettier.io/) 进行格式化。

## Go Mod

你不应当导入 `github.com/bangumi/server/pkg` 以外的任何路径。

具体可用的包见 [pkg/readme.md](./pkg)

## License

Source is licensed under the GNU AGPLv3 license that can be found in the [LICENSE.txt](https://github.com/bangumi/server/blob/master/LICENSE.txt) file.
