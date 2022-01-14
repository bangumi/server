基于 python3.8 的新 api server

## 开发环境

python 版本: 3.8

使用[poetry](https://github.com/python-poetry/poetry)进行依赖管理。

```shell
git clone https://github.com/bangumi/server bangumi-server
cd bangumi-server
```

进入虚拟环境

```shell
python -m venv .venv # MUST use python 3.8
source .venv/bin/activate # enable virtualenv
```

安装依赖

```shell
poetry install --remove-untracked
```

安装 git hook

```shell
pre-commit install
```

### 设置

可设置的环境变量

- `MYSQL_HOST` 默认 `127.0.0.1`
- `MYSQL_PORT` 默认 `3306`
- `MYSQL_DB` 默认 `bangumi`
- `MYSQL_USER` 默认 `user`
- `MYSQL_PASS` 默认 `password`
- `REDIS_URI` 默认 `redis://127.0.0.1:6379/0`

你也可以把配置放在 `./env/dev` 文件中，如果在环境变量中找不到对应的值，会在这个文件中查找

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

相关项目整体说明 [bangumi/dev-docs](https://github.com/bangumi/dev-docs)

### 项目结构

Web 框架 [fastapi](https://github.com/tiangolo/fastapi)

ORM 类定义在 [pol/db/tables.py](./pol/db/tables.py) 文件。

路由位于 [pol/api](./pol/api) 文件夹。

### 后端环境

redis 和 mysql 都在此 docker-compose 内。 https://github.com/bangumi/dev-env

如果你不是 docker 用户，请自行启动 mysql 和 redis 并导入`bangumi/dev-env` 仓库内的数据。

启动 web 服务器，默认为 `3000` 端口，在代码修改后会自动重启。

```shell
watchgod scripts.dev.main
```

## 提交 Pull Request

如果你的 PR 是新功能，最好先发个 issue 讨论一下要不要实现，避免 PR 提交之后新功能本身被否决的问题。

如果已经存在相关的 issue，最好先在 issue 内回复一下自己的意向，或者创建一个 Draft PR 关联对应的 issue，避免撞车问题。

## 测试

测试基于 pytest

### 运行测试(需要数据库)

```shell
pytest --e2e --database --redis
```

默认的 `pytest` 命令仅会运行一些简单地单元测试，其他 flag 包括:

- `--e2e` 允许 e2e 测试。
- `--database` 允许需要 mysql 的测试。
- `--redis` 允许需要 redis 的测试。

如果一个测试同时需要 mysql 和 redis，需要同时提供 `--darabase` 和 `--redis` 选项参数才会运行。

### 编写测试

参照 [tests/app/test_base_router.py](./tests/app/test_base_router.py) 文件。在测试函数中添加`client`
参数获取对应的 HTTP 测试客户端。`client` 是一个 `requests.Session` 的实例，可以使用 `requests` 的各种函数参数。

[详细文档](https://www.starlette.io/testclient/)

## 代码风格

以 LF 为换行符

### python

启用 [pre-commit](https://github.com/pre-commit/pre-commit)

```shell
pre-commit install
```

pre-commit 会在当前仓库安装一个 git hook，在每次 commit 前自动运行。

也可以手动运行

```shell
pre-commit run #only check changed files
pre-commit run --all-files # check all files
```

lint: flake8

### 配置文件

非 python 文件(yaml, json, markdown 等)使用 [prettier](https://prettier.io/) 进行格式化。

## pol

pol 来源于我的旧项目名，没有特殊含义。

## License

BSD 3-Clause License

[LICENSE](https://github.com/bangumi/server/blob/master/LICENSE.md)
