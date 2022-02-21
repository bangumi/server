# Backend Server written in golang

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Bangumi/server/go?style=flat-square)
[![Codecov](https://img.shields.io/codecov/c/github/Bangumi/server/go?style=flat-square)](https://app.codecov.io/gh/Bangumi/server/go)

## Requirements

- go 1.17+
- GNU make

## Init

```bash
git clone --recursive https://github.com/bangumi/server.git
cd chii
make install
```

## Go Mod

你不应当导入 `github.com/bangumi/server/pkg` 以外的任何路径。

具体可用的包见 [pkg/readme.md](./pkg)

## License

chii 以 [GNU AGPLv3](./LICENSE.txt) 协议开源。
