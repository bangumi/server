# Openapi

- `v0.yaml`： `api.bgm.tv` 域名的公开 API。
- `private.yaml`： `next.bgm.tv/p/` 路径下的前端私有 API。

你需要 [nodejs](https://nodejs.org/) 来测试 openapi 定义：

```bash
npm test
```

## 编辑

可以使用 `npm start` 来启动一个前端服务器来显示 Swagger UI.

这个 server 会对 openapi 定义进行一些额外的处理。包含 livereload server ，会在文件修改后自动刷新浏览器页面。
