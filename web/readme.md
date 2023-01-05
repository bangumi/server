# req

请求

# res

响应

# session

session manager

# rate

登录限流

# middleware

- `middleware/ua` 检查请求 User-Agent
- `middleware/origin` 检查请求 Origin
- `middleware/recovery` panic-recover

# handler

路由应该是 `internal/web/handler.Handler` 的一个方法。

使用 `res.NewError(code int, message string) error` 或者类似的 `res.BadRequest(msg string)` 来返回 http 响应。

# frontend

前端 Demo <https://next.bgm.tv/demo/login>
