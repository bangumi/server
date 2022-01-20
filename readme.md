基于 python 的新后端服务器

## 开发环境

python 版本: 3.8

使用 [poetry](https://github.com/python-poetry/poetry) 进行依赖管理。

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

bangumi 相关项目整体说明 [bangumi/dev-docs](https://github.com/bangumi/dev-docs)

### 项目结构

Web 框架 [fastapi](https://github.com/tiangolo/fastapi)

ORM 类定义在 [pol/db/tables.py](./pol/db/tables.py) 文件。

路由位于 [pol/api](./pol/api) 文件夹。

### 后端环境

redis 和 mysql 都在此 docker-compose 内 <https://github.com/bangumi/dev-env> 。

如果你不使用 docker ，请自行启动 mysql 和 redis 并导入 `bangumi/dev-env` 仓库内的数据。

### 运行服务器

默认在 `3000` 端口，在代码修改后会自动重启。

```shell
watchgod scripts.dev.main
```

访问 <http://127.0.0.1:3000/v0/>。

## 提交 Pull Request

如果你的 PR 是新功能，最好先发个 issue 讨论一下要不要实现，避免 PR 提交之后新功能本身被否决的问题。

如果已经存在相关的 issue，可以先在 issue 内回复一下自己的意向，或者创建一个 Draft PR 关联对应的 issue，避免撞车问题。

## [测试](./tests/readme.md)

## 代码风格

### python

基本上，代码格式基于 black 和 isort。

#### Formatter

我们使用 [pre-commit](https://github.com/pre-commit/pre-commit) 来管理代码格式化工具。

你可以手动运行

```shell
pre-commit run # only check changed files
pre-commit run --all-files # check all files
```

也可以安装 git hooks，在每次 commit 时自动运行。

```shell
pre-commit install
```

绝大多数情况下，你不需要因此而手动修改代码，仅需要保证相关格式化工具正常运行即可，

#### Lint

CI 中还会运行 flake8 和 mypy

### 配置文件

非 python 文件(yaml, json, markdown 等)使用 [prettier](https://prettier.io/) 进行格式化。

## License

BSD 3-Clause License

[LICENSE](https://github.com/bangumi/server/blob/master/LICENSE.md)
