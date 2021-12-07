开发工具:

- go 1.17+
- GNU make

## 设置

可以设置以下环境变量:

- `MYSQL_HOST` 默认 `127.0.0.1`
- `MYSQL_PORT` 默认 `3306`
- `MYSQL_DB` 默认 `bangumi`
- `MYSQL_USER` **无默认值，必需设置**
- `MYSQL_PASS` **无默认值，必需设置**
- `DB_DEBUG` 是否在控制台输出所有的SQL

## 测试

```shell
make test
```

## 代码风格

使用 [golangci-lint](https://github.com/golangci/golangci-lint) 进行静态分析。

非 go 文件(yaml,json,markdown 等)使用 [prettier](https://prettier.io/) 进行格式化。
