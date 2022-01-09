from typing import Any, Dict, Type, Union, Optional
from urllib.parse import quote

import orjson
from pydantic import Field, BaseModel
from starlette.requests import Request
from starlette.responses import JSONResponse
from starlette.datastructures import URL


class ErrorDetail(BaseModel):
    title: str
    description: str
    detail: Any = Field(..., description="can be anything")


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


def not_found(request: Request):
    return HTTPException(
        status_code=404,
        title="Not Found",
        description="resource can't be found in the database or has been removed",
        detail={
            "path_info": dict(request.path_params),
            "query_info": dict(request.query_params),
        },
    )
