# 测试

测试基于 [Pytest](https://docs.pytest.org/en/stable/)

同时可以查看 [FastAPI 的测试文档](https://fastapi.tiangolo.com/tutorial/testing/)

## 运行

```shell
pytest --e2e --database --redis
```

默认的 `pytest` 命令仅会运行一些简单地单元测试，用于筛选测试的 flag 包括:

- `--e2e` 允许 e2e 测试。
- `--database` 允许需要 mysql 的测试。
- `--redis` 允许需要 redis 的测试。

如果一个测试同时需要多个 flag，如 mysql 和 redis，需要同时提供 `--darabase` 和 `--redis` 选项参数才会运行。

## 编写测试

简单地单元测试（如 `wiki_test` 和 `util_test`）在 `pol` 文件夹中，以 `${filename}_test.py` 命名。

其他更加复杂的测试在`tests/`文件夹中。

[tests/conftest.py](./conftest.py) 中提供了一些测试工具，可以用于 mock fastapi 的 depends 或者数据库数据，可以参照 [./tests/app/api_v0/test_subject.py](./app/api_v0/test_subject.py) 文件进行 mock。

`@pytest.mark.env(...)` 用于标记测试所需要的 flag，所需要的 flag 在`conftest.py`的 `pytest_addoption`函数中。

如果你还需要额外的测试数据，也可以在 `dev-env` 仓库中提交 PR 进行添加。
