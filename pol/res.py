from typing import Any, Dict, Type, Union, Optional
from urllib.parse import quote

import orjson
from starlette.responses import JSONResponse
from starlette.datastructures import URL


class ORJSONResponse(JSONResponse):
    media_type = "application/json"

    def render(self, content: Any) -> bytes:
        return orjson.dumps(content)


class HTTPException(Exception):
    def __init__(
        self,
        status_code: int,
        title: str,
        description: str,
        detail: Any = None,
        headers: Optional[Dict[str, Any]] = None,
    ) -> None:
        self.status_code = status_code
        self.detail = detail
        self.headers = headers or {}
        self.title = title
        self.description = description


class HTTPRedirect(HTTPException):
    def __init__(
        self,
        url: Union[str, URL],
        status_code: int = 307,
        headers: dict = None,
    ) -> None:
        super().__init__(
            status_code=status_code, headers=headers, title="", description=""
        )
        self.headers["location"] = quote(str(url), safe=":/%#?=@[]!$&'()*+,;")


def response(model: Type = None, description: str = None) -> Dict[str, Any]:
    d: Dict[str, Any] = {}
    if model is not None:
        d["model"] = model
    if description:
        d["description"] = description
    return d


def public_cache(second: int):
    return {"cache-control": f"public, max-age={second}"}
