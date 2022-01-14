# HTTP handler

所有的 `fastapi.APIRouter` 都应该以 `pol.router.ErrorCatchRoute` 作为 `route_class` 参数，用来捕获内部异常。

example:

```python
from fastapi import APIRouter

from pol.router import ErrorCatchRoute

router = APIRouter(tags=["章节"], route_class=ErrorCatchRoute)

@router.get("/path")
async def handler():
    ...
```

对于需要权限的请求，使用 `pol.api.v0.depends.auth.optional_user` 或者 `pol.api.v0.depends.auth.get_current_user` 获取当前访问的用户。

可以参照 [v0/me.py](./v0/me.py) 以及 [v0/subject.py](./v0//subject.py) 的 `获取条目`
