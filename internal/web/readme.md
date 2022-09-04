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

使用 `http.Status*` 作为返回码，不要使用 `fiber.Status*`。

使用 `res.NewError(code int, message string) error` 来返回 http 响应。

如果是意料之外的错误，需要使用 `Handler.InternalError(c *fiber.Ctx, err error, message string, logFields ...zap.Field) error {`

此方法会打一个 log，并且返回 http 500 响应。

```golang
package handler

import (
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetCurrentUser(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)
	if !u.login {
		return res.Unauthorized("need login")
	}

	data, err := h.service.FetchUserData(c.UserContext(), u.ID)
	if err != nil {
		return h.InternalError(c, err, "failed to get user", log.UserID(u.ID), u.LogRequestID())
	}

	return res.JSON(c, data)
}
```

# frontend

前端 Demo <https://next.bgm.tv/demo/login>
