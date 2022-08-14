# Openapi

- `v0.yaml`： `api.bgm.tv` 域名的公开 API。
- `private.yaml`： `next.bgm.tv/p/` 路径下的前端私有 API。

你需要 [nodejs](https://nodejs.org/) 来测试 openapi 定义：

```bash
npm test
```

## 编辑

https://github.com/swagger-api/swagger-editor#docker

```
$ docker run -p 8061:8080 -v $(pwd)/openapi:/tmp -e SWAGGER_FILE=/tmp/v0.yaml swaggerapi/swagger-editor
# or
$ docker run -p 8061:8080 -v $(pwd)/openapi:/tmp -e SWAGGER_FILE=/tmp/private.yaml swaggerapi/swagger-editor
```
