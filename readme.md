基于 python 的新 api server

## 开发环境

python 版本: 3.8

依赖管理: [poetry](https://github.com/python-poetry/poetry)

web 框架: [fastapi](https://github.com/tiangolo/fastapi)

quick start:

```shell
git clone https://github.com/bangumi/server bangumi-server
cd bangumi-server
python -m venv .venv # MUST use python 3.8
source .venv/bin/activate
poetry install --remove-untracked
pre-commit install
```

### 设置

可设置的环境变量

- `MYSQL_HOST` 默认 `127.0.0.1`
- `MYSQL_PORT` 默认 `3306`
- `MYSQL_DB` 默认 `bangumi`
- `MYSQL_USER` **无默认值**
- `MYSQL_PASS` **无默认值**

启动服务器

```shell
uvicorn pol.server:app --reload --port 3000
```

### 后端环境

https://github.com/bangumi/dev-env

## 测试(需要数据库)

```shell
pytest
```

## 代码风格

以 LF 为换行符

### python

启用 [pre-commit](https://github.com/pre-commit/pre-commit)

```shell
pre-commit insall
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
