开发工具:

- go 1.17+
- GNU make

## 测试

```shell
make test
```

## 代码风格

使用 [golangci-lint](https://github.com/golangci/golangci-lint) 进行静态分析。

非 go 文件(yaml,json,markdown 等)使用 [prettier](https://prettier.io/) 进行格式化。
